package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit.com/db"
	"minitwit.com/model"
	"minitwit.com/sim"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	sim.UpdateLatest(r)
	dec := json.NewDecoder(r.Body)
	var rv model.DeleteData
	err := dec.Decode(&rv)
	if err != nil {
		fmt.Println("Error in requestData")
	}
	fmt.Println("yeeeeet")
	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}

	if !db.IsNil(rv.User) && r.Method == "POST" {
		toDeleteUsername := rv.User
		toDeleteUser_id, _ := db.Get_user_id(toDeleteUsername)
		query := `DELETE FROM user WHERE user_id = ?`

		sqlite_db, _ := db.Connect_db()

		_, err = sqlite_db.Exec(query, toDeleteUser_id)

		fmt.Println(err)
	}
}
