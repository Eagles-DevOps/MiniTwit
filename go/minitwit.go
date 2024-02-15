package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net"
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

	funcMap := template.FuncMap{"getavatar": func(url string, size int) string {
		return gravatar_url(url, size)
	},
		"gettimestamp": func(time int64) string {
			return format_datetime(time)
		},
	}
	tpl, err = template.New("timeline.html").Funcs(funcMap).ParseGlob("templates/*.html") // We need to add the funcs that we want to use before parsing

	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/", timeline)
	r.HandleFunc("/public", public_timeline)
	r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/login", Login)
	r.HandleFunc("/register", Register)
	r.HandleFunc("/logout", Logout)

	r.HandleFunc("/{username}/follow", follow_user)
	r.HandleFunc("/{username}/unfollow", unfollow_user)
	r.HandleFunc("/{username}", user_timeline)

	db, err = connect_db()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer db.Close()

	//content, err := query_db("SELECT user_id FROM user WHERE username IN (?, ?, ?)", []any{"Roger Histand", "Ayako Yestramski", "Leonora Alford"}, false)

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
	//Todo should not actually be an error to not find any rows, fix when solution to identifying the empty interface returned exists
	return nil, nil
}

// """Format a timestamp for display."""
func format_datetime(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 @ 15:04")
	//return strconv.FormatInt(timestamp, 10)
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
		return nil, err
	}

	session, err := store.Get(r, "user-session")
	user_id, ok := session.Values["user_id"].(any)

	if !ok {
		fmt.Println("Session ended")
		return nil, err
	}

	user, err := query_db("SELECT * FROM user WHERE user_id = ?", []any{user_id}, true)
	if err != nil {
		fmt.Println("Unable to query for user data in before_request()")
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

	_, err := before_request(r)

	if err != nil {
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
		return
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
		return
	}

	session.AddFlash("You are now following %s", username)
	http.Redirect(w, r, "/"+username, http.StatusSeeOther)
}

func unfollow_user(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	println("displaying username for " + username)

	_, err := before_request(r)
	if err != nil {
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
		return
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
		return
	}
	session.AddFlash("You are no longer following %s", username)
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

// """Registers a new message for the user."""
func add_message(w http.ResponseWriter, r *http.Request) {
	_, err := before_request(r)

	if err != nil {
		http.Error(w, "You need to login before you can post a message", http.StatusUnauthorized)
		return
	}
	text := r.FormValue("text")
	if text != "" {
		db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)", user_id, text, int(time.Now().Unix()))
	} else {
		fmt.Printf("You need to write a message in the text form")
		http.Error(w, "You need to write a message in the text form", http.StatusBadRequest)
	}
	session, _ := store.Get(r, "user-session")
	session.AddFlash("Your message was recorded")
	err = session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// """Shows a users timeline or if no user is logged in it will
// redirect to the public timeline.  This timeline shows the user's
// messages as well as all the messages of followed users."""
func timeline(w http.ResponseWriter, r *http.Request) {
	_, ip, _ := net.SplitHostPort(r.RemoteAddr)
	fmt.Println("We got a visitor from: ", ip)

	var err error
	user, err = before_request(r)

	if err != nil || checkNilInterface2(user) {
		http.Redirect(w, r, "/public", http.StatusFound)
	} else {
		userMap := user.(map[any]any)
		username := userMap["username"].(string)
		usernameURL := fmt.Sprintf("/%s", username)
		http.Redirect(w, r, usernameURL, http.StatusFound)
	}
}

// """Displays the latest messages of all users."""
func public_timeline(w http.ResponseWriter, r *http.Request) {

	var query = `SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND message.author_id = user.user_id
	ORDER BY message.pub_date desc limit ?`

	messages, err := query_db(query, []any{PER_PAGE}, false)
	if err != nil {
		println("Error when trying to query the database: ", err)
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}

	err = tpl.ExecuteTemplate(w, "timeline.html", messages)
	if err != nil {
		println("Error trying to execute template: ", err)
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
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
		fmt.Println("Error when trying to find the user in the database: ", err)
		http.Error(w, "Error when trying to find the user in the database", http.StatusNotFound)
		return
	}

	fmt.Println("Query for profile_user data...")

	profile_user, err := query_db("SELECT * FROM user WHERE username = ?", []any{username}, true)
	if err != nil {
		//replace with flash popup
		http.Error(w, "Error when trying to find the profile user in the database", http.StatusNotFound)
		return
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
		fmt.Println("User Timeline: Error when trying to query the database", err)
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
		return
	}

	dict := make(map[string]any)
	dict["messages"] = messages
	dict["followed"] = followed
	dict["profile_user"] = profile_user

	fmt.Println("Rendering template...")

	err = tpl.ExecuteTemplate(w, "timeline.html", dict)
	if err != nil {
		fmt.Println("Error when trying to execute the template: ", err)
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "login.html", nil)

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

// ChatGPT
func checkNilInterface2(i interface{}) bool {
	if i == nil || (i != nil && i == interface{}(nil)) {
		return true
	} else {
		return false
	}
}
