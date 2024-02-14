package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	DATABASE   = "./minitwit.db"
	PER_PAGE   = 30
	DEBUG      = true
	SECRET_KEY = "development key"
)

var db *sql.DB
var f []byte

var store = sessions.NewCookieStore([]byte("SESSIONKEY"))
var session *sessions.Session
var user any
var user_id any
var tpl *template.Template

func main() {
	var err error

	funcMap := template.FuncMap{"getavatar": func(url string, size int) string {
		return gravatar_url(url, size)
	},
		"gettimestamp": func(time int64) string {
			return format_datetime(time)
		},
	}
	tpl, err = template.New("timeline.html").Funcs(funcMap).ParseGlob("templates/*.html") // we need to add the funcs that we want to use before parsing

	//tpl, err = template.ParseGlob("templates/*.html")

	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/timeline", timeline)
	r.HandleFunc("/public_timeline", public_timeline)
	r.HandleFunc("/{username}", user_timeline)

	r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/{username}/follow", follow_user)
	r.HandleFunc("/{username}/unfollow", unfollow_user)
	r.HandleFunc("/login", Login)
	r.HandleFunc("/register", Register)
	r.HandleFunc("/logout", Logout)

	db, err = connect_db()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	//content, err := query_db("SELECT user_id FROM user WHERE username IN (?, ?, ?)", []any{"Roger Histand", "Ayako Yestramski", "Leonora Alford"}, false)
	//dt := format_datetime(time.Now())
	//id_string := strconv.FormatInt(int64(id), 10)
	//output := gravatar_url("anam@itu.dk", 80)

	//fmt.Println("Content: ", content, err)
	fmt.Println("Listening on port 15000...")
	err = http.ListenAndServe(":15000", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// "/"
func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// """Returns a new connection to the database."""
func connect_db() (*sql.DB, error) {
	fmt.Println("Connecting to database...")
	return sql.Open("sqlite3", DATABASE)
}

// """Creates the database tables."""
func init_db() ([]byte, error) {
	fmt.Println("Initializing database...")
	return os.ReadFile("schema.sql")
}

// """Queries the database and returns a list of dictionaries."""
func query_db(query string, args []any, one bool) (any, error) {
	cur, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer cur.Close()

	var rv []map[any]any
	cols, err := cur.Columns()
	if err != nil {
		return nil, fmt.Errorf("error retrieving columns: %w", err)
	}
	for cur.Next() {
		row := make([]any, len(cols))
		for i := range row {
			row[i] = new(any)
		}
		err = cur.Scan(row...)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		dict := make(map[any]any)
		for i, col := range cols {
			dict[col] = *(row[i].(*any))
		}
		rv = append(rv, dict)
		if one {
			break
		}
	}

	if err = cur.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	if len(rv) != 0 {
		if one {
			return rv[0], nil
		}
		return rv, nil
	}
	return nil, nil
}

// """Format a timestamp for display."""
func format_datetime(timestamp int64) string {
	return strconv.FormatInt(timestamp, 10)
	//return timestamp.Format("2006-01-02 @ 15:04")
}

// """Return the gravatar image for the given email address."""
func gravatar_url(email string, size int) string {
	hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("http://www.gravatar.com/avatar/%s?d=identicon&s=%d", hex.EncodeToString(hash[:]), size)
}

// """Convenience method to look up the id for a username."""
func get_user_id(username string) (any, error) {
	var user_id int
	rv := db.QueryRow("SELECT user_id FROM user WHERE username = ?",
		username)
	err := rv.Scan(&user_id)
	if err != nil {
		return nil, err
	}
	return user_id, err
}

// """Make sure we are connected to the database each request and look
// up the current user so that we know he's there.

func before_request(r *http.Request) (any, error) {
	var err error
	db, err = connect_db()
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	session, err := store.Get(r, "user-session")
	user_id, ok := session.Values["user_id"].(any)

	if !ok {
		fmt.Println("Session ended")
		return nil, err
	}

	user, err := query_db("SELECT * FROM user WHERE user_id = ?", []any{user_id}, true)
	if err != nil {
		return nil, err
	}
	fmt.Println("user: ", user)
	return user, err
}

// """Closes the database again at the end of the request."""
func after_request(response http.Response) http.Response {
	db.Close()
	return response
}

func follow_user(w http.ResponseWriter, r *http.Request) {
	//"""Adds the current user as follower of the given user."""
	vars := mux.Vars(r)
	username := vars["username"]
	println("Now following " + username)

	if user == nil {
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
	}
	whom_id, err := get_user_id(username)
	if err != nil {
		http.Error(w, "Error when trying to find the user in the database", http.StatusNotFound)
		return
	}
	who_id := user_id
	_, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", who_id, whom_id)
	if err != nil {
		http.Error(w, "Error when trying to insert data into database", http.StatusInternalServerError)
	}
	fmt.Printf("You are now following %s", username)
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func unfollow_user(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	println("displaying username for " + username)
	if user == nil {
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
	}
	whom_id, err := get_user_id(username)
	if err != nil {
		http.Error(w, "Error when trying to find the user in the database", http.StatusNotFound)
		return
	}
	who_id := user_id
	_, err = db.Exec("DELETE FROM follower WHERE who_id=? and whom_id=?", who_id, whom_id)
	if err != nil {
		http.Error(w, "Error when trying to delete data from database", http.StatusInternalServerError)
	}
	fmt.Printf("You are no longer following %s", username)
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

// """Registers a new message for the user."""
func add_message(w http.ResponseWriter, r *http.Request) {
	if user == nil {
		http.Error(w, "You need to login before you can post a message", http.StatusUnauthorized)
	}
	text := r.FormValue("text")
	if text != "" {
		db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)", user_id, text, int(time.Now().Unix()))
	} else {
		fmt.Printf("You need to write a message in the text form")
	}
	fmt.Printf("Your message was recorded")
	http.Redirect(w, r, "/timeline", http.StatusFound)
}

// TODO: include the followed and profile_user functionalities
func render_template(w http.ResponseWriter, r *http.Request, tmplt string, query string, args []any, one bool, followed any, profile_user any) {
	messages, err := query_db(query, args, false)
	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}

	//_template, err := template.ParseFiles(tmplt) // this breaks it, but we've also already parsed the file
	//if err != nil {//
	//	http.Error(w, "Error when trying to parse the template", http.StatusInternalServerError)
	//}
	err = tpl.ExecuteTemplate(w, tmplt, messages)
	if err != nil {
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
}

// """Shows a users timeline or if no user is logged in it will
// redirect to the public timeline.  This timeline shows the user's
// messages as well as all the messages of followed users."""
func timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: ", r.RemoteAddr)

	var err error
	user, err = before_request(r)

	if err != nil {
		http.Redirect(w, r, "/public", http.StatusFound)
	} else {
		userMap := user.(map[any]any)
		username := userMap["username"].(string)
		usernameURL := fmt.Sprintf("/%s", username)
		http.Redirect(w, r, usernameURL, http.StatusFound)
	}

	//render_template(w, r, "timeline.html", `SELECT message.* FROM message
	//WHERE message.message_id = 1`, []any{}, false, nil, nil)

	/*
	  `WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
	      user.user_id = ? OR
	      user.user_id IN (SELECT whom_id FROM follower
	                              WHERE who_id = ?))
	  ORDER BY message.pub_date DESC LIMIT ?`, []any{"user_id", "user_id", PER_PAGE}, false, nil, nil)
	*/

}

// """Displays the latest messages of all users."""
func public_timeline(w http.ResponseWriter, r *http.Request) {
	/*
		var data, _ = query_db(`SELECT message.*, user.* FROM message, user
		WHERE message.flagged = 0 AND message.author_id = user.user_id
		ORDER BY message.pub_date desc limit ?`, []any{PER_PAGE}, false)
	*/
	var query = `SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND message.author_id = user.user_id
	ORDER BY message.pub_date desc limit ?`
	render_template(w, r, "timeline.html", query, []any{PER_PAGE}, false, nil, nil)
	/*
		if err := tpl.ExecuteTemplate(w, "timeline.html", data); err != nil {
			panic(err)
		}
	*/
}

// """Display's a users tweets."""
func user_timeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	println("displaying username for " + username)

	fmt.Println("Calling before_request() in user_timeline...")

	_, err := before_request(r)

	//Uncertain how to handle the case where user is not logged in. Currently redirecting to /public
	if err != nil {
		//replace with flash popup
		//http.Error(w, "Error when trying to find the user in the database", http.StatusNotFound)
		http.Redirect(w, r, "/public", http.StatusSeeOther)
	}

	fmt.Println("Query for profile_user data...")

	profile_user, err := query_db("SELECT * FROM user WHERE username = ?", []any{username}, true)
	if err != nil {
		//replace with flash popup
		//http.Error(w, "Error when trying to find the profile user in the database", http.StatusNotFound)
		http.Redirect(w, r, "/public", http.StatusSeeOther)
	}

	profileuserMap := profile_user.(map[any]any)
	profile_user_id := profileuserMap["user_id"]

	var followed bool = false
	_, err = query_db(`select 1 from follower where
        follower.who_id = ? and follower.whom_id = ?`, []any{user_id, profile_user_id}, true)

	if err == nil {
		followed = true
	}

	fmt.Println("Query for user_timeline...")

	var query = `SELECT message.*, user.* FROM message, user WHERE
	user.user_id = message.author_id AND user.user_id = ?
	ORDER BY message.pub_date desc limit ?`

	messages, err := query_db(query, []any{profile_user_id, PER_PAGE}, false)
	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}

	dict := make(map[string]any)
	dict["messages"] = messages
	dict["followed"] = followed
	dict["profile_user"] = profile_user

	fmt.Println("Rendering template...")

	err = tpl.ExecuteTemplate(w, "timeline.html", dict)
	if err != nil {
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "login_test.html", nil)

	} else if r.Method == "POST" {
		var user_id_val any
		fmt.Println("POST, render login")
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := query_db("select * from user where username = ?", []any{username}, true)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			fmt.Println("Database error")
			return
		}

		if user == nil {
			http.Error(w, "Invalid username", http.StatusBadRequest)
			return
		}

		// Assuming user is a map with key 'pw_hash'
		userMap := user.(map[any]any)
		pwHash := userMap["pw_hash"].(string)

		if !checkPasswordHash(password, pwHash) {
			http.Error(w, "Invalid password", http.StatusBadRequest)
			return
		}

		// Set session data
		session, _ := store.Get(r, "user-session")
		session.Options = &sessions.Options{
			Path:   "/",
			MaxAge: 3600, // 1 hour in seconds
			//MaxAge: 5,
			HttpOnly: true, // Recommended for security
		}

		user_id_val, err = get_user_id(username)
		if err != nil {
			fmt.Println("Cant get User ID")
		}
		//values needs to be from a name form
		session.Values["user_id"] = user_id_val
		session.Save(r, w)

		// Redirect to timeline
		fmt.Println("Logged in redirecting to timeline")
		http.Redirect(w, r, "/timeline", http.StatusSeeOther)
		//tpl.ExecuteTemplate(w, "login_test.html", nil)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		tpl.ExecuteTemplate(w, "register.html", nil)

	} else if r.Method == "POST" {

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		var error_s string

		// Validate form input
		if username == "" {
			error_s = "You have to enter a username"
			fmt.Println(error_s)
		} else if !strings.Contains(email, "@") {
			error_s = "You have to enter a valid email address"
			fmt.Println(error_s)

		} else if password == "" {
			error_s = "You have to enter a password"
			fmt.Println(error_s)

		} else if password != password2 {
			error_s = "The two passwords do not match"
			fmt.Println(error_s)

		} else if _, err := get_user_id(username); err == nil {
			error_s = "The username is already taken"
			fmt.Println(error_s)
		} else {
			// Hash the password
			hashedPassword, err := hashPassword(password)
			if err != nil {
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				fmt.Println("Error hashing the password")
				return
			}

			// Insert the new user into the database
			_, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				fmt.Println("Database error")
				return
			}

			fmt.Println("User added")
			tpl.ExecuteTemplate(w, "login_test.html", nil)
		}
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	//check if there is any session
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error getting session data")
		tpl.ExecuteTemplate(w, "login_test.html", nil)
	} else {
		// Logout session
		user_id_val, ok := session.Values["user_id"].(any)
		if !ok {
			fmt.Println("Session ended")
		} else {
			fmt.Println("Logging of:", user_id_val)
			session.Options.MaxAge = -1
			err = session.Save(r, w)
			if err != nil {
				fmt.Println("Error saving the session")
				return
			}
			fmt.Println("Logged off")
		}
	}

	//return to /public
	tpl.ExecuteTemplate(w, "login_test.html", nil)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// # add some filters to jinja and set the secret key and debug mode
// # from the configuration.
// app.jinja_env.filters['datetimeformat'] = format_datetime
// app.jinja_env.filters['gravatar'] = gravatar_url
// app.secret_key = SECRET_KEY
// app.debug = DEBUG

// if __name__ == '__main__':
//     app.run(host="0.0.0.0")
