package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func Get_latest(w http.ResponseWriter, r *http.Request) {
	content, _ := os.ReadFile("./latest_processed_sim_action_id.txt")
	latest_command_id, err := strconv.Atoi(string(content))
	if err != nil {
		latest_command_id = -1
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Latest int `json:"latest"`
	}{
		Latest: latest_command_id,
	})
}
