package api

import (
	"fmt"
	"minitwit-api/db"
	"minitwit-api/db/postgres"
	"minitwit-api/db/sqlite"
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
	pgImpl.CreateUsers(&users)
	elapsed := time.Since(start)

	fmt.Println("Users migrated in ", elapsed)

	start = time.Now()
	var followers = sqliteImpl.GetAllFollowers()
	pgImpl.CreateFollowers(&followers)
	elapsed = time.Since(start)

	fmt.Println("Followers migrated in ", elapsed)

	start = time.Now()
	var messages = sqliteImpl.GetAllMessages()
	pgImpl.CreateMessages(&messages)
	elapsed = time.Since(start)

	fmt.Println("Messages migrated in ", elapsed)

	db.SetDb(pgImpl)

}
