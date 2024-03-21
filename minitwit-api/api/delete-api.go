package api

import (
	"encoding/json"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit-api/db"
	"minitwit-api/model"
	"minitwit-api/sim"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	lg.Info("Delete handler invoked")
	db, err := db.GetDb()
	if err != nil {
		lg.Error("Could not get database: ", err)
	}
	sim.UpdateLatest(r)
	dec := json.NewDecoder(r.Body)
	var rv model.DeleteData
	err = dec.Decode(&rv)
	if err != nil {
		lg.Error("Error decoding request data: ", err)
	}
	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		lg.Warn("Unauthorized access attempt to Delete")
		return
	}
	if rv.User != "" && r.Method == "POST" {
		toDeleteUsername := rv.User
		lg.Info("Deleting user: ", toDeleteUsername)

		toDeleteUser_id, _ := db.Get_user_id(toDeleteUsername)
		lg.Info("User ID to delete: ", toDeleteUser_id)

		db.QueryDelete([]int{toDeleteUser_id})
		lg.Info("User deleted successfully")

	} else {
		lg.Warn("Invalid request: username missing or request method not POST")
	}

	lg.Info("Delete completed")
	w.WriteHeader(http.StatusOK)
}
