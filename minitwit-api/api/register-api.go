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
	lg.Info("Register handler invoked")
	db, err := db.GetDb()
	if err != nil {
		lg.Error("Could not get database: ", err)
	}
	sim.UpdateLatest(r)

	var rv model.RegisterData
	err = json.NewDecoder(r.Body).Decode(&rv)
	if err != nil {
		lg.Error("Error decoding request body: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		user_id, err := db.Get_user_id(rv.Username)
		if err != nil {
			lg.Error("Error fetching user ID: ", err)
		}

		errMsg := ""

		if rv.Username == "" {
			errMsg = "You have to enter a username"
		} else if rv.Email == "" || !strings.Contains(rv.Email, "@") {
			errMsg = "You have to enter a valid email address"
		} else if rv.Pwd == "" {
			errMsg = "You have to enter a password"
		} else if !db.IsZero(user_id) {
			errMsg = "The username is already taken"
		} else {
			hash_pw := hashPassword(rv.Pwd)
			db.QueryRegister([]string{rv.Username, rv.Email, hash_pw})
			lg.Info("User registered successfully", rv.Username)
			w.WriteHeader(http.StatusNoContent)
		}
		if errMsg != "" {
			lg.Error("Registration error: ", errMsg)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func hashPassword(password string) string {
	return fmt.Sprintf("%d", xxhash.Sum64([]byte(password)))
}
