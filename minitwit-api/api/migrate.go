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

	// CLEAN POSTGRES
	pgImpl.DeleteAllData()
	fmt.Println("deleted all data")

	// MIGRATE USERS
	start := time.Now()
	users := sqliteImpl.GetAllUsers()
	fmt.Println("Users Count: ", len(users))
	err := pgImpl.CreateUsers(users)
	if err != nil {
		fmt.Println("Error migrating users:", err)
	}
	fmt.Println("Users migrated in ", time.Since(start))

	// MIGRATE FOLLOWERS
	start = time.Now()
	followers := sqliteImpl.GetAllFollowers()
	fmt.Println("Followers Count: ", len(followers))
	err = pgImpl.CreateFollowers(followers)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Followers migrated in ", time.Since(start))

	// MIGRATE MESSAGES
	start = time.Now()
	messages := sqliteImpl.GetAllMessages()
	fmt.Println("Messages Count: ", len(messages))
	err = pgImpl.CreateMessages(messages)
	if err != nil {
		fmt.Println("Error Creting:", err)
	}
	fmt.Println("Messages migrated in ", time.Since(start))

	// Set the db implementation to use
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
