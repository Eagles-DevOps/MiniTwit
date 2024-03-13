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
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}

	var rv model.DeleteData

	err := json.NewDecoder(r.Body).Decode(&rv)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if rv.User != "" && r.Method == "POST" {
		toDeleteUsername := rv.User
		toDeleteUser_id, err := db.Get_user_id(toDeleteUsername)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err = db.QueryDelete([]int{toDeleteUser_id})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}
