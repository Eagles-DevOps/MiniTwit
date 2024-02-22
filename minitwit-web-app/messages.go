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

	r.HandleFunc("/messages", messages)

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
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
        WHERE message.flagged = 0 AND message.author_id = user.user_id
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

func no_msgs(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func messages(w http.ResponseWriter, r *http.Request) {
	updateLatest(r)

	from_sim_response := is_req_from_simulator(w, r)
	if !from_sim_response {
		fmt.Println("inside")
		return
	}
	no_msg := no_msgs(r, "no", 100)

	if r.Method == "GET" {
		messages := getMessages([]any{no_msg}, false)
		println("msgs: ", messages)

		err := json.NewEncoder(w).Encode(struct {
			Status int            `json:"status"`
			Msgs   []FilteredMsgs `json:"content"`
		}{
			Status: 200,
			Msgs:   messages,
		})
		fmt.Println("error: ", err)
	}
}
