package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"minitwit.com/db"
	"minitwit.com/model"
	"minitwit.com/sim"
)

func Messages(w http.ResponseWriter, r *http.Request) {
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	no_msg := no_msgs(r, "no", 100)

	if r.Method == "GET" {
		messages := db.GetMessages([]any{no_msg}, false)

		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)

		if err := encoder.Encode(messages); err != nil {
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
	no_msg := no_msgs(r, "no", 100)

	if r.Method == "GET" {
		user_id, err := db.Get_user_id(username)
		if err != nil {
			http.Error(w, "Error getting the user_id", http.StatusNotFound)
			return
		}
		messages := db.GetMessagesForUser([]any{user_id, no_msg}, false)

		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)

		if err := encoder.Encode(messages); err != nil {
			http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {

		body, _ := io.ReadAll(r.Body)

		var rv model.RequestMessageData

		err := json.Unmarshal(body, &rv)
		if err != nil {
			fmt.Println("Error decoding JSON data:", err)
			http.Error(w, "Error decoding JSON data", http.StatusBadRequest)
			return
		}

		user_id, err := db.Get_user_id(username)
		if err != nil {
			fmt.Println("Error getting the user_id")
			http.Error(w, "Error getting the user_id", http.StatusNotFound)
			return
		}
		sqlite_db, err := db.Connect_db()
		defer sqlite_db.Close()
		query := `INSERT INTO message (author_id, text, pub_date, flagged)
		VALUES (?, ?, ?, 0)`

		_, err = sqlite_db.Exec(query, user_id, rv.Content, int(time.Now().Unix()))
		if err != nil {
			fmt.Println("Error when trying to insert data into the database")
			fmt.Println(err)
			return
		}
		fmt.Println("Executed query")
		w.WriteHeader(http.StatusNoContent)
	}
}

func no_msgs(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
