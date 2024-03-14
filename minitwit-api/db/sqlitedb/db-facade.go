package sqlite

import (
	"errors"
	"fmt"
	"log"
	"minitwit-api/model"
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SqliteDbImplementation struct {
	// Implement the methods defined in the Idb interface here
	db *gorm.DB
}

var (
	readWritesDatabase = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "minitwit_database_read_writes_total",
			Help: "Counts reads and writes to database.",
		},
		[]string{"func_name", "action", "status"},
	)
)

/*
var (

	entityCounterDatabase = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "minitwit_postgres_entity_numbers_total",
			Help: "Counts the total number",
		},
		[]string{"entity_type"},
	)

)
*/
func (sqliteImpl *SqliteDbImplementation) Connect_db() {
	dbPath := os.Getenv("SQLITEPATH")
	if len(dbPath) == 0 {
		dbPath = "./sqlite/minitwit.db"
	}
	fmt.Println("dbPath set to:", dbPath)

	dir := filepath.Dir(dbPath)
	_, err := os.Stat(dir)

	if err == nil {
		fmt.Println("directory of the database exists")
	} else if os.IsNotExist(err) {
		fmt.Println("directory of the database does not exist, will create new one")
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Fatal Error: creating directory for db: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Println("db directory created")
		}
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			IgnoreRecordNotFoundError: true,
		},
	)

	sqliteImpl.db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		fmt.Println("Error connecting to the database ", err)
		readWritesDatabase.WithLabelValues("Connect_db", "connect", "fail").Inc()
		return
	}
	sqliteImpl.db.AutoMigrate(&model.User{}, &model.Follower{}, &model.Message{})
	readWritesDatabase.WithLabelValues("Connect_db", "connect", "success").Inc()
	// fmt.Println("user count is:")
	// fmt.Println(sqliteImpl.QueryUserCount())
	// fmt.Println("message count is:")
	// fmt.Println(sqliteImpl.QueryMessageCount())
	// fmt.Println("follower count is:")
	// fmt.Println(sqliteImpl.QueryFollowerCount())

}

func (sqliteImpl *SqliteDbImplementation) QueryUserCount() int { // To be called each time the counters are reset (when building the image)

	var count int64
	sqliteImpl.db.Model(&model.User{}).Count(&count)
	return int(count)
}
func (sqliteImpl *SqliteDbImplementation) QueryMessageCount() int { // To be called each time the counters are reset (when building the image)

	var count int64
	sqliteImpl.db.Model(&model.Message{}).Count(&count)
	return int(count)
}
func (sqliteImpl *SqliteDbImplementation) QueryFollowerCount() int { // To be called each time the counters are reset (when building the image)

	var count int64
	sqliteImpl.db.Model(&model.Follower{}).Count(&count)
	return int(count)
}
func (sqliteImpl *SqliteDbImplementation) QueryRegister(args []string) {
	user := &model.User{
		Username: args[0],
		Email:    args[1],
		PwHash:   args[2],
	}
	res := sqliteImpl.db.Create(user)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryRegister", "write", "fail").Inc()
	}
	readWritesDatabase.WithLabelValues("QueryRegister", "write", "success").Inc()
}

func (sqliteImpl *SqliteDbImplementation) QueryMessage(message *model.Message) {
	res := sqliteImpl.db.Create(message)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryMessage", "write", "fail").Inc()
	}
	readWritesDatabase.WithLabelValues("QueryMessage", "write", "success").Inc()

}

func (sqliteImpl *SqliteDbImplementation) QueryFollow(args []int) {
	follower := &model.Follower{
		WhoID:  args[0],
		WhomID: args[1],
	}
	res := sqliteImpl.db.Create(follower)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryFollow", "write", "fail").Inc()
	}
	readWritesDatabase.WithLabelValues("QueryFollow", "write", "success").Inc()
}

