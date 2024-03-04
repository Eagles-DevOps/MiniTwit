package model

type Follower struct {
	WhoID  uint
	WhomID uint
}

func (Follower) TableName() string {
	return "follower"
}
