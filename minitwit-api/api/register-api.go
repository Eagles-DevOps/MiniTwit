package api

import (
	"encoding/json"
	"fmt"
	"minitwit-api/db"
	"minitwit-api/model"
	"minitwit-api/sim"
	"net/http"
	"strings"

	"github.com/cespare/xxhash"

	_ "github.com/mattn/go-sqlite3"
)

func Register(w http.ResponseWriter, r *http.Request) {
	sim.UpdateLatest(r)

	var rv model.RegisterData
	err := json.NewDecoder(r.Body).Decode(&rv)
	if err != nil {
		fmt.Println("Error in decoding the JSON, register", err)
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
			hash_pw := hashPassword(rv.Pwd)
			if err != nil {
				fmt.Println("Error hashing the password")
				return
			}

			db.DoExec("register", []any{rv.Username, rv.Email, hash_pw})
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

func hashPassword(password string) string {
	hashed := xxhash.Sum64([]byte(password))
	hashedStr := fmt.Sprintf("%d", hashed)

	return hashedStr
}
