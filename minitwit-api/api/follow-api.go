package api

import (
	"encoding/json"
	"log"
	"minitwit-api/db"
	"minitwit-api/model"
	"net/http"
	"strconv"

	"minitwit-api/sim"

	"github.com/gorilla/mux"
)

func Follow(w http.ResponseWriter, r *http.Request) {
	db, err := db.GetDb()
	if err != nil {
		log.Fatalf("Could not get database: %v", err)
	}
	vars := mux.Vars(r)
	username := vars["username"]
	sim.UpdateLatest(r)

	is_auth := sim.Is_authenticated(w, r)
	if !is_auth {
		return
	}
	user_id, _ := db.Get_user_id(username)
	if db.IsZero(user_id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	no_flws := no_followees(r)

	var rv model.FollowData
	if r.Method == "POST" {
		err := json.NewDecoder(r.Body).Decode(&rv)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}
	if r.Method == "POST" && rv.Follow != "" {

		follow_username := rv.Follow
		follow_user_id, _ := db.Get_user_id(follow_username)

		if db.IsZero(follow_user_id) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		db.QueryFollow([]int{user_id, follow_user_id})
		w.WriteHeader(http.StatusNoContent)

	} else if r.Method == "POST" && rv.Unfollow != "" {

		unfollow_username := rv.Unfollow
		unfollow_user_id, err := db.Get_user_id(unfollow_username)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		db.QueryUnfollow([]int{user_id, unfollow_user_id})
		w.WriteHeader(http.StatusNoContent)

	} else if r.Method == "GET" {
		followees := db.GetFollowees([]int{user_id, no_flws})

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(struct {
			Follows []string `json:"follows"`
		}{
			Follows: followees,
		})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func no_followees(r *http.Request) int {
	value := r.URL.Query().Get("no")
	if value != "" {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
	}
	return 100
}
