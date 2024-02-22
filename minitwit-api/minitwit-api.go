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

	r.HandleFunc("/register", api.Register)
	r.HandleFunc("/msgs", api.Messages)

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
