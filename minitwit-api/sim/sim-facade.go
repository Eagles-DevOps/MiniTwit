package sim

import (
	"encoding/json"
	"minitwit-api/db"
	"minitwit-api/logger"
	"net/http"
	"strconv"
)

var lg = logger.InitializeLogger()

func UpdateLatest(r *http.Request) {
	r.ParseForm()
	latest := r.Form.Get("latest")
	if latest != "" {
		db, err := db.GetDb()
		if err != nil {
			lg.Fatal("Failed to get DB", err)
		}
		val, _ := strconv.Atoi(latest)
		db.SetCount("sim", val)
	}
}

func Is_authenticated(w http.ResponseWriter, r *http.Request) bool {
	from_simulator := r.Header.Get("Authorization")
	if from_simulator != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		errMsg := "You are not authorized to use this resource!"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		json.NewEncoder(w).Encode(struct {
			Status   int    `json:"status"`
			ErrorMsg string `json:"error_msg"`
		}{
			Status:   http.StatusForbidden,
			ErrorMsg: errMsg,
		})
		return false
	}
	return true
}
