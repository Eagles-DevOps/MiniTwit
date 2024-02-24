package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"minitwit.com/db"
	"minitwit.com/model"
)

func is_req_from_simulator1(w http.ResponseWriter, r *http.Request) bool {
	from_simulator := r.Header.Get("Authorization")
	errMsg := ""
	if from_simulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		errMsg = "You are not authorized to use this resource!"

		_ = json.NewEncoder(w).Encode(struct {
			Status   int    `json:"status"`
			ErrorMsg string `json:"error_msg"`
		}{
			Status:   403,
			ErrorMsg: errMsg,
		})
		return false
	}
	return true
}

func Delete(w http.ResponseWriter, r *http.Request) {
	db.UpdateLatest(r)
	dec := json.NewDecoder(r.Body)
	var rv model.Deletion_resp
	err := dec.Decode(&rv)
	if err != nil {
		fmt.Println("Error in requestData")
	}
	fmt.Println("yeeeeet")
	from_sim_response := is_req_from_simulator1(w, r)
	if !from_sim_response {
		fmt.Println("inside")
		return
	}

	if rv.User != "" && r.Method == "POST" {
		toDeleteUsername := rv.User
		toDeleteUser_id, _ := db.Get_user_id(toDeleteUsername)
		query := `DELETE FROM user WHERE user_id = ?`

		sqlite_db, _ := db.Connect_db()

		_, err = sqlite_db.Exec(query, toDeleteUser_id)

		fmt.Println(err)

	}

}
