package api

import (
	"minitwit-api/logger"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit-api/db"
	"minitwit-api/sim"
)

var lg = logger.InitializeLogger()

func Cleandb(w http.ResponseWriter, r *http.Request) {
	lg.Info("Cleandb handler invoked")
	db, err := db.GetDb()
	if err != nil {
		lg.Error("Could not get database: ", err)
	}
	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		lg.Warn("Unauthorized access attempt to Cleandb")
		return
	}
	user_ids := make([]int, 4)
	usernames := []string{"a", "b", "c", "test"}

	for i, username := range usernames {
		user_id, _ := db.Get_user_id(username)
		user_ids[i] = user_id
		lg.Info("Retrieved user ID for username: ", username, ", ID: ", user_id)
	}

	for _, userID := range user_ids {
		if !db.IsZero(userID) {
			db.QueryDelete([]int{userID})
			lg.Info("Deleted user with ID: ", userID)
		} else {
			lg.Info("Skipping deletion for user ID: ", userID)
		}
	}
	lg.Info("Cleandb completed successfully")
	w.WriteHeader(http.StatusOK)
}
