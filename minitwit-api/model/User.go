package model

type User struct {
	UserID   uint   `gorm:"column:user_id;primaryKey"`
	Username string `gorm:"column:username;not null"`
	Email    string `gorm:"column:email;not null"`
	PwHash   string `gorm:"column:pw_hash;not null"`
}

func (User) TableName() string {
	return "user"
}
