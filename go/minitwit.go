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
	tpl, err = template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
<<<<<<< HEAD

	r.HandleFunc("/login", Login)
	r.HandleFunc("/register", Register)
	r.HandleFunc("/logout", Logout)
	r.HandleFunc("/", timeline)
=======
    r.HandleFunc("/login", Login)
	r.HandleFunc("/register", Register)
	r.HandleFunc("/logout", Logout)
>>>>>>> 60760cfb1f7172d0ac66d93cc430ec8f2cfb071a
	r.HandleFunc("/public", public_timeline)
	r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/{username}/follow", follow_user)
	r.HandleFunc("/{username}/unfollow", unfollow_user)
<<<<<<< HEAD
=======
    r.HandleFunc("/{username}", user_timeline)
    r.HandleFunc("/", timeline)


>>>>>>> 60760cfb1f7172d0ac66d93cc430ec8f2cfb071a

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
func format_datetime(timestamp time.Time) string {
	return timestamp.Format("2006-01-02 @ 15:04")
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
func before_request(r *http.Request) {
	var err error
	db, err = connect_db()
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	if user_id != nil {
		user, err := query_db("SELECT * FROM user WHERE user_id = ?", []any{"user_id"}, true)
		if err != nil {
			return
		}
		fmt.Println("user: ", user)
	}
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
	session.AddFlash("You are now following %s", username)
	http.Redirect(w, r, "/"+username, http.StatusSeeOther)
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
	session.AddFlash("You are no longer following %s", username)
	http.Redirect(w, r, "/"+username, http.StatusSeeOther)
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
	session.AddFlash("Your message was recorded")
	http.Redirect(w, r, "/timeline", http.StatusSeeOther)
}

// """Shows a users timeline or if no user is logged in it will
// redirect to the public timeline.  This timeline shows the user's
// messages as well as all the messages of followed users."""
func timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: ", r.RemoteAddr)
	if user == nil {
		http.Redirect(w, r, "/public", http.StatusSeeOther)
	}
	messages, err := query_db(`SELECT message.*, user.* FROM message, user
    WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
        user.user_id = ? OR
        user.user_id IN (SELECT whom_id FROM follower
                                WHERE who_id = ?))
    ORDER BY message.pub_date DESC LIMIT ?`, []any{"user_id", "user_id", PER_PAGE}, false)

	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}

	err = tpl.ExecuteTemplate(w, "timeline.html", messages)
	if err != nil {
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
}

// """Displays the latest messages of all users."""
func public_timeline(w http.ResponseWriter, r *http.Request) {
	messages, err := query_db(`SELECT message.*, user.* FROM message, user
    WHERE message.flagged = 0 AND message.author_id = user.user_id
    ORDER BY message.pub_date desc limit ?`, []any{PER_PAGE}, false)
	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}

	err = tpl.ExecuteTemplate(w, "timeline.html", messages)
	if err != nil {
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
}

// """Display's a users tweets."""
func user_timeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	println("displaying username for " + username)
	profile_user, err := query_db("SELECT * FROM user WHERE username = ?", []any{username}, true)
	if err != nil {
		http.Error(w, "Error when trying to find the profile user in the database", http.StatusNotFound)
		return
	}
	profile_user_id, err := get_user_id(username)
	if err != nil {
		return
	}
	if user == nil {
		http.Error(w, "Error when trying to find the user in the database", http.StatusNotFound)
		return
	}

	_, err = query_db(`select 1 from follower where
        follower.who_id = ? and follower.whom_id = ?`, []any{user_id, profile_user_id}, true)

	var followed bool = false
	if err == nil {
		followed = true
	}
	messages, err := query_db(`SELECT message.*, user.* FROM message, user WHERE
	user.user_id = message.author_id AND user.user_id = ?
	ORDER BY message.pub_date desc limit ?`, []any{profile_user_id, PER_PAGE}, false)
	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}

	dict := make(map[string]any)
	dict["messages"] = messages
	dict["followed"] = followed
	dict["profile_user"] = profile_user

	err = tpl.ExecuteTemplate(w, "timeline.html", dict)
	if err != nil {
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "login_test.html", nil)

	} else if r.Method == "POST" {
		fmt.Println("POST, render login")
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := query_db("select * from user where username = ?", []any{username}, true)
		if err != nil {
			http.Error(w, "Invalid username", http.StatusInternalServerError)
			return
		}
		// Assuming user is a map with key 'pw_hash'
		userMap := user.(map[any]any)
		pwHash := userMap["pw_hash"].(string)

		err = checkPasswordHash(password, pwHash)
		if err != nil {
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
		user_id, err = get_user_id(username)
		if err != nil {
			fmt.Println("Can't find the user_id in database")
		}
		//setting the session values
		session.Values["user_id"] = user_id
		session.Save(r, w)

		// Redirect to timeline
		session.AddFlash("You were logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

		// Validate form input
		if username == "" {
			fmt.Println("You have to enter a username")

		} else if !strings.Contains(email, "@") {
			fmt.Println("You have to enter a valid email address")

		} else if password == "" {
			fmt.Println("You have to enter a password")

		} else if password != password2 {
			fmt.Println("The two passwords do not match")

		} else if _, err := get_user_id(username); err == nil {
			fmt.Println("The username is already taken")

		} else {
			// Hash the password
			hashedPassword, err := hashPassword(password)
			if err != nil {
				http.Error(w, "Error hashing the password", http.StatusInternalServerError)
				fmt.Println("Error hashing the password")
				return
			}

			// Insert the new user into the database
			_, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
			if err != nil {
				http.Error(w, "Error when trying to insert data into the database", http.StatusInternalServerError)
				fmt.Println("Error when trying to insert data into the database")
				return
			}
			fmt.Println("You were successfully registered and can login now")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	//check if there is any session
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Println("Error getting session data")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		// Logout session
		session.AddFlash("You were logged out")
		session.Values["user_id"] = nil
		session.Options.MaxAge = -1
		err = session.Save(r, w)
		if err != nil {
			fmt.Println("Error in saving the session data")
		}
		http.Redirect(w, r, "/public", http.StatusSeeOther)
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

// # add some filters to jinja and set the secret key and debug mode
// # from the configuration.
// app.jinja_env.filters['datetimeformat'] = format_datetime
// app.jinja_env.filters['gravatar'] = gravatar_url
// app.secret_key = SECRET_KEY
// app.debug = DEBUG

// if __name__ == '__main__':
//     app.run(host="0.0.0.0")
