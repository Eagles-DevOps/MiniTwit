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

func QueryRegister(args []any) {
	user := &model.User{
		Username: args[0].(string),
		Email:    args[1].(string),
		PwHash:   args[2].(string),
	}
	db.Create(user)
}

func QueryMessage(args []any) {
	message := &model.Message{
		AuthorID: args[0].(uint),
		Text:     args[1].(string),
		PubDate:  args[2].(int),
		Flagged:  false,
	}
	db.Create(message)
}

func QueryFollow(args []any) {
	follower := &model.Follower{
		WhoID:  args[0].(uint),
		WhomID: args[1].(uint),
	}
	db.Create(follower)
}

func QueryUnfollow(args []any) {
	db.Where("who_id = ? AND whom_id = ?", args[0].(uint), args[1].(uint)).Delete(&model.Follower{})
}

func QueryDelete(args []any) {
	db.Delete(&model.User{}, args[0].(uint))
}

func GetMessages(args []any, one bool) []map[string]any {
	var messages []model.Message
	db.Where("flagged = 0").Order("pub_date DESC").Limit(args[0].(int)).Find(&messages)

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

func GetMessagesForUser(args []any, one bool) []map[string]any {
	var messages []model.Message
	db.Where("flagged = 0 AND author_id = ?", args[0].(uint)).Order("pub_date DESC").Limit(args[1].(int)).Find(&messages)

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

func GetFollowees(args []any, one bool) []string {
	var followees []string
	db.Table("user").
		Select("user.username").
		Joins("inner join follower ON follower.whom_id=user.user_id").
		Where("follower.who_id = ?", args[0].(uint)).
		Limit(args[1].(int)).
		Scan(&followees)

	return followees
}

func Get_user_id(username string) (any, error) {
	var user model.User
	res := db.Where("username = ?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with username '%s' not found", username)
		}
		return nil, fmt.Errorf("error querying database: %v", res.Error)
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
