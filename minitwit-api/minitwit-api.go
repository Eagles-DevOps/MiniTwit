package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"minitwit-api/api"

	"github.com/gorilla/mux"

	"minitwit-api/db"

	"github.com/robfig/cron/v3"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
	c.AddFunc("* * * * *", backup)
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

	dbPath := os.Getenv("SQLITEPATH")
	if dbPath == "" {
		dbPath = "./sqlite/minitwit.db"
	}

	creds := credentials.NewStatic("DO00BXQ8HLVHRM7YL773", "dCk8uuNv7zAgZVW7mfvyIjzglZfBoGSbBoSecZskXJo", "", credentials.SignatureV4)

	minioClient, err := minio.New("minitwit.ams3.digitaloceanspaces.com", &minio.Options{
		Creds:  creds,
		Secure: true,
	})
	if err != nil {
		fmt.Printf("Error creating Minio client: %v\n", err)
		return
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	fileName := fmt.Sprintf("minitwit-%s.db", timestamp)

	_, err = minioClient.FPutObject(context.Background(), "minitwit", fileName, dbPath, minio.PutObjectOptions{})
	if err != nil {
		fmt.Printf("Error uploading database file: %v\n", err)
		return
	}

	fmt.Println("Backup completed successfully")
}