func (sqliteImpl *SqliteDbImplementation) QueryUnfollow(args []int) {
	res := sqliteImpl.db.Where("who_id = ? AND whom_id = ?", args[0], args[1]).Delete(&model.Follower{})
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryUnfollow", "write", "fail").Inc()
	}
	readWritesDatabase.WithLabelValues("QueryUnfollow", "write", "success").Inc()
}

func (sqliteImpl *SqliteDbImplementation) QueryDelete(args []int) {
	res := sqliteImpl.db.Delete(&model.User{}, args[0])
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryDelete", "write", "fail").Inc()
	}
	readWritesDatabase.WithLabelValues("QueryDelete", "write", "success").Inc()
}

func (sqliteImpl *SqliteDbImplementation) GetMessages(args []int) []map[string]any {
	var messages []model.Message
	res := sqliteImpl.db.Where("flagged = 0").Order("pub_date DESC").Limit(args[0]).Find(&messages)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("GetMessages", "read", "fail").Inc()
	}

	var Messages []map[string]any
	for _, msg := range messages {
		var user model.User
		sqliteImpl.db.First(&user, msg.AuthorID)

		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	readWritesDatabase.WithLabelValues("GetMessages", "read", "success").Inc()
	return Messages
}

func (sqliteImpl *SqliteDbImplementation) GetMessagesForUser(args []int) []map[string]any {
	var messages []model.Message
	res := sqliteImpl.db.Where("flagged = 0 AND author_id = ?", args[0]).Order("pub_date DESC").Limit(args[1]).Find(&messages)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("GetMessagesForUser", "read", "fail").Inc()
	}

	var Messages []map[string]any

	for _, msg := range messages {
		var user model.User
		sqliteImpl.db.First(&user, msg.AuthorID)

		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	readWritesDatabase.WithLabelValues("GetMessagesForUser", "read", "success").Inc()
	return Messages
}

func (sqliteImpl *SqliteDbImplementation) GetFollowees(args []int) []string {
	var followees []string
	res := sqliteImpl.db.Table("user").
		Select("user.username").
		Joins("inner join follower ON follower.whom_id=user.user_id").
		Where("follower.who_id = ?", args[0]).
		Limit(args[1]).
		Scan(&followees)

	if res.Error != nil {
		readWritesDatabase.WithLabelValues("GetFollowees", "read", "fail").Inc()
	}
	readWritesDatabase.WithLabelValues("GetFollowees", "read", "success").Inc()
	return followees
}

func (sqliteImpl *SqliteDbImplementation) Get_user_id(username string) (int, error) {
	var user model.User
	res := sqliteImpl.db.Where("username = ?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			readWritesDatabase.WithLabelValues("Get_user_id", "read", "fail").Inc()
			return 0, fmt.Errorf("user with username '%s' not found", username)

		}
		readWritesDatabase.WithLabelValues("Get_user_id", "read", "fail").Inc()
		return 0, fmt.Errorf("error querying database: %v", res.Error)
	}
	readWritesDatabase.WithLabelValues("Get_user_id", "read", "success").Inc()
	return user.UserID, nil
}

func (sqliteImpl *SqliteDbImplementation) GetAllUsers() []model.User {
	var allUsers []model.User
	if err := sqliteImpl.db.Find(&allUsers).Error; err != nil {
		fmt.Println("Error fetching users:", err)
		return nil
	}
	return allUsers
}

func (sqliteImpl *SqliteDbImplementation) GetAllMessages() []model.Message {
	var allMessages []model.Message
	if err := sqliteImpl.db.Find(&allMessages).Error; err != nil {
		fmt.Println("Error fetching messages:", err)
		return nil
	}
	return allMessages
}

func (sqliteImpl *SqliteDbImplementation) GetAllFollowers() []model.Follower {
	var allFollowers []model.Follower
	if err := sqliteImpl.db.Find(&allFollowers).Error; err != nil {
		fmt.Println("Error fetching messages:", err)
		return nil
	}
	return allFollowers
}

func (sqliteImpl *SqliteDbImplementation) IsNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}

func (sqliteImpl *SqliteDbImplementation) IsZero(i int) bool {
	if i == 0 {
		return true
	} else {
		return false
	}
}
