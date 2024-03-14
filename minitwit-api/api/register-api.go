package api

import (
	"encoding/json"
	"fmt"
	"log"
	"minitwit-api/db"
	"minitwit-api/model"
	"minitwit-api/sim"
	"net/http"
	"strings"

	"github.com/cespare/xxhash"

	_ "github.com/mattn/go-sqlite3"
)

func Register(w http.ResponseWriter, r *http.Request) {
	db, err := db.GetDb()
	if err != nil {
		log.Fatalf("Could not get database: %v", err)
	}
	sim.UpdateLatest(r)

	var rv model.RegisterData
	err = json.NewDecoder(r.Body).Decode(&rv)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
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
		} else if !db.IsZero(user_id) {
			errMsg = "The username is already taken"
		} else {
			hash_pw := hashPassword(rv.Pwd)
			db.QueryRegister([]string{rv.Username, rv.Email, hash_pw})
			w.WriteHeader(http.StatusNoContent)
		}
		if errMsg != "" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func hashPassword(password string) string {
	return fmt.Sprintf("%d", xxhash.Sum64([]byte(password)))
}
