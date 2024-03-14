package api

import (
	"fmt"
	"minitwit-api/db"
	"minitwit-api/db/postgres"
	sqlite "minitwit-api/db/sqlitedb"
	"net/http"
	"time"
)

func Migrate(w http.ResponseWriter, r *http.Request) {
	pgImpl := &postgres.PostgresDbImplementation{}
	sqliteImpl := &sqlite.SqliteDbImplementation{}

	pgImpl.Connect_db()
	sqliteImpl.Connect_db()

	start := time.Now()
	var users = sqliteImpl.GetAllUsers()
	err := pgImpl.CreateUsers(&users)
	if err != nil {
		fmt.Println(err)
	}
	elapsed := time.Since(start)

	fmt.Println("Users migrated in ", elapsed)

	start = time.Now()
	var followers = sqliteImpl.GetAllFollowers()
	fmt.Print(len(followers))
	err = pgImpl.CreateFollowers(&followers)
	if err != nil {
		fmt.Println(err)
	}
	elapsed = time.Since(start)

	fmt.Println("Followers migrated in ", elapsed)

	start = time.Now()
	var messages = sqliteImpl.GetAllMessages()
	fmt.Print(len(messages))
	err = pgImpl.CreateMessages(&messages)
	if err != nil {
		fmt.Println(err)
	}
	elapsed = time.Since(start)

	fmt.Println("Messages migrated in ", elapsed)

	db.SetDb(pgImpl)
}

func StatsPg(w http.ResponseWriter, r *http.Request) {
	pgImpl := &postgres.PostgresDbImplementation{}
	pgImpl.Connect_db()

	usersBefore := pgImpl.QueryUserCount()
	followersBefore := pgImpl.QueryFollowerCount()
	messagesBefore := pgImpl.QueryMessageCount()

	fmt.Printf("Users: %d\n", usersBefore)
	fmt.Printf("Followers: %d\n", followersBefore)
	fmt.Printf("Messages: %d\n", messagesBefore)
}

func StatsSqlite(w http.ResponseWriter, r *http.Request) {
	sqliteImpl := &sqlite.SqliteDbImplementation{}
	sqliteImpl.Connect_db()

	// Count entities before migration
	usersBefore := sqliteImpl.QueryUserCount()
	followersBefore := sqliteImpl.QueryFollowerCount()
	messagesBefore := sqliteImpl.QueryMessageCount()

	fmt.Printf("Users: %d\n", usersBefore)
	fmt.Printf("Followers: %d\n", followersBefore)
	fmt.Printf("Messages: %d\n", messagesBefore)
}
