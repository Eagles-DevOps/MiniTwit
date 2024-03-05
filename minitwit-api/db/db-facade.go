package db

import (
	"database/sql"
	"fmt"
	"log"
	"minitwit-api/model"
	"os"
	"path/filepath"

	"github.com/go-gorp/gorp"
)

var db *gorp.DbMap

func Connect_db() *gorp.DbMap {
	dbPath := os.Getenv("SQLITEPATH")
	if len(dbPath) == 0 {
		dbPath = "./sqlite/minitwit.db"
	}
	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if _ = os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
		}
	}
	sqliteDb, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	db = &gorp.DbMap{Db: sqliteDb, Dialect: gorp.SqliteDialect{}}

	db.AddTableWithName(model.User{}, "user").SetKeys(true, "UserID")
	db.AddTableWithName(model.Follower{}, "follower")
	db.AddTableWithName(model.Message{}, "message").SetKeys(true, "MessageID")

	if err := db.CreateTablesIfNotExists(); err != nil {
		log.Fatalf("Error creating tables: %v", err)
	}
	return db
}

func QueryRegister(args []string) {

	user := &model.User{
		Username: args[0],
		Email:    args[1],
		PwHash:   args[2],
	}
	db.Insert(user)
}

func QueryMessage(message *model.Message) {

	db.Insert(message)
}

func QueryFollow(args []int) {

	follower := &model.Follower{
		WhoID:  args[0],
		WhomID: args[1],
	}
	db.Insert(follower)
}

func QueryUnfollow(args []any) {
	db.Exec("DELETE FROM follower WHERE WhoID=? AND WhomID=?", args...)
}

func QueryDelete(args []any) {
	db.Exec("DELETE FROM message WHERE AuthorID=?", args...)
	db.Exec("DELETE FROM follower WHERE WhoID=? OR WhomID=?", args...)
	db.Exec("DELETE FROM user WHERE UserID=?", args...)
}

func GetMessages(args []int, one bool) []map[string]any {
	var messages []model.Message

	_, err := db.Select(&messages, "SELECT * FROM message WHERE Flagged = 0 ORDER BY PubDate DESC LIMIT ?", args[0])
	if err != nil {
		log.Fatalf("Error retrieving messages: %v", err)
	}
	var Messages []map[string]any
	for _, msg := range messages {
		var user model.User
		err := db.SelectOne(&user, "SELECT * FROM user WHERE UserID=?", msg.AuthorID)
		if err != nil {
			log.Fatalf("Error retrieving user: %v", err)
		}
		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	return Messages
}

func GetMessagesForUser(args []int, one bool) []map[string]any {
	var messages []model.Message

	_, err := db.Select(&messages, "SELECT * FROM message WHERE Flagged = 0 AND AuthorID = ? ORDER BY PubDate DESC LIMIT ?", args[0], args[1])
	if err != nil {
		log.Fatalf("Error retrieving messages: %v", err)
	}

	var Messages []map[string]any
	for _, msg := range messages {
		var user model.User
		err := db.SelectOne(&user, "SELECT * FROM user WHERE UserID=?", msg.AuthorID)
		if err != nil {
			log.Fatalf("Error retrieving user: %v", err)
		}
		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	return Messages
}

func GetFollowees(args []int, one bool) []string {
	var followees []string

	_, err := db.Select(&followees, `
		SELECT user.Username 
		FROM user 
		INNER JOIN follower 
		ON follower.WhomID = user.UserID 
		WHERE follower.WhoID = ? 
		LIMIT ?`, args[0], args[1])
	if err != nil {
		log.Fatalf("Error retrieving followees: %v", err)
	}
	return followees
}

func Get_user_id(username string) (int, error) {
	var userID int

	err := db.SelectOne(&userID, "SELECT UserID FROM user WHERE Username=?", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user with username '%s' not found", username)
		}
		return 0, fmt.Errorf("error querying database: %v", err)
	}
	return userID, nil
}

func IsNil(i any) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}

func IsUserIDZero(i int) bool {
	if i == 0 {
		return true
	} else {
		return false
	}
}
