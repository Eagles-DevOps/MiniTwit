package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DATABASE = "./minitwit.db"
	PER_PAGE = 30
)

var db *sql.DB

func main() {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	r.HandleFunc("/latest", get_latest).Methods("GET")

	fmt.Println("Listening on port 15000...")
	err := http.ListenAndServe(":15000", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	db, err := connect_db()
	if err != nil {
		fmt.Println("Error connecting to the database")
	}
	defer db.Close()
}

func connect_db() (db *sql.DB, err error) {
	fmt.Println("Connecting to database...")
	return sql.Open("sqlite3", DATABASE)
}

func get_latest(w http.ResponseWriter, r *http.Request) {
	content, _ := os.ReadFile("./latest_processed_sim_action_id.txt")
	latest_processed_command_id, err := strconv.Atoi(string(content))
	if err != nil {
		latest_processed_command_id = -1
	}
	err = json.NewEncoder(w).Encode(struct {
		Latest int `json:"latest"`
	}{
		Latest: latest_processed_command_id,
	})

}
