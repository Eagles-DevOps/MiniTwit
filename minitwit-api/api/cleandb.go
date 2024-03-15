package api

import (
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit-api/db"
	"minitwit-api/sim"
)

func Cleandb(w http.ResponseWriter, r *http.Request) {
	db, err := db.GetDb()
	if err != nil {
		log.Fatalf("Could not get database: %v", err)
	}
	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	user_ids := make([]int, 4)
	usernames := []string{"a", "b", "c", "test"}

	for i, username := range usernames {
		user_id, _ := db.Get_user_id(username)
		user_ids[i] = user_id
	}

	for _, userID := range user_ids {
		if !db.IsZero(userID) {
			db.QueryDelete([]int{userID})
		}
	}
	w.WriteHeader(http.StatusOK)
}
