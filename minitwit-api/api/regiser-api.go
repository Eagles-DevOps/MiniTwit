package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"minitwit.com/db"
	"minitwit.com/model"
)

func Register(w http.ResponseWriter, r *http.Request) {
	db.UpdateLatest(r)
	dec := json.NewDecoder(r.Body)
	var rv model.RequestRegisterData
	err := dec.Decode(&rv)
	fmt.Println("requestData: ", rv)

	if err != nil {
		fmt.Println("Error in requestData")
	}
	errMsg := ""

	if r.Method == "POST" {

		if rv.Username == "" {
			errMsg = "You have to enter a username"
		}
		if rv.Email == "" || !strings.Contains(rv.Email, "@") {
			errMsg = "You have to enter a valid email address"
		}
		if rv.Pwd == "" {
			errMsg = "You have to enter a password"
		}
		user_id, _ := db.Get_user_id(rv.Username)
		if !db.IsNil(user_id) {
			errMsg = "The username is already taken"
		}
		hash_pw, err := db.HashPassword(rv.Pwd)
		if err != nil {
			fmt.Println("Error hashing the password")
			return
		}
		db, err := db.Connect_db()
		query := "INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)"
		_, err = db.Exec(query, rv.Username, rv.Email, hash_pw)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}
