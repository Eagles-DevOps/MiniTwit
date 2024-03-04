package model

type User struct {
	UserID   uint
	Username string
	Email    string
	PwHash   string
}

func (User) TableName() string {
	return "user"
}
