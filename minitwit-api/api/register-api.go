package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"minitwit.com/db"
	"minitwit.com/model"
)

func Register(w http.ResponseWriter, r *http.Request) {
	db.UpdateLatest(r)
	body, _ := io.ReadAll(r.Body)
	var rv model.RequestRegisterData
	err := json.Unmarshal([]byte(body), &rv)
	if err != nil {
		http.Error(w, "Error decoding JSON data", http.StatusBadRequest)
		return
	}
	fmt.Println("requestData: ", rv)

	if err != nil {
		fmt.Println("Error in requestData")
	}
	errMsg := ""

	if r.Method == "POST" {
		user_id, _ := db.Get_user_id(rv.Username)
		w.Header().Set("Content-Type", "application/json")

		if rv.Username == "" {
			errMsg = "You have to enter a username"
		} else if rv.Email == "" || !strings.Contains(rv.Email, "@") {
			errMsg = "You have to enter a valid email address"
		} else if rv.Pwd == "" {
			errMsg = "You have to enter a password"
		} else if !db.IsNil(user_id) {
			errMsg = "The username is already taken"
		} else {
			sqlite_db, err := db.Connect_db()
			defer sqlite_db.Close()
			if err != nil {
				fmt.Println("Error when connecting to the database")
				return
			}
			hash_pw, err := db.HashPassword(rv.Pwd)
			if err != nil {
				fmt.Println("Error hashing the password")
				return
			}
			query := "INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)"
			_, err = sqlite_db.Exec(query, rv.Username, rv.Email, hash_pw)
			if err != nil {
				fmt.Println("Error when trying to insert data into the database")
				return
			}
		}
		if errMsg != "" {
			Error := struct {
				Status int    `json:"status"`
				Msg    string `json:"error_msg"`
			}{
				Status: http.StatusBadRequest,
				Msg:    errMsg,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Error)
		} else {
			w.WriteHeader(http.StatusNoContent)
			fmt.Println("Queried")
		}
	}
}

func Is_req_from_simulator(w http.ResponseWriter, r *http.Request) bool {
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

func Follow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	db.UpdateLatest(r)
	dec := json.NewDecoder(r.Body)
	var rt model.Follow_resp
	err := dec.Decode(&rt)
	if err != nil {
		fmt.Println("Error in requestData")
	}
	from_sim_response := Is_req_from_simulator(w, r)
	if !from_sim_response {
		fmt.Println("inside")
		return
	}

	user_id, _ := db.Get_user_id(username)
	if db.IsNil(user_id) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	no_flws := no_followers(r, "no", 100)

	if r.Method == "POST" && rt.Follow != "" {

		follows_username := rt.Follow
		follows_user_id, _ := db.Get_user_id(follows_username)

		if db.IsNil(follows_user_id) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		query := `INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`
		sqlite_db, _ := db.Connect_db()
		defer sqlite_db.Close()
		_, err := sqlite_db.Exec(query, user_id, follows_user_id)

		if err != nil {
			fmt.Println("Error querying the database")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(http.StatusOK)

	} else if r.Method == "POST" && rt.Unfollow != "" {

		unfollows_username := rt.Unfollow
		unfollows_user_id, err := db.Get_user_id(unfollows_username)

		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		query := `DELETE FROM follower WHERE who_id=? and WHOM_id=?`
		sqlite_db, _ := db.Connect_db()
		defer sqlite_db.Close()
		_, err = sqlite_db.Exec(query, user_id, unfollows_user_id)

		json.NewEncoder(w).Encode(http.StatusOK)

	} else if r.Method == "GET" {
		followers := db.GetFollowers([]any{user_id, no_flws}, false)
		var followers_response model.Followers_response
		followers_response.Follows = followers

		json.NewEncoder(w).Encode(struct {
			Followers []string `json:"follows"`
		}{
			Followers: followers_response.Follows,
		})
	}
}

func no_followers(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
