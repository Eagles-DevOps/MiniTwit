package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"
    "log"
    "golang.org/x/crypto/bcrypt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
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
	//r.HandleFunc("/", handle)
	//r.HandleFunc("/", timeLine)
	//r.HandleFunc("/public", public_timeline)
	//r.HandleFunc("/<username>", user_timeline)
	//r.HandleFunc("/<username>/follow", follow_user)
	//r.HandleFunc("/<username>/unfollow", unfollow_user)
	//r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/login", Login)
	//r.HandleFunc("/register", register)
	//r.HandleFunc("/logout", logout)
	//http.Handle("/", r)

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
	fmt.Println("Listening on port 8080...")
	err = http.ListenAndServe("localhost:8080", r)
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
	//session, _ := store.Get(r, "session-name")
	//user_id := session.Values["user_id"]
	//fmt.Println("user_id: ", user_id)
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

func follow_user(username string, w http.ResponseWriter, r *http.Request) {
	//"""Adds the current user as follower of the given user."""
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
	//http.Redirect(w, r, "/<username>", http.StatusFound)
}

func unfollow_user(username string, w http.ResponseWriter, r *http.Request) {
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
	//http.Redirect(w, r, "/<username>", http.StatusFound)
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
	//http.Redirect(w, r, "/", http.StatusFound)
}

// TODO: include the followed and profile_user functionalities
func render_template(w http.ResponseWriter, r *http.Request, tmplt string, query string, args []any, one bool, followed any, profile_user any) {
	messages, err := query_db(query, args, false)
	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusInternalServerError)
	}
	_template, err := template.ParseFiles(tmplt)
	if err != nil {
		http.Error(w, "Error when trying to parse the template", http.StatusInternalServerError)
	}
	err = _template.Execute(w, messages)
	if err != nil {
		http.Error(w, "Error when trying to execute the template", http.StatusInternalServerError)
	}
}

// """Shows a users timeline or if no user is logged in it will
// redirect to the public timeline.  This timeline shows the user's
// messages as well as all the messages of followed users."""
func timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: ", r.RemoteAddr)
	if user == nil {
		//http.Redirect(w, r, "/", http.StatusFound)
	}
	render_template(w, r, "timeline.html", `SELECT message.*, user.* FROM message, user
    WHERE message.flagged = 0 AND message.author_id = user.user_id AND (
        user.user_id = ? OR
        user.user_id IN (SELECT whom_id FROM follower
                                WHERE who_id = ?))
    ORDER BY message.pub_date DESC LIMIT ?`, []any{"user_id", "user_id", PER_PAGE}, false, nil, nil)
}

// """Displays the latest messages of all users."""
func public_timeline(w http.ResponseWriter, r *http.Request) {
	render_template(w, r, "timeline.html", `SELECT message.*, user.* FROM message, user
    WHERE message.flagged = 0 AND message.author_id = user.user_id
    ORDER BY message.pub_date desc limit ?`, []any{PER_PAGE}, false, nil, nil)
}

// """Display's a users tweets."""
func user_timeline(w http.ResponseWriter, r *http.Request, username string) {
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
	followed, err := query_db(`select 1 from follower where
        follower.who_id = ? and follower.whom_id = ?`, []any{user_id, profile_user_id}, true)

	if err != nil {
		http.Error(w, "Error when trying to query the database", http.StatusNotFound)
		return
	}
	render_template(w, r, "timeline.html", `SELECT message.*, user.* FROM message, user WHERE
        user.user_id = message.author_id AND user.user_id = ?
        ORDER BY message.pub_date desc limit ?`, []any{profile_user_id, PER_PAGE}, false, followed, profile_user)
}

func Login(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        fmt.Println("GET, render login")
        var name = "- Roman"

        data := struct {
            Name string
        }{
            Name: name,
        }
    
        tpl.ExecuteTemplate(w, "login_test.html", data)

    }else if r.Method == "POST"{
        fmt.Println("POST, render login")
        username := r.FormValue("username")
        password := r.FormValue("password")

        user, err := query_db("select * from user where username = ?", []any{username}, true)
        if err != nil {
            //http.Error(w, "Database error", http.StatusInternalServerError)
            fmt.Println("Database error")
            return
        }

        if user == nil {
            //http.Error(w, "Invalid username", http.StatusBadRequest)
            return
        }

        // Assuming user is a map with key 'pw_hash'
        userMap := user.(map[any]any)
        pwHash := userMap["pw_hash"].(string)
        
        // Check password (assuming you have a function checkPasswordHash)
        if !checkPasswordHash(password, pwHash) {
            //http.Error(w, "Invalid password", http.StatusBadRequest)
            return
        }


        // Set session data
        session, _ := store.Get(r, "user-session")
        session.Options = &sessions.Options{
            Path:     "/",
            //MaxAge:   3600, // 1 hour in seconds
            MaxAge: 20,
            HttpOnly: true,  // Recommended for security
        }
        
        //values needs to be from a name form
        session.Values["name"] = "Tester"
        session.Save(r,w)

        //session.Values["user_id"] = userMap["user_id"]
        //session.Save(r, w)

        // Redirect to timeline
        //http.Redirect(w, r, "/timeline", http.StatusSeeOther)
        fmt.Println("Logged in")
        tpl.ExecuteTemplate(w, "login_test.html", nil)
    }
}

