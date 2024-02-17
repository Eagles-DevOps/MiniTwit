package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html"
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
		"url_for": func(routename string, username string) string { //with help from chatGPT
			switch routename {
			case "unfollow":
				return "/" + username + "/unfollow"
			case "follow":
				return "/" + username + "/follow"
			case "add_message":
				return "/add_message"
			case "timeline":
				return "/"
			case "public_timeline":
				return "/public"
			case "logout":
				return "/logout"
			case "login":
				return "/login"
			case "register":
				return "/register"
			default:
				return "/"
			}

		},
		"formatUsernameUrl": func(username string) string {
			return strings.Replace(username, " ", "%20", -1)
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
	r.HandleFunc("/logout", Logout)
	r.HandleFunc("/register", Register)

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
	//fmt.Println(session)
	user_id, ok := session.Values["user_id"]

	if !ok {
		fmt.Println("Session ended")
		return nil, err
	}

	user, err := query_db("SELECT * FROM user WHERE user_id = ?", []any{user_id}, true)
	if err != nil {
		fmt.Println("Unable to query for user data in before_request()")
		return nil, err
	}
	//fmt.Println("user: ", user)
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
		println("Error when trying to find the user in the database in follow", err.Error())
		http.Error(w, "Error when trying to find the user in the database in follow", http.StatusNotFound)
		return
	}

	who_id := user_id
	_, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", who_id, whom_id)
	if err != nil {
		http.Error(w, "Error when trying to insert data into database", http.StatusInternalServerError)
		return
	}
	message := fmt.Sprintf("You are now following &#34;%s&#34;", username)
	_message := html.UnescapeString(message)
	setFlash(r, w, _message)
	http.Redirect(w, r, "/"+username, http.StatusSeeOther)
}

func unfollow_user(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	println("displaying username for " + username)

	_, err := before_request(r)
	if err != nil {
		println("Error beforerequest in unfollow", err.Error())
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
		return
	}

	whom_id, err := get_user_id(username)
	if err != nil {
		println("Error when trying to find the user in the database in unfollow", err.Error())
		http.Error(w, "Error when trying to find the user in the database in unfollow", http.StatusNotFound)
		return
	}

	who_id := user_id
	_, err = db.Exec("DELETE FROM follower WHERE who_id=? and whom_id=?", who_id, whom_id)
	if err != nil {
		http.Error(w, "Error when trying to delete data from database", http.StatusInternalServerError)
		return
	}

	message := fmt.Sprintf("You are no longer following &#34;%s&#34;", username)
	_message := html.UnescapeString(message)
	setFlash(r, w, _message)
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
	setFlash(r, w, "Your message was recorded")

	http.Redirect(w, r, "/", http.StatusFound)
}

