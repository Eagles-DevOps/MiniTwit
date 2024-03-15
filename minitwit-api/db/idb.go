package db

import "minitwit-api/model"

type Idb interface {
	Connect_db()
	QueryMessage(message *model.Message)
	QueryFollow(args []int)
	QueryUnfollow(args []int)
	QueryDelete(args []int)
	QueryRegister(args []string)
	GetMessages(args []int) []map[string]any
	GetMessagesForUser(args []int) []map[string]any
	GetFollowees(args []int) []string
	Get_user_id(username string) (int, error)
	IsNil(i interface{}) bool
	IsZero(i int) bool

	GetAllUsers() []model.User
	GetAllFollowers() []model.Follower
	GetAllMessages() []model.Message
}
