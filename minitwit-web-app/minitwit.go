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
	DATABASE = "./minitwit.db"
	PER_PAGE = 30
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("SESSIONKEY"))
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

	db, err = before_request()
	if err != nil {
		fmt.Println("Error connecting to the database")
	}
	defer after_request()

	fmt.Println("Listening on port 15000...")
	err = http.ListenAndServe(":15000", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// """Returns a new connection to the database."""
func connect_db() (db *sql.DB, err error) {
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
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 @ 15:04")
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

// """Gets the session"""
func getSession(r *http.Request) (*sessions.Session, error) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return nil, err
	}
	return session, err
}

// """Gets the user in the session"""
func getUser(r *http.Request) (any, any, error) {
	session, _ := getSession(r)
	user_id, ok := session.Values["user_id"]

	if !ok {
		fmt.Println("No user in the session")
		return nil, nil, fmt.Errorf("no user in the session")
	}
	user, err := query_db("SELECT * FROM user WHERE user_id = ?", []any{user_id}, true)
	if err != nil {
		fmt.Println("Unable to query for user data in getUser()")
		return nil, nil, err
	}
	return user, user_id, err
}

// """Opens the database before the request."""
func before_request() (*sql.DB, error) {
	return connect_db()
}

// """Closes the database again at the end of the request."""
func after_request() {
	db.Close()
}

