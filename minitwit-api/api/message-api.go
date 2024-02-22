package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

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
		fmt.Println("inside")
		return
	}
	no_msg := no_msgs(r, "no", 100)

	if r.Method == "GET" {
		messages := db.GetMessages([]any{no_msg}, false)
		println("msgs: ", messages)

		err := json.NewEncoder(w).Encode(struct {
			Status int                     `json:"status"`
			Msgs   []model.FilteredMessage `json:"content"`
		}{
			Status: 200,
			Msgs:   messages,
		})
		fmt.Println("error: ", err)
	}
}
