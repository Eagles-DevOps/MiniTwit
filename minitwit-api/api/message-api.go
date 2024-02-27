package api

import (
	"encoding/json"
	"fmt"
	"minitwit-api/db"
	"minitwit-api/model"
	"net/http"
	"strconv"
	"time"

	"minitwit-api/sim"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func Messages(w http.ResponseWriter, r *http.Request) {
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	no_msg := no_msgs(r, 100)

	if r.Method == "GET" {
		messages := db.GetMessages([]any{no_msg}, false)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(messages)

		if err != nil {
			http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
			return
		}
	}
}

func Messages_per_user(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	no_msg := no_msgs(r, 100)

	user_id, err := db.Get_user_id(username)
	if err != nil {
		http.Error(w, "Error getting the user_id", http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		messages := db.GetMessagesForUser([]any{user_id, no_msg}, false)

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(messages)
		if err != nil {
			http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
			return
		}

	} else if r.Method == "POST" {
		var rv model.MessageData

		err := json.NewDecoder(r.Body).Decode(&rv)
		if err != nil {
			fmt.Println("Error in decoding the JSON, message", err)
		}

		sqlite_db, err := db.Connect_db()
		defer sqlite_db.Close()

		query := `INSERT INTO message (author_id, text, pub_date, flagged)
		VALUES (?, ?, ?, 0)`

		_, err = sqlite_db.Exec(query, user_id, rv.Content, int(time.Now().Unix()))
		if err != nil {
			fmt.Println("Error when trying to insert data into the database")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func no_msgs(r *http.Request, defaultValue int) int {
	value := r.URL.Query().Get("no")
	if value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}
