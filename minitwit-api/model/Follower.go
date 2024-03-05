package model

type Follower struct {
	WhoID  int `gorm:"column:who_id;"`
	WhomID int `gorm:"column:whom_id;"`
}

func (Follower) TableName() string {
	return "follower"
}
