package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit-api/db"
	"minitwit-api/sim"
)

func Cleandb(w http.ResponseWriter, r *http.Request) {
	sqlite_db, _ := db.Connect_db()
	defer sqlite_db.Close()

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	user_ids := make([]any, 4)
	usernames := []string{"a", "b", "c", "test"}

	for i, username := range usernames {
		user_id, _ := db.Get_user_id(username)
		user_ids[i] = user_id
	}

	for _, userID := range user_ids {
		if !db.IsNil(userID) {
			query := `DELETE FROM user WHERE user_id = ?`
			_, err := sqlite_db.Exec(query, userID)
			fmt.Println(userID, err)
		}
	}
	json.NewEncoder(w).Encode(http.StatusOK)
}
