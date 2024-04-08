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
	"gorm.io/gorm/clause"
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

var (
	sqliteUserGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_sql_user_numbers_total",
			Help: "Counts the totsqlal number of users",
		},
	)
)
var (
	sqliteFollowGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_sql_follow_numbers_total",
			Help: "Counts the total sql of followers",
		},
	)
)
var (
	sqlitemessageGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_sql_message_numbers_total",
			Help: "Counts the total number of message",
		},
	)
)

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
	sqliteImpl.db.AutoMigrate(&model.User{}, &model.Follower{}, &model.Message{}, &model.Count{})
	readWritesDatabase.WithLabelValues("Connect_db", "connect", "success").Inc()

	sqliteUserGauge.Set(sqliteImpl.QueryUserCount())
	sqliteFollowGauge.Set(sqliteImpl.QueryFollowerCount())
	sqlitemessageGauge.Set(sqliteImpl.QueryMessageCount())

}

func (sqliteImpl *SqliteDbImplementation) QueryUserCount() float64 { // To be called each time the counters are reset (when building the image)

	var count int64
	sqliteImpl.db.Model(&model.User{}).Count(&count)
	return float64(count)
}
func (sqliteImpl *SqliteDbImplementation) QueryMessageCount() float64 { // To be called each time the counters are reset (when building the image)

	var count int64
	sqliteImpl.db.Model(&model.Message{}).Count(&count)
	return float64(count)
}
func (sqliteImpl *SqliteDbImplementation) QueryFollowerCount() float64 { // To be called each time the counters are reset (when building the image)

	var count int64
	sqliteImpl.db.Model(&model.Follower{}).Count(&count)
	return float64(count)
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
	sqliteUserGauge.Inc()
	readWritesDatabase.WithLabelValues("QueryRegister", "write", "success").Inc()
}

func (sqliteImpl *SqliteDbImplementation) QueryMessage(message *model.Message) {
	res := sqliteImpl.db.Create(message)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryMessage", "write", "fail").Inc()
	}
	sqlitemessageGauge.Inc()
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
	sqliteFollowGauge.Inc()
	readWritesDatabase.WithLabelValues("QueryFollow", "write", "success").Inc()
}

func (sqliteImpl *SqliteDbImplementation) QueryUnfollow(args []int) {
	res := sqliteImpl.db.Where("who_id = ? AND whom_id = ?", args[0], args[1]).Delete(&model.Follower{})
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryUnfollow", "write", "fail").Inc()
	}
	sqliteFollowGauge.Dec()
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
	var users []model.User

	// Perform batched retrieval
	res := sqliteImpl.db.Find(&users)

	if res.Error != nil {
		fmt.Println("Error:", res.Error)
	}

	return users
}

func (sqliteImpl *SqliteDbImplementation) GetAllMessages() []model.Message {
	var messages []model.Message

	res := sqliteImpl.db.Find(&messages)

	if res.Error != nil {
		fmt.Println("Error:", res.Error)
	}

	return messages
}

func (sqliteImpl *SqliteDbImplementation) GetAllFollowers() []model.Follower {
	var followers []model.Follower
	res := sqliteImpl.db.Find(&followers)

	if res.Error != nil {
		fmt.Println("Error:", res.Error)
	}
	return followers
}

// GetCount implements db.Idb.
func (sqliteImpl *SqliteDbImplementation) GetCount(key string) int {
	var sim model.Count
	sqliteImpl.db.Where("key = ?", key).First(&sim)

	return sim.Value
}

// SetCount implements db.Idb.
func (sqliteImpl *SqliteDbImplementation) SetCount(key string, value int) error {
	// Upsert operation
	upsert := sqliteImpl.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},              // Unique columns
		DoUpdates: clause.AssignmentColumns([]string{"value"}), // Columns to update
	}).Create(&model.Count{Key: key, Value: value})

	if upsert.Error != nil {
		log.Fatalf("failed to upsert record: %v", upsert.Error)
		return upsert.Error
	}
	return nil
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