// """Adds the current user as follower of the given user."""
func follow_user(w http.ResponseWriter, r *http.Request) {
	user, user_id, err := getUser(r)
	if err != nil || isNil(user) {
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	username := vars["username"]
	println("Now following " + username)

	whom_id, err := get_user_id(username)
	if err != nil {
		http.Error(w, "Followuser: Error when trying to find the user in the database in follow", http.StatusNotFound)
		return
	}
	_, err = db.Exec("INSERT INTO follower (who_id, whom_id) VALUES (?, ?)", user_id, whom_id)
	if err != nil {
		fmt.Println("Error when trying to insert data into the database")
		return
	}
	message := fmt.Sprintf("You are now following &#34;%s&#34;", username)
	setFlash(w, r, message)
	http.Redirect(w, r, "/"+username, http.StatusSeeOther)
}

// """Removes the current user as follower of the given user."""
func unfollow_user(w http.ResponseWriter, r *http.Request) {
	user, user_id, err := getUser(r)
	if err != nil || isNil(user) {
		http.Error(w, "You need to login before you can follow the user", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	username := vars["username"]
	println("displaying username for " + username)

	whom_id, err := get_user_id(username)
	if err != nil {
		http.Error(w, "Error when trying to find the user in the database in unfollow", http.StatusNotFound)
		return
	}
	_, err = db.Exec("DELETE FROM follower WHERE who_id=? and whom_id=?", user_id, whom_id)
	if err != nil {
		fmt.Println("Error when trying to delete data from database")
		return
	}
	message := fmt.Sprintf("You are no longer following &#34;%s&#34;", username)
	setFlash(w, r, message)
	http.Redirect(w, r, "/"+username, http.StatusFound)
}

// """Registers a new message for the user."""
func add_message(w http.ResponseWriter, r *http.Request) {
	user, user_id, err := getUser(r)
	if err != nil || isNil(user) {
		http.Error(w, "You need to login before you can post a message", http.StatusUnauthorized)
		return
	}
	text := r.FormValue("text")
	if text != "" {
		db.Exec("INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)", user_id, text, int(time.Now().Unix()))
		setFlash(w, r, "Your message was recorded")
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// """This data is parsed to the templates"""
type Data struct {
	Message       any
	User          any
	Profileuser   any
	Req           string
	Followed      any
	FlashMessages any
}

// """Shows a users timeline or if no user is logged in it will
// redirect to the public timeline.  This timeline shows the user's
// messages as well as all the messages of followed users."""
func timeline(w http.ResponseWriter, r *http.Request) {
	_, ip, _ := net.SplitHostPort(r.RemoteAddr)
	fmt.Println("We got a visitor from: ", ip)

	user, user_id, err := getUser(r)
	if err != nil || isNil(user) {
		http.Redirect(w, r, "/public", http.StatusFound)
	} else {
		var query = `SELECT message.*, user.* FROM message, user
        WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
            user.user_id = ? OR
            user.user_id IN (SELECT whom_id FROM follower
                                    where who_id = ?))
        ORDER BY message.pub_date desc limit ?`

		messages, err := query_db(query, []any{user_id, user_id, PER_PAGE}, false)
		if err != nil {
			fmt.Println("Timeline: Error when trying to query the database", err)
			return
		}
		flash := getFlash(w, r)
		profile_user := user

		d := Data{
			User:          user,
			Profileuser:   profile_user,
			Message:       messages,
			FlashMessages: flash,
		}

		err = tpl.ExecuteTemplate(w, "timeline.html", d)
		if err != nil {
			fmt.Println("Error when trying to execute the template: ", err)
			return
		}
	}
}

// """Displays the latest messages of all users."""
func public_timeline(w http.ResponseWriter, r *http.Request) {
	user, _, err := getUser(r)
	if err != nil || isNil(user) {
		println("public timeline: the user is not logged in")
	}
	var query = `SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND message.author_id = user.user_id
	ORDER BY message.pub_date desc limit ?`

	messages, err := query_db(query, []any{PER_PAGE}, false)
	if err != nil {
		println("Error when trying to query the database: ", err)
		return
	}
	flash := getFlash(w, r)

	d := Data{Message: messages,
		User:          user,
		Req:           r.RequestURI,
		FlashMessages: flash,
	}
	err = tpl.ExecuteTemplate(w, "timeline.html", d)
	if err != nil {
		println("Error trying to execute template: ", err)
		return
	}
}

// """Display's a users tweets."""
func user_timeline(w http.ResponseWriter, r *http.Request) {
	user, user_id, err := getUser(r)
	if err != nil || isNil(user) {
		setFlash(w, r, "You need to login before you can see the user's timeline")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		//fmt.Println("Error when trying to find the user in the database: ", err)
		//http.Error(w, "Error when trying to find the user in the database", http.StatusNotFound)
		return
	}
	vars := mux.Vars(r)
	username := vars["username"]

	profile_user, err := query_db("SELECT * FROM user WHERE username = ?", []any{username}, true)
	if err != nil || isNil(profile_user) {
		fmt.Println("Error when trying to find the profile user in the database: ", err)
		return
	}
	profileuserMap := profile_user.(map[any]any)
	profile_user_id := profileuserMap["user_id"]

	followed := false
	usr, err := query_db(`select 1 from follower where
        follower.who_id = ? and follower.whom_id = ?`, []any{user_id, profile_user_id}, true)

	if err == nil && usr != nil {
		followed = true
	}
	var query = `SELECT message.*, user.* FROM message, user WHERE
	user.user_id = message.author_id AND user.user_id = ?
	ORDER BY message.pub_date desc limit ?`

	messages, err := query_db(query, []any{profile_user_id, PER_PAGE}, false)
	if err != nil {
		fmt.Println("User Timeline: Error when trying to query the database", err)
		return
	}
	flash := getFlash(w, r)

	fmt.Println(user_id)
	d := Data{Message: messages,
		Followed:      followed,
		User:          user,
		Profileuser:   profile_user,
		FlashMessages: flash,
	}
	err = tpl.ExecuteTemplate(w, "timeline.html", d)
	if err != nil {
		fmt.Println("Error when trying to execute the template: ", err)
		return
	}
}

// """Logs the user in."""
func Login(w http.ResponseWriter, r *http.Request) {
	user, _, err := getUser(r)
	if err == nil && !(isNil(user)) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else if r.Method == "GET" {
		reload(w, r, "", "login.html")

	} else if r.Method == "POST" {
		fmt.Println("POST, render login")
		username := r.FormValue("username")
		password := r.FormValue("password")

		user, err := query_db("select * from user where username = ?", []any{username}, true)
		if err != nil || isNil(user) {
			reload(w, r, "Invalid username", "login.html")
			return
		}
		userMap := user.(map[any]any)
		pwHash := userMap["pw_hash"].(string)

		err = checkPasswordHash(password, pwHash)
		if err != nil {
			reload(w, r, "Invalid password", "login.html")
			return
		}
		session, _ := getSession(r)
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   3600, // 1 hour in seconds
			HttpOnly: true, // Recommended for security
		}
		user_id, err := get_user_id(username)
		if err != nil {
			panic("This is not allowed happen!")
		}
		session.Values["user_id"] = user_id
		session.Save(r, w)
		setFlash(w, r, "You were logged in")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

// """Registers the user."""
func Register(w http.ResponseWriter, r *http.Request) {
	user, _, err := getUser(r)
	if err == nil && !(isNil(user)) {
		http.Redirect(w, r, "/", http.StatusSeeOther)

	} else if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "register.html", nil)

	} else if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if username == "" {
			reload(w, r, "You have to enter a username", "register.html")
			return

		} else if !strings.Contains(email, "@") {
			reload(w, r, "You have to enter a valid email address", "register.html")
			return

		} else if password == "" {
			reload(w, r, "You have to enter a password", "register.html")
			return

		} else if password != password2 {
			reload(w, r, "The two passwords do not match", "register.html")
			return

		} else if id, _ := get_user_id(username); id != nil {
			reload(w, r, "The username is already taken", "register.html")
			return

		} else {
			hashedPassword, err := hashPassword(password)
			if err != nil {
				fmt.Println("Error hashing the password")
				return
			}
			_, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
			if err != nil {
				fmt.Println("Database error")
				return
			}
			setFlash(w, r, "You were successfully registered and can login now")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	}
}

// """Logs the user out"""
func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		fmt.Println("Error getting session data")
	} else {
		setFlash(w, r, "You were logged out")
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
func isNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}

func setFlash(w http.ResponseWriter, r *http.Request, message string) {
	session, _ := getSession(r)
	session.AddFlash(html.UnescapeString(message))
	session.Save(r, w)
}

func getFlash(w http.ResponseWriter, r *http.Request) []interface{} {
	session, err := getSession(r)
	if err != nil {
		return nil
	} else {
		flashes := session.Flashes()
		session.Save(r, w)
		return flashes
	}
}

func reload(w http.ResponseWriter, r *http.Request, message string, template string) {
	d := Data{}
	if message != "" {
		setFlash(w, r, message)
	}
	d.FlashMessages = getFlash(w, r)
	tpl.ExecuteTemplate(w, template, d)
}
