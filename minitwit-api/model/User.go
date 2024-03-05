package model

type User struct {
	UserID   int
	Username string
	Email    string
	PwHash   string
}

func (User) TableName() string {
	return "user"
}
