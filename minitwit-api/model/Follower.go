package model

type Follower struct {
	WhoID  int
	WhomID int
}

func (Follower) TableName() string {
	return "follower"
}
