package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"minitwit.com/api"
)

func main() {

	r := mux.NewRouter()

	//r.HandleFunc("/register", api.Register)
	r.HandleFunc("/msgs", api.Messages)
	r.HandleFunc("/latest", api.Get_latest).Methods("GET")
	r.HandleFunc("/{username}", api.Messages_per_user)

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
