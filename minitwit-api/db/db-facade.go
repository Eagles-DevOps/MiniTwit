package db

import (
	"database/sql"
	"fmt"

	"minitwit.com/model"
)

const (
	DATABASE = "./minitwit.db"
	PER_PAGE = 30
)

//var db *sql.DB

func Connect_db() (db *sql.DB, err error) {
	fmt.Println("Connecting to database...")
	return sql.Open("sqlite3", DATABASE)
}

func GetMessages(args []any, one bool) []model.FilteredMessage {
	query := `SELECT message.*, user.* FROM message, user
        WHERE message.flagged = 0 AND message.author_id = user.user_id
        ORDER BY message.pub_date DESC LIMIT ?`

	db, _ := Connect_db()
	cur, _ := db.Query(query, args...)
	defer cur.Close()

	var Filtered []model.FilteredMessage

	// TODO handle empty db

	for cur.Next() {
		var rv model.UserMessageRow
		_ = cur.Scan(&rv.Message_id, &rv.Author_id, &rv.Text, &rv.Pub_date, &rv.Flagged, &rv.User_id, &rv.Username, &rv.Email, &rv.Pw_hash)

		println("values: ", rv.Message_id, rv.Author_id, rv.Text, rv.Pub_date, rv.Flagged, rv.User_id, rv.Username, rv.Email, rv.Pw_hash)

		filteredMsg := model.FilteredMessage{
			Text:     rv.Text,
			Pub_date: rv.Pub_date,
			Username: rv.Username,
		}
		println("flitered: ", filteredMsg.Text, filteredMsg.Pub_date, filteredMsg.Username)
		Filtered = append(Filtered, filteredMsg)
		fmt.Println("result: ", Filtered)
	}
	return Filtered
}
