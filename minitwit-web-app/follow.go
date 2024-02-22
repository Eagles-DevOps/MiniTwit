package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DATABASE = "./minitwit.db"
	PER_PAGE = 30
)

var db *sql.DB

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/fllws/{username}", follow)

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	db, err := connect_db()
	if err != nil {
		fmt.Println("Error connecting to the database")
	}
	defer db.Close()
}

func connect_db() (db *sql.DB, err error) {
	fmt.Println("Connecting to database...")
	return sql.Open("sqlite3", DATABASE)
}

type RV struct {
	Message_id int
	Author_id  int
	Text       string
	Pub_date   int
	Flagged    bool
	User_id    int
	Username   string
	Email      string
	Pw_hash    string
}

type FilteredMsgs struct {
	Text     string
	Pub_date int
	Username string
}

func getMessages(args []any, one bool) []FilteredMsgs {
	query := `SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND
	user.user_id = message.author_id AND user.user_id = ?
	ORDER BY message.pub_date DESC LIMIT ?`

	db, _ := connect_db()
	cur, _ := db.Query(query, args...)
	defer cur.Close()

	var Filtered []FilteredMsgs

	for cur.Next() {
		var rv RV
		_ = cur.Scan(&rv.Message_id, &rv.Author_id, &rv.Text, &rv.Pub_date, &rv.Flagged, &rv.User_id, &rv.Username, &rv.Email, &rv.Pw_hash)

		println("values: ", rv.Message_id, rv.Author_id, rv.Text, rv.Pub_date, rv.Flagged, rv.User_id, rv.Username, rv.Email, rv.Pw_hash)

		filteredMsg := FilteredMsgs{
			Text:     rv.Text,
			Pub_date: rv.Pub_date,
			Username: rv.Username,
		}
		println("flitered: ", filteredMsg.Text, filteredMsg.Pub_date, filteredMsg.Username)
		Filtered = append(Filtered, filteredMsg)
		fmt.Println("result: ", Filtered)
	}
	return Filtered
}

func updateLatest(r *http.Request) {
	r.ParseForm()
	parsedCommandID := -1
	latest := r.Form.Get("latest")
	if latest != "" {
		parsedCommandID, _ = strconv.Atoi(latest)
	}
	if parsedCommandID != -1 {
		_ = os.WriteFile("./latest_processed_sim_action_id.txt", []byte(strconv.Itoa(parsedCommandID)), 0644)
	}
}

func is_req_from_simulator(w http.ResponseWriter, r *http.Request) bool {
	from_simulator := r.Header.Get("Authorization")
	errMsg := ""
	if from_simulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		errMsg = "You are not authorized to use this resource!"

		_ = json.NewEncoder(w).Encode(struct {
			Status   int    `json:"status"`
			ErrorMsg string `json:"error_msg"`
		}{
			Status:   403,
			ErrorMsg: errMsg,
		})
		return false
	}
	return true
}

// """Queries the database and returns a list of dictionaries."""
func query_db(query string, args []any, one bool) (any, error) {
	db, _ := connect_db()
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

// """Convenience method to look up the id for a username."""
func get_user_id(username string) (any, error) {
	user_id, err := query_db("SELECT user_id FROM user WHERE username = ?", []any{username}, true)
	if !isNil(user_id) {
		userID := user_id.(map[any]any)
		user_id_val := userID["user_id"]
		return user_id_val, err
	}
	return nil, err
}

func no_followers(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getFollowers(args []any, one bool) []string {
	query := `SELECT user.username FROM user
			INNER JOIN follower ON follower.whom_id=user.user_id
			WHERE follower.who_id=?
			LIMIT ?`

	db, _ := connect_db()
	cur, _ := db.Query(query, args...)
	defer cur.Close()
	var Filtered []string

	for cur.Next() {
		var username string
		_ = cur.Scan(&username)

		println("values: ", username)
		Filtered = append(Filtered, username)
	}
	return Filtered
}

type Followers_response struct {
	follows []string
}

func follow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	updateLatest(r)

	from_sim_response := is_req_from_simulator(w, r)
	if !from_sim_response {
		fmt.Println("inside")
		return
	}

	user_id, _ := get_user_id(username)
	if isNil(user_id) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	no_flws := no_followers(r, "no", 100)

	if r.Method == "POST" && r.URL.Query().Get("follow") != "" {

		follows_username := r.URL.Query().Get("follow")
		fmt.Println(follows_username)
		follows_user_id, _ := get_user_id(follows_username)

		if isNil(follows_user_id) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		query := `INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`
		db, _ := connect_db()
		_, err := db.Exec(query, user_id, follows_user_id)

		if err != nil {
			fmt.Println("Error querying the database")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(struct {
			Status int `json:"status"`
		}{
			Status: 204,
		})

	} else if r.Method == "POST" && r.URL.Query().Get("unfollow") != "" {

		unfollows_username := r.URL.Query().Get("unfollow")
		fmt.Println(unfollows_username)
		unfollows_user_id, err := get_user_id(unfollows_username)

		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		query := `DELETE FROM follower WHERE who_id=? and WHOM_id=?`
		db, _ := connect_db()
		_, err = db.Exec(query, user_id, unfollows_user_id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(struct {
			Status int `json:"status"`
		}{
			Status: 204,
		})

	} else if r.Method == "GET" {
		followers := getFollowers([]any{user_id, no_flws}, false)
		fmt.Println("Followers: ", followers)
		var followers_response Followers_response
		followers_response.follows = followers
		fmt.Println("Followers_response: ", followers_response)

		json.NewEncoder(w).Encode(struct {
			Followers []string `json:"content"`
		}{
			Followers: followers_response.follows,
		})
	}
}

// ChatGPT
func isNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}
