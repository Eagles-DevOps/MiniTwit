package api

import (
	"fmt"
	"minitwit-api/db"
	"minitwit-api/db/postgres"
	sqlite "minitwit-api/db/sqlitedb"
	"net/http"
<<<<<<< HEAD
	"time"
)

func Migrate(w http.ResponseWriter, r *http.Request) {
=======
	"os"
	"time"
)

var allowMigration = os.Getenv("ALLOWMIGRATION")

func Migrate(w http.ResponseWriter, r *http.Request) {

	if allowMigration != "true" {
		fmt.Println("Unallowed migration attempted")
		return
	}
	allowMigration = "false"

>>>>>>> main
	pgImpl := &postgres.PostgresDbImplementation{}
	sqliteImpl := &sqlite.SqliteDbImplementation{}

	pgImpl.Connect_db()
	sqliteImpl.Connect_db()

	start := time.Now()
	var users = sqliteImpl.GetAllUsers()
	fmt.Println("Users to migrate: ", len(users))

	var err error

	err = pgImpl.CreateUsers(&users)
	if err != nil {
		fmt.Println(err.Error())
	}

	elapsed := time.Since(start)

	fmt.Println("Users migrated in ", elapsed)

	start = time.Now()
	var followers = sqliteImpl.GetAllFollowers()
	fmt.Println("Followers to migrate: ", len(followers))

	for i := 0; i < len(followers); i += 10000 {
		if i+10000 >= len(followers) {
			flw := followers[i:]
			err = pgImpl.CreateFollowers(&flw)
		} else {
			flw := followers[i : i+10000]
			err = pgImpl.CreateFollowers(&flw)
		}
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	elapsed = time.Since(start)

	fmt.Println("Followers migrated in ", elapsed)

	start = time.Now()
	var messages = sqliteImpl.GetAllMessages()
	fmt.Print(len(messages))
	for i := 0; i < len(messages); i += 10000 {
		if i+10000 >= len(messages) {
			msgs := messages[i:]
			err = pgImpl.CreateMessages(&msgs)
		} else {
			msgs := messages[i : i+10000]
			err = pgImpl.CreateMessages(&msgs)
		}
		if err != nil {
			fmt.Println(err.Error())
		}
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
