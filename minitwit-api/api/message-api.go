package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"

	"minitwit.com/db"
	"minitwit.com/model"
)

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

func Messages(w http.ResponseWriter, r *http.Request) {
	updateLatest(r)

	from_sim_response := is_req_from_simulator(w, r)
	if !from_sim_response {
		return
	}
	no_msg := no_msgs(r, "no", 100)

	if r.Method == "GET" {
		messages := db.GetMessages([]any{no_msg}, false)

		for _, msg := range messages {
			if content, exists := msg["content"]; exists {
				if user, exists := msg["user"]; exists {
					if content == "Hello!" && user == "Viktoria" {
						fmt.Println("jackpot")
					}
				}
			}
		}

		/*responseData := struct {
			Msgs []map[string]any `json:"content"`
		}{
			Msgs: messages,
		}
		*/
		w.Header().Set("Content-Type", "application/json")

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
	updateLatest(r)

	from_sim_response := is_req_from_simulator(w, r)
	if !from_sim_response {
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

		for _, msg := range messages {
			fmt.Println("msg: ", msg)
			content, contentExists := msg["content"].(string)
			user, userExists := msg["user"].(string)

			if contentExists && userExists && content == "Hello!" && user == "Viktoria" {
				fmt.Println("jackpot")
			}
		}

		/*
			responseData := struct {
				Msgs []map[string]any `json:"content"`
			}{
				Msgs: messages,
			}
		*/

		w.Header().Set("Content-Type", "application/json")

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
		w.Header().Set("Content-Type", "application/json")
		fmt.Println("Executed query")
		w.WriteHeader(http.StatusNoContent)
	}
}

// TODO refactor out
func Get_latest(w http.ResponseWriter, r *http.Request) {
	content, _ := os.ReadFile("./latest_processed_sim_action_id.txt")
	latest_processed_command_id, err := strconv.Atoi(string(content))
	if err != nil {
		latest_processed_command_id = -1
	}
	err = json.NewEncoder(w).Encode(struct {
		Latest int `json:"latest"`
	}{
		Latest: latest_processed_command_id,
	})

}
