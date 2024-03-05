package model

type Message struct {
	MessageID int    `gorm:"column:message_id;primaryKey"`
	AuthorID  int    `gorm:"column:author_id;not null"`
	Text      string `gorm:"column:text;not null"`
	PubDate   int    `gorm:"column:pub_date;"`
	Flagged   bool   `gorm:"column:flagged;"`
}

func (Message) TableName() string {
	return "message"
}
