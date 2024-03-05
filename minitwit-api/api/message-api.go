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
		messages := db.GetMessages([]int{no_msg}, false)
		err := json.NewEncoder(w).Encode(messages)

		if err != nil {
			w.WriteHeader(http.StatusForbidden)
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
		w.WriteHeader(http.StatusForbidden)
		fmt.Println("error in user_id")
		return
	}

	if r.Method == "GET" {
		messages := db.GetMessagesForUser([]int{user_id, no_msg}, false)

		err = json.NewEncoder(w).Encode(messages)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

	} else if r.Method == "POST" {
		var rv model.MessageData

		err := json.NewDecoder(r.Body).Decode(&rv)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Println("error decoding mesg")
			return
		}
		message := &model.Message{
			AuthorID: user_id,
			Text:     rv.Content,
			PubDate:  int(time.Now().Unix()),
			Flagged:  false,
		}
		db.QueryMessage(message)
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
