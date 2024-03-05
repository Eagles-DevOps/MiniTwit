package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit-api/db"
	"minitwit-api/model"
	"minitwit-api/sim"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	sim.UpdateLatest(r)
	dec := json.NewDecoder(r.Body)
	var rv model.DeleteData
	err := dec.Decode(&rv)
	if err != nil {
		fmt.Println("Error in requestData")
	}
	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	if rv.User != "" && r.Method == "POST" {
		toDeleteUsername := rv.User
		toDeleteUser_id, _ := db.Get_user_id(toDeleteUsername)
		db.QueryDelete([]int{toDeleteUser_id})
	}
}
