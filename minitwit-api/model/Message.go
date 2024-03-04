package model

type Message struct {
	MessageID uint
	AuthorID  uint
	Text      string
	PubDate   int
	Flagged   bool
}

func (Message) TableName() string {
	return "message"
}
