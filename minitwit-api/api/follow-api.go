package api

import (
	"encoding/json"
	"minitwit-api/db"
	"minitwit-api/model"
	"net/http"
	"strconv"

	"minitwit-api/sim"

	"github.com/gorilla/mux"
)

func Follow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	user_id, _ := db.Get_user_id(username)
	if db.IsNil(user_id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	no_flws := no_followees(r, 100)

	var rv model.FollowData
	if r.Method == "POST" {
		err := json.NewDecoder(r.Body).Decode(&rv)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		}
	}

	if r.Method == "POST" && rv.Follow != "" {

		follow_username := rv.Follow
		follow_user_id, _ := db.Get_user_id(follow_username)

		if db.IsNil(follow_user_id) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		query := `INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`
		err := db.DoExec(query, []any{user_id, follow_user_id})

		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	} else if r.Method == "POST" && rv.Unfollow != "" {

		unfollow_username := rv.Unfollow
		unfollow_user_id, err := db.Get_user_id(unfollow_username)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		query := `DELETE FROM follower WHERE who_id=? and WHOM_id=?`

		err = db.DoExec(query, []any{user_id, unfollow_user_id})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	} else if r.Method == "GET" {
		followees := db.GetFollowees([]any{user_id, no_flws}, false)

		w.WriteHeader(http.StatusNoContent)
		err := json.NewEncoder(w).Encode(followees)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func no_followees(r *http.Request, defaultValue int) int {
	value := r.URL.Query().Get("no")
	if value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}
