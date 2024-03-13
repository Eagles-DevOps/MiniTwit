package sim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func UpdateLatest(r *http.Request) {
	r.ParseForm()
	latest := r.Form.Get("latest")
	if latest != "" {
		err := os.WriteFile("./latest_processed_sim_action_id.txt", []byte((latest)), 0644)
		if err != nil {
			fmt.Println("Error writing to ./latest_processed_sim_action_id.txt")
		}
	}

}

func Is_authenticated(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth != "Basic c2ltdWxhdG9yOnN1cGVyX3NhZmUh" {
		errMsg := "You are not authorized to use this resource!"

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
