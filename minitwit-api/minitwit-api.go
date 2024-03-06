package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"minitwit-api/api"

	"github.com/gorilla/mux"

	"minitwit-api/db"

	"os/exec"

	"github.com/robfig/cron/v3"
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

	c := cron.New()
	if c == nil {
		log.Fatal("Error creating cron instance")
	}
	c.AddFunc("*/15 * * * *", backup)
	c.Start()
	defer c.Stop()

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func backup() {
	fmt.Println("Starting backup of the database...")

	if err := os.MkdirAll("./backups", 0755); err != nil {
		fmt.Printf("Error creating destination directory: %s\n", err)
		return
	}
	cmd := exec.Command("scp", "-i", "~/.ssh/terraform", "-o", "StrictHostKeyChecking=no", "root@188.166.201.66:/tmp/sqlitedb-api/minitwit.db", "./backups/minitwit.db")

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running scp command: %s\n", err)
		return
	}

	fmt.Println("Backup completed successfully")
}
