package model

type Message struct {
	MessageID int
	AuthorID  int
	Text      string
	PubDate   int
	Flagged   bool
}

func (Message) TableName() string {
	return "message"
}
