package model

type FilteredMessage struct {
	Content  string `json:"content"`
	Pub_date int    `json:"pub_date"`
	User     string `json:"user"`
}
