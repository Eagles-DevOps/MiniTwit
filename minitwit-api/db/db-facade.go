package db

import (
	"database/sql"
	"fmt"
	"minitwit-api/model"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func Init() {
	fmt.Println("Initializing database...")
	query := `create table if not exists user (
		user_id integer primary key autoincrement,
		username string not null,
		email string not null,
		pw_hash string not null
	  );
	  
	  create table if not exists follower (
		who_id integer,
		whom_id integer
	  );
	  
	  create table if not exists message (
		message_id integer primary key autoincrement,
		author_id integer not null,
		text string not null,
		pub_date integer,
		flagged integer
	  );`

	db, _ := Connect_db()
	db.Exec(query)
	defer db.Close()
}

func Connect_db() (db *sql.DB, err error) {
	dbPath := os.Getenv("SQLITEPATH")
	if len(dbPath) == 0 {
		dbPath = "./sqlite/minitwit.db"
	}

	dir := filepath.Dir(dbPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if _ = os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory: %v\n", err)
		}
	}

	return sql.Open("sqlite3", dbPath)
}

func DoExec(endpoint string, args []any) error { //used for all post request
	db, _ := Connect_db()

	defer db.Close()
	query := ""
	switch endpoint {
	case "message":
		query = `INSERT INTO message (author_id, text, pub_date, flagged)
		VALUES (?, ?, ?, 0)`
	case "follow":
		query = `INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`
	case "unfollow":
		query = `DELETE FROM follower WHERE who_id=? and WHOM_id=?`
	case "delete":
		query = `DELETE FROM user WHERE user_id = ?`
	case "register":
		query = "INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)"
	}

	if query == "" {
		fmt.Println("Wrong endpoint given for POST request, can't fetch query")
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		fmt.Println("Error when trying to execute query:", query)
		fmt.Println("Error:", err)
		return err
	}
	return nil
}

func GetMessages(args []any, one bool) []map[string]any {
	query := `SELECT message.*, user.* FROM message, user
        WHERE message.flagged = 0 AND message.author_id = user.user_id
        ORDER BY message.pub_date DESC LIMIT ?`

	db, _ := Connect_db()
	defer db.Close()
	cur, _ := db.Query(query, args...)
	defer cur.Close()

	var Messages []map[string]any

	for cur.Next() {
		var rv model.UserMessageRow
		_ = cur.Scan(&rv.Message_id, &rv.Author_id, &rv.Text, &rv.Pub_date, &rv.Flagged, &rv.User_id, &rv.Username, &rv.Email, &rv.Pw_hash)

		msg := make(map[string]any)
		msg["content"] = rv.Text
		msg["pub_date"] = rv.Pub_date
		msg["user"] = rv.Username

		Messages = append(Messages, msg)
	}
	return Messages
}

func GetMessagesForUser(args []any, one bool) []map[string]any {
	query := `SELECT message.*, user.* FROM message, user
	WHERE message.flagged = 0 AND
	user.user_id = message.author_id AND user.user_id = ?
	ORDER BY message.pub_date DESC LIMIT ?`

	db, _ := Connect_db()
	defer db.Close()
	cur, _ := db.Query(query, args...)
	defer cur.Close()

	var Messages []map[string]any

	for cur.Next() {
		var rv model.UserMessageRow
		_ = cur.Scan(&rv.Message_id, &rv.Author_id, &rv.Text, &rv.Pub_date, &rv.Flagged, &rv.User_id, &rv.Username, &rv.Email, &rv.Pw_hash)

		msg := make(map[string]any)
		msg["content"] = rv.Text
		msg["pub_date"] = rv.Pub_date
		msg["user"] = rv.Username

		Messages = append(Messages, msg)
	}
	return Messages
}

func GetFollowees(args []any, one bool) []string {
	query := `SELECT user.username FROM user
			INNER JOIN follower ON follower.whom_id=user.user_id
			WHERE follower.who_id=?
			LIMIT ?`

	db, _ := Connect_db()
	defer db.Close()
	cur, _ := db.Query(query, args...)
	defer cur.Close()
	var Followees []string

	for cur.Next() {
		var username string
		_ = cur.Scan(&username)

		Followees = append(Followees, username)
	}
	return Followees
}

func Get_user_id(username string) (any, error) {
	user_id, err := Query_db("SELECT user_id FROM user WHERE username = ?", []any{username}, true)
	if IsNil(user_id) {
		return nil, fmt.Errorf("user with username '%s' not found: %w", username, err)
	}
	userID := user_id.(map[any]any)
	user_id_val := userID["user_id"]
	return user_id_val, err
}

func Query_db(query string, args []any, one bool) (any, error) {
	db, _ := Connect_db()
	defer db.Close()
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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func IsNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}
