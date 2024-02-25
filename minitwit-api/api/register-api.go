package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"minitwit.com/db"
	"minitwit.com/model"
	"minitwit.com/sim"
)

func Register(w http.ResponseWriter, r *http.Request) {
	sim.UpdateLatest(r)

	var rv model.RequestRegisterData
	err := json.NewDecoder(r.Body).Decode(&rv)
	if err != nil {
		fmt.Println("Error in decoding the JSON", err)
	}

	if r.Method == "POST" {
		user_id, _ := db.Get_user_id(rv.Username)

		errMsg := ""

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
			Response := struct {
				Status int    `json:"status"`
				Msg    string `json:"error_msg"`
			}{
				Status: http.StatusBadRequest,
				Msg:    errMsg,
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
