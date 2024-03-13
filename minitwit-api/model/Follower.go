package model

type Follower struct {
	WhoID  int `gorm:"column:who_id;index:idx_member"`
	WhomID int `gorm:"column:whom_id;index:idx_member"`
}

func (Follower) TableName() string {
	return "follower"
}
