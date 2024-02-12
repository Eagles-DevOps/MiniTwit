package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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

func main() {
	db, _ = connect_db()

	r := mux.NewRouter()
	r.HandleFunc("/", handle)
	//r.HandleFunc("/", timeLine)
	//r.HandleFunc("/public", public_timeline)
	//r.HandleFunc("/<username>", user_timeline)
	//r.HandleFunc("/<username>/follow", follow_user)
	//r.HandleFunc("/<username>/unfollow", unfollow_user)
	//r.HandleFunc("/add_message", add_message).Methods("POST")
	//r.HandleFunc("/login", login)
	//r.HandleFunc("/register", register)
	//r.HandleFunc("/logout", logout)
	//http.Handle("/", r)
	content := query_db("SELECT user_id FROM user WHERE username IN (?, ?, ?)", []any{"Roger Histand", "Ayako Yestramski", "Leonora Alford"}, false)
	//dt := format_datetime(time.Now())
	//id_string := strconv.FormatInt(int64(id), 10)
	//output := gravatar_url("anam@itu.dk", 80)

	fmt.Println("Content: ", content)
	fmt.Print("Listening on port 5000...")
	http.ListenAndServe(":5000", r)
}

// "/"
func handle(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// """Returns a new connection to the database."""
func connect_db() (*sql.DB, error) {
	return sql.Open("sqlite3", DATABASE)
}

// """Creates the database tables."""
func init_db() ([]byte, error) {
	return os.ReadFile("schema.sql")
}

// """Queries the database and returns a list of dictionaries."""
// variable one must be false as a default
func query_db(query string, args []any, one bool) any {
	cur, err := db.Query(query, args...)
	if err != nil {
		return nil
	}
	defer cur.Close()

	var rv []map[any]any
	cols, err := cur.Columns()
	if err != nil {
		return nil
	}
	for cur.Next() {
		row := make([]any, len(cols))
		for i := range row {
			row[i] = new(any)
		}
		err = cur.Scan(row...)
		if err != nil {
			break
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
	if len(rv) != 0 {
		if one {
			return rv[0]
		}
		return rv
	}
	return nil
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
func get_user_id(username string) int {
	var user_id int
	rv := db.QueryRow("SELECT user_id FROM user WHERE username = ?",
		username)
	err := rv.Scan(&user_id)

	if err != sql.ErrNoRows {
		return user_id
	}
	return 0
}

// """Make sure we are connected to the database each request and look
// up the current user so that we know he's there.
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSIONKEY")))

func before_request(r *http.Request) {
	db, _ = connect_db()
	session, _ := store.Get(r, "session-name")
	user_id := session.Values["user_id"]
	fmt.Println("user_id: ", user_id)
	if user_id != nil {
		user := query_db("select * from user where user_id = ?", []any{"user_id"}, true)
		fmt.Println("user: ", user)
	}
}

// """Closes the database again at the end of the request."""
func after_request(response http.Response) http.Response {
	db.Close()
	return response
}

// @app.after_request
// def after_request(response):
//     """Closes the database again at the end of the request."""
//     g.db.close()
//     return response

// @app.route('/')
// def timeline():
//     """Shows a users timeline or if no user is logged in it will
//     redirect to the public timeline.  This timeline shows the user's
//     messages as well as all the messages of followed users.
//     """
//     print("We got a visitor from: " + str(request.remote_addr))
//     if not g.user:
//         return redirect(url_for('public_timeline'))
//     offset = request.args.get('offset', type=int)
//     return render_template('timeline.html', messages=query_db('''
//         select message.*, user.* from message, user
//         where message.flagged = 0 and message.author_id = user.user_id and (
//             user.user_id = ? or
//             user.user_id in (select whom_id from follower
//                                     where who_id = ?))
//         order by message.pub_date desc limit ?''',
//         [session['user_id'], session['user_id'], PER_PAGE]))

// @app.route('/public')
// def public_timeline():
//     """Displays the latest messages of all users."""
//     return render_template('timeline.html', messages=query_db('''
//         select message.*, user.* from message, user
//         where message.flagged = 0 and message.author_id = user.user_id
//         order by message.pub_date desc limit ?''', [PER_PAGE]))

// @app.route('/<username>')
// def user_timeline(username):
//     """Display's a users tweets."""
//     profile_user = query_db('select * from user where username = ?',
//                             [username], one=True)
//     if profile_user is None:
//         abort(404)
//     followed = False
//     if g.user:
//         followed = query_db('''select 1 from follower where
//             follower.who_id = ? and follower.whom_id = ?''',
//             [session['user_id'], profile_user['user_id']], one=True) is not None
//     return render_template('timeline.html', messages=query_db('''
//             select message.*, user.* from message, user where
//             user.user_id = message.author_id and user.user_id = ?
//             order by message.pub_date desc limit ?''',
//             [profile_user['user_id'], PER_PAGE]), followed=followed,
//             profile_user=profile_user)

// @app.route('/<username>/follow')
// def follow_user(username):
//     """Adds the current user as follower of the given user."""
//     if not g.user:
//         abort(401)
//     whom_id = get_user_id(username)
//     if whom_id is None:
//         abort(404)
//     g.db.execute('insert into follower (who_id, whom_id) values (?, ?)',
//                 [session['user_id'], whom_id])
//     g.db.commit()
//     flash('You are now following "%s"' % username)
//     return redirect(url_for('user_timeline', username=username))

// @app.route('/<username>/unfollow')
// def unfollow_user(username):
//     """Removes the current user as follower of the given user."""
//     if not g.user:
//         abort(401)
//     whom_id = get_user_id(username)
//     if whom_id is None:
//         abort(404)
//     g.db.execute('delete from follower where who_id=? and whom_id=?',
//                 [session['user_id'], whom_id])
//     g.db.commit()
//     flash('You are no longer following "%s"' % username)
//     return redirect(url_for('user_timeline', username=username))

// @app.route('/add_message', methods=['POST'])
// def add_message():
//     """Registers a new message for the user."""
//     if 'user_id' not in session:
//         abort(401)
//     if request.form['text']:
//         g.db.execute('''insert into message (author_id, text, pub_date, flagged)
//             values (?, ?, ?, 0)''', (session['user_id'], request.form['text'],
//                                   int(time.time())))
//         g.db.commit()
//         flash('Your message was recorded')
//     return redirect(url_for('timeline'))

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
