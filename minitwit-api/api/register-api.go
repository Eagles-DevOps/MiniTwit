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
		w.WriteHeader(http.StatusForbidden)
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
			query := "INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)"
			db.DoExec(query, []any{rv.Username, rv.Email, hash_pw})
		}
		if errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)
			err := json.NewEncoder(w).Encode(errMsg)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}
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
