package db

import (
	"errors"
	"fmt"
	"minitwit-api/model"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func Connect_db() {
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
	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to the database ", err)
		return
	}
	db.AutoMigrate(&model.User{}, &model.Follower{}, &model.Message{})
}

func QueryRegister(args []string) {
	user := &model.User{
		Username: args[0],
		Email:    args[1],
		PwHash:   args[2],
	}
	db.Create(user)
}

func QueryMessage(message *model.Message) {
	db.Create(message)
}

func QueryFollow(args []int) {
	follower := &model.Follower{
		WhoID:  args[0],
		WhomID: args[1],
	}
	db.Create(follower)
}

func QueryUnfollow(args []int) {
	db.Where("who_id = ? AND whom_id = ?", args[0], args[1]).Delete(&model.Follower{})
}

func QueryDelete(args []int) {
	db.Delete(&model.User{}, args[0])
}

func GetMessages(args []int) []map[string]any {
	var messages []model.Message
	db.Where("flagged = 0").Order("pub_date DESC").Limit(args[0]).Find(&messages)

	var Messages []map[string]any
	for _, msg := range messages {
		var user model.User
		db.First(&user, msg.AuthorID)

		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	return Messages
}

func GetMessagesForUser(args []int) []map[string]any {
	var messages []model.Message
	db.Where("flagged = 0 AND author_id = ?", args[0]).Order("pub_date DESC").Limit(args[1]).Find(&messages)

	var Messages []map[string]any

	for _, msg := range messages {
		var user model.User
		db.First(&user, msg.AuthorID)

		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	return Messages
}

func GetFollowees(args []int) []string {
	var followees []string
	db.Table("user").
		Select("user.username").
		Joins("inner join follower ON follower.whom_id=user.user_id").
		Where("follower.who_id = ?", args[0]).
		Limit(args[1]).
		Scan(&followees)

	return followees
}

func Get_user_id(username string) (int, error) {
	var user model.User
	res := db.Where("username = ?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("user with username '%s' not found", username)
		}
		return 0, fmt.Errorf("error querying database: %v", res.Error)
	}
	return user.UserID, nil
}

func IsNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}

func IsZero(i int) bool {
	if i == 0 {
		return true
	} else {
		return false
	}
}
