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

func Connect_db() (db *sql.DB, err error) {
	fmt.Println("Connecting to database...")
	return sql.Open("sqlite3", DATABASE)
}

func Query_db(query string, args []any, one bool) (any, error) {
	db, _ := Connect_db()
	cur, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer cur.Close()

	var rv []map[any]any
	cols, err := cur.Columns()
	if err != nil {
		return nil, fmt.Errorf("error retrieving columns: %w", err)
	}
	for cur.Next() {
		row := make([]any, len(cols))
		for i := range row {
			row[i] = new(any)
		}
		err = cur.Scan(row...)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		dict := make(map[any]any)
		for i, col := range cols {
			dict[col] = *(row[i].(*any))
		}
		rv = append(rv, dict)
		if one {
			break
		}
	}

	if err = cur.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	if len(rv) != 0 {
		if one {
			return rv[0], nil
		}
		return rv, nil
	}
	return nil, nil
}

func Get_user_id(username string) (any, error) {
	user_id, err := Query_db("SELECT user_id FROM user WHERE username = ?", []any{username}, true)
	if !isNil(user_id) {
		userID := user_id.(map[any]any)
		user_id_val := userID["user_id"]
		return user_id_val, err
	}
	return nil, err
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

func GetMessagesForUser(args []any, one bool) []model.FilteredMessage {
	query := `SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND
	user.user_id = message.author_id AND user.user_id = ?
	ORDER BY message.pub_date DESC LIMIT ?`

	db, _ := Connect_db()
	cur, _ := db.Query(query, args...)
	defer cur.Close()

	var Filtered []model.FilteredMessage

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

// ChatGPT
func isNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}
