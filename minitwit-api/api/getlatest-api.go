package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func Get_latest(w http.ResponseWriter, r *http.Request) {
	lg.Info("Get latest handler invoked ")
	content, err := os.ReadFile("./latest_processed_sim_action_id.txt")
	if err != nil {
		lg.Error("Error reading sim file: ", err)
	}

	latest_command_id, err := strconv.Atoi(string(content))
	if err != nil {
		lg.Error("Error converting string to int: ", err)
		latest_command_id = -1
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Latest int `json:"latest"`
	}{
		Latest: latest_command_id,
	})
}