type Data struct { //used to inject the html template with the requestURI (to figure out if we're on public or user timeline)
	Message       any
	User          any
	Req           string
	Followed      any
	USERID        any
	FlashMessages any
	ErrMsg        string
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
		user_id := userMap["user_id"]

		var query = `SELECT message.*, user.* FROM message, user
        WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
            user.user_id = ? OR
            user.user_id IN (SELECT whom_id FROM follower
                                    where who_id = ?))
        ORDER BY message.pub_date desc limit ?`

		messages, err := query_db(query, []any{user_id, user_id, PER_PAGE}, false)
		if err != nil {
			fmt.Println("Timeline: Error when trying to query the database", err)
			http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
			return
		}

		println("user: ", messages)

		flash := getFlash(r, w)

		fmt.Println(user_id)
		d := Data{
			User:          user,
			Message:       messages,
			USERID:        user_id,
			FlashMessages: flash,
		}

		err = tpl.ExecuteTemplate(w, "newtimeline.html", d)
		if err != nil {
			fmt.Println("Error when trying to execute the template: ", err)
			http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
			return
		}

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

	flash := getFlash(r, w)
	d := Data{Message: messages, Req: r.RequestURI, FlashMessages: flash}
	err = tpl.ExecuteTemplate(w, "timeline.html", d)
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
		//http.Redirect(w, r, "/public", http.StatusFound)
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
	if checkNilInterface2(profile_user) {
		http.Redirect(w, r, "/public", http.StatusFound)
		return
	}
	profileuserMap := profile_user.(map[any]any)
	profile_user_id := profileuserMap["user_id"]
	fmt.Println(user_id)
	fmt.Println(profile_user_id)

	var followed bool = false
	usr, err := query_db(`select 1 from follower where
        follower.who_id = ? and follower.whom_id = ?`, []any{user_id, profile_user_id}, true)

	if err == nil && usr != nil {
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
	//fmt.Println(messages)

	flash := getFlash(r, w)

	fmt.Println(user_id)
	d := Data{Message: messages,
		Followed:      followed,
		User:          profile_user,
		Req:           r.RequestURI,
		USERID:        user_id,
		FlashMessages: flash,
	}

	fmt.Println("Rendering template...")
	//err = tpl.ExecuteTemplate(w, "timeline.html", dict)
	err = tpl.ExecuteTemplate(w, "timeline.html", d)
	if err != nil {
		fmt.Println("Error when trying to execute the template: ", err)
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
		return
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	usr, _ := before_request(r)

	if usr != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	if r.Method == "GET" {

		flash := getFlash(r, w)
		if flash != nil {

			fmt.Println("Login no ses: ", flash)

			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "login.html", d)
		} else {
			fmt.Println("No session in login")
			tpl.ExecuteTemplate(w, "login.html", nil)
		}

	} else if r.Method == "POST" {
		fmt.Println("POST, render login")
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := query_db("select * from user where username = ?", []any{username}, true)
		if err != nil || checkNilInterface2(user) {
			setFlash(r, w, "Invalid username")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "login.html", d)
			return
		}

		// Assuming user is a map with key 'pw_hash'
		userMap := user.(map[any]any)
		pwHash := userMap["pw_hash"].(string)

		err = checkPasswordHash(password, pwHash)
		if err != nil {
			setFlash(r, w, "Invalid password")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "login.html", d)
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
			panic("This is not allowed happen!")
		}
		//setting the session values
		session.Values["user_id"] = user_id
		session.Save(r, w)
		setFlash(r, w, "You were logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	usr, _ := before_request(r)

	if usr != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	if r.Method == "GET" {

		tpl.ExecuteTemplate(w, "register.html", nil)

	} else if r.Method == "POST" {

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		// Validate form input
		if username == "" {
			setFlash(r, w, "You have to enter a username")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "register.html", d)
			return

		} else if !strings.Contains(email, "@") {
			setFlash(r, w, "You have to enter a valid email address")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "register.html", d)
			return

		} else if password == "" {
			setFlash(r, w, "You have to enter a password")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "register.html", d)
			return

		} else if password != password2 {
			setFlash(r, w, "The two passwords do not match")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "register.html", d)
			return

		} else if id, _ := get_user_id(username); id != nil {
			setFlash(r, w, "The username is already taken")
			flash := getFlash(r, w)
			d := Data{
				FlashMessages: flash,
			}
			tpl.ExecuteTemplate(w, "register.html", d)
			return

		} else {
			// Hash the password
			hashedPassword, err := hashPassword(password)
			if err != nil {
				fmt.Println("Error hashing the password")
				return
			}

			// Insert the new user into the database
			_, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
			if err != nil {
				fmt.Println("Database error")
				return
			}
			setFlash(r, w, "You were successfully registered and can login now")
			//setFlash(r, w, "and can login now")

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
		setFlash(r, w, "You were logged out")
		delete(session.Values, "user_id")
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

func setFlash(r *http.Request, w http.ResponseWriter, message string) {
	session, _ := store.Get(r, "user-session")
	session.AddFlash(message)
	session.Save(r, w)
}

func getFlash(r *http.Request, w http.ResponseWriter) []interface{} {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return nil
	} else {
		flashes := session.Flashes()
		session.Save(r, w)
		return flashes
	}
}