func RegisterTMP(){
    username := "Tester"
    email := "tester@test.com"
    password := "test"
    password2 := "test"

    var error_s string

    // Validate form input
    if username == "" {
        error_s = "You have to enter a username"
    } else if !strings.Contains(email, "@") {
        error_s = "You have to enter a valid email address"
    } else if password == "" {
        error_s = "You have to enter a password"
    } else if password != password2 {
        error_s = "The two passwords do not match"
    } else if _, err := get_user_id(username); err == nil {
        error_s = "The username is already taken"
    } else {
        // Hash the password
        hashedPassword, err := hashPassword(password)
        if err != nil {
            //http.Error(w, "Error hashing password", http.StatusInternalServerError)
            fmt.Println("Error hashing the password")
            return
        }

        // Insert the new user into the database
        _, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
        if err != nil {
            //http.Error(w, "Database error", http.StatusInternalServerError)
            fmt.Println("Database error")
            return
        }

        fmt.Println("User added")
        return
    }

    fmt.Println(error_s)
}

/* func LogOut(){
    //delete(session.Values, "Tester")
    fmt.Println("session: ", session)
    return
    
    //return to /public
} */

/* func Login(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Render login template
        tmpl, err := template.ParseFiles("login.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tmpl.Execute(w, nil)
    } else if r.Method == "POST" {
        // Process login form submission
        r.ParseForm()
        username := r.FormValue("username")
        password := r.FormValue("password")

        user, err := query_db("select * from user where username = ?", []any{username}, true)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        if user == nil {
            http.Error(w, "Invalid username", http.StatusBadRequest)
            return
        }

        // Assuming user is a map with key 'pw_hash'
        userMap := user.(map[any]any)
        pwHash := userMap["pw_hash"].(string)
        
        // Check password (assuming you have a function checkPasswordHash)
        if !checkPasswordHash(password, pwHash) {
            http.Error(w, "Invalid password", http.StatusBadRequest)
            return
        }

        // Set session data
        session, _ := store.Get(r, "session-name")
        session.Values["user_id"] = userMap["user_id"]
        session.Save(r, w)

        // Redirect to timeline
        http.Redirect(w, r, "/timeline", http.StatusSeeOther)
    }
} */
/* 
func registerHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Render the registration template
        tmpl, err := template.ParseFiles("register.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tmpl.Execute(w, nil)
    } else if r.Method == "POST" {
        // Process the registration form
        r.ParseForm()

        username := r.FormValue("username")
        email := r.FormValue("email")
        password := r.FormValue("password")
        password2 := r.FormValue("password2")

        var error string

        // Validate form input
        if username == "" {
            error = "You have to enter a username"
        } else if !strings.Contains(email, "@") {
            error = "You have to enter a valid email address"
        } else if password == "" {
            error = "You have to enter a password"
        } else if password != password2 {
            error = "The two passwords do not match"
        } else if _, err := get_user_id(username); err == nil {
            error = "The username is already taken"
        } else {
            // Hash the password
            hashedPassword, err := hashPassword(password)
            if err != nil {
                http.Error(w, "Error hashing password", http.StatusInternalServerError)
                return
            }

            // Insert the new user into the database
            _, err = db.Exec("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
            if err != nil {
                http.Error(w, "Database error", http.StatusInternalServerError)
                return
            }

            // Redirect to login page after successful registration
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Render the registration template with error message
        tmpl, err := template.ParseFiles("register.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tmpl.Execute(w, map[string]string{"Error": error})
    }
}
 */

// @app.route('/login', methods=['GET', 'POST'])
// def login():
//     """Logs the user in."""
//     if g.user:
//         return redirect(url_for('timeline'))
//     error = None
//     if request.method == 'POST':
//         user = query_db('''select * from user where
//             username = ?''', [request.form['username']], one=True)
//         if user is None:
//             error = 'Invalid username'
//         elif not check_password_hash(user['pw_hash'],
//                                      request.form['password']):
//             error = 'Invalid password'
//         else:
//             flash('You were logged in')
//             session['user_id'] = user['user_id']
//             return redirect(url_for('timeline'))
//     return render_template('login.html', error=error)

// @app.route('/register', methods=['GET', 'POST'])
// def register():
//     """Registers the user."""
//     if g.user:
//         return redirect(url_for('timeline'))
//     error = None
//     if request.method == 'POST':
//         if not request.form['username']:
//             error = 'You have to enter a username'
//         elif not request.form['email'] or \
//                  '@' not in request.form['email']:
//             error = 'You have to enter a valid email address'
//         elif not request.form['password']:
//             error = 'You have to enter a password'
//         elif request.form['password'] != request.form['password2']:
//             error = 'The two passwords do not match'
//         elif get_user_id(request.form['username']) is not None:
//             error = 'The username is already taken'
//         else:
//             g.db.execute('''insert into user (
//                 username, email, pw_hash) values (?, ?, ?)''',
//                 [request.form['username'], request.form['email'],
//                  generate_password_hash(request.form['password'])])
//             g.db.commit()
//             flash('You were successfully registered and can login now')
//             return redirect(url_for('login'))
//     return render_template('register.html', error=error)

// @app.route('/logout')
// def logout():
//     """Logs the user out"""
//     flash('You were logged out')
//     session.pop('user_id', None)
//     return redirect(url_for('public_timeline'))

// # add some filters to jinja and set the secret key and debug mode
// # from the configuration.
// app.jinja_env.filters['datetimeformat'] = format_datetime
// app.jinja_env.filters['gravatar'] = gravatar_url
// app.secret_key = SECRET_KEY
// app.debug = DEBUG

// if __name__ == '__main__':
//     app.run(host="0.0.0.0")

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
