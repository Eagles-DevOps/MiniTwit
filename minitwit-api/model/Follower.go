package model

type Follower struct {
	WhoID  int `gorm:"column:who_id;index:idx_member;primaryKey"`
	WhomID int `gorm:"column:whom_id;index:idx_member;primaryKey"`
}

func (Follower) TableName() string {
	return "follower"
}
