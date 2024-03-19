package api

import (
	"encoding/json"
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
	lg.Info("Messages handler invoked")
	db, err := db.GetDb()
	if err != nil {
		lg.Error("Could not get database", err)
	}
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		lg.Warn("Unauthorized access attempt to Messages")
		return
	}
	no_msg := no_msgs(r)

	if r.Method == "GET" {
		messages := db.GetMessages([]int{no_msg})

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(messages)

		if err != nil {
			lg.Error("Error encoding JSON response:", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
}

func Messages_per_user(w http.ResponseWriter, r *http.Request) {
	db, err := db.GetDb()
	if err != nil {
		lg.Error("Could not get database - Messages per user", err)
	}
	vars := mux.Vars(r)
	username := vars["username"]
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		lg.Warn("Unauthorized access attempt to Messages_perUser")
		return
	}
	no_msg := no_msgs(r)

	user_id, err := db.Get_user_id(username)
	if err != nil {
		lg.Error("Error getting user ID", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		messages := db.GetMessagesForUser([]int{user_id, no_msg})

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(messages)
		if err != nil {
			lg.Error("Error encoding JSON response: ", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

	} else if r.Method == "POST" {
		var rv model.MessageData

		err := json.NewDecoder(r.Body).Decode(&rv)
		if err != nil {
			lg.Error("Error decoding request body: ", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		message := &model.Message{
			AuthorID: user_id,
			Text:     rv.Content,
			PubDate:  int(time.Now().Unix()),
			Flagged:  false,
		}
		db.QueryMessage(message)
		lg.Info("Message posted", user_id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func no_msgs(r *http.Request) int {
	value := r.URL.Query().Get("no")
	if value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			lg.Error("Error parsing 'no' query parameter: ", err)
			return intValue
		}
	}
	return 100
}
