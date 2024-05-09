package api

import (
	"encoding/json"
	"minitwit-api/db"
	"net/http"
)

func Get_latest(w http.ResponseWriter, r *http.Request) {
	lg.Info("Get latest handler invoked ")
	db, err := db.GetDb()
	if err != nil {
		lg.Fatal("Could not get database", err)
	}
	count := db.GetCount("sim")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Latest int `json:"latest"`
	}{
		Latest: count,
	})
}
