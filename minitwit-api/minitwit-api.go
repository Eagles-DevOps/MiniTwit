package main

import (
	"fmt"
	"log"
	"net/http"

	"minitwit-api/api"

	"github.com/gorilla/mux"

	"minitwit-api/db"
)

func main() {
	db.Connect_db()
	r := mux.NewRouter()

	r.HandleFunc("/register", api.Register)
	r.HandleFunc("/msgs", api.Messages)
	r.HandleFunc("/msgs/{username}", api.Messages_per_user).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", api.Follow)
	r.HandleFunc("/latest", api.Get_latest).Methods("GET")
	r.HandleFunc("/cleandb", api.Cleandb)
	r.HandleFunc("/delete", api.Delete)

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
