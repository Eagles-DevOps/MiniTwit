package postgres

import (
	"errors"
	"fmt"
	"log"
	"minitwit-api/logger"
	"minitwit-api/model"
	"net/url"
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gorm_logger "gorm.io/gorm/logger"
)

var lg = logger.InitializeLogger()

type PostgresDbImplementation struct {
	// Implement the methods defined in the Idb interface here
	db *gorm.DB
}

var (
	readWritesDatabase = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "minitwit_postgres_read_writes_total",
			Help: "Counts reads and writes to database.",
		},
		[]string{"func_name", "action", "status"},
	)
)
var (
	userGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_postgres_user_numbers_total",
			Help: "Counts the total number of users",
		},
	)
)
var (
	followerGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_postgres_follower_numbers_total",
			Help: "Counts the total number of followers",
		},
	)
)
var (
	messageGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_postgres_message_numbers_total",
			Help: "Counts the total number of message",
		},
	)
)

func (pgImpl *PostgresDbImplementation) Connect_db() {

	user := os.Getenv("POSTGRES_USER")
	pw := os.Getenv("POSTGRES_PW")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	dbname := os.Getenv("POSTGRES_DB_NAME")

	dsn := url.URL{
		User:   url.UserPassword(user, pw),
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   dbname,
	}

	newLogger := gorm_logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gorm_logger.Config{
			IgnoreRecordNotFoundError: true,
		},
	)
	var err error
	pgImpl.db, err = gorm.Open(postgres.Open(dsn.String()), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		lg.Error("Error connecting to the database ", err)
		readWritesDatabase.WithLabelValues("Connect_db", "connect", "fail").Inc()
		return
	}

	pgImpl.db.AutoMigrate(&model.User{}, &model.Follower{}, &model.Message{}, &model.Count{})
	readWritesDatabase.WithLabelValues("Connect_db", "connect", "success").Inc()
	lg.Info("Successfully connected to the database")

	userGauge.Set(pgImpl.QueryUserCount())
	followerGauge.Set(pgImpl.QueryFollowerCount())
	messageGauge.Set(pgImpl.QueryMessageCount())

}

func (pgImpl *PostgresDbImplementation) QueryUserCount() float64 { // To be called each time the counters are reset (when building the image)

	var count int64
	pgImpl.db.Model(&model.User{}).Count(&count)
	return float64(count)
}
func (pgImpl *PostgresDbImplementation) QueryMessageCount() float64 { // To be called each time the counters are reset (when building the image)

	var count int64
	pgImpl.db.Model(&model.Message{}).Count(&count)
	return float64(count)
}
func (pgImpl *PostgresDbImplementation) QueryFollowerCount() float64 { // To be called each time the counters are reset (when building the image)

	var count int64
	pgImpl.db.Model(&model.Follower{}).Count(&count)
	return float64(count)
}

func (pgImpl *PostgresDbImplementation) QueryRegister(args []string) {
	user := &model.User{
		Username: args[0],
		Email:    args[1],
		PwHash:   args[2],
	}
	res := pgImpl.db.Create(user)
	if res.Error != nil {
		lg.Error("Error registering user: ", res.Error)
		readWritesDatabase.WithLabelValues("QueryRegister", "write", "fail").Inc()
		return
	}
	userGauge.Inc()
	readWritesDatabase.WithLabelValues("QueryRegister", "write", "success").Inc()
	lg.Info("User registered successfully: ", user.Username)

}

func (pgImpl *PostgresDbImplementation) QueryMessage(message *model.Message) {
	res := pgImpl.db.Create(message)
	if res.Error != nil {
		lg.Error("Error creating message: ", res.Error)
		readWritesDatabase.WithLabelValues("QueryMessage", "write", "fail").Inc()
		return
	}
	messageGauge.Inc()
	readWritesDatabase.WithLabelValues("QueryMessage", "write", "success").Inc()
	lg.Info("Message created successfully: ", message.Text)
}

func (pgImpl *PostgresDbImplementation) QueryFollow(args []int) {
	follower := &model.Follower{
		WhoID:  args[0],
		WhomID: args[1],
	}
	res := pgImpl.db.Create(follower)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryFollow", "write", "fail").Inc()
		lg.Error("Error creating follower: ", res.Error)
		return
	}
	followerGauge.Inc()
	lg.Info("Follower created successfully: ", follower.WhoID, " -> ", follower.WhomID)
	readWritesDatabase.WithLabelValues("QueryFollow", "write", "success").Inc()
}

func (pgImpl *PostgresDbImplementation) QueryUnfollow(args []int) {
	res := pgImpl.db.Where("who_id = ? AND whom_id = ?", args[0], args[1]).Delete(&model.Follower{})
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryUnfollow", "write", "fail").Inc()
		lg.Error("Error unfollowing user: ", res.Error)
		return
	}
	followerGauge.Dec()
	lg.Info("User unfollowed successfully: ", args[0], " -> ", args[1])
	readWritesDatabase.WithLabelValues("QueryUnfollow", "write", "success").Inc()
}

func (pgImpl *PostgresDbImplementation) QueryDelete(args []int) {
	res := pgImpl.db.Delete(&model.User{}, args[0])
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("QueryDelete", "write", "fail").Inc()
		lg.Error("Error deleting user: ", res.Error)
		return
	}
	readWritesDatabase.WithLabelValues("QueryDelete", "write", "success").Inc()
	lg.Info("User deleted successfully: ", args[0])
}

func (pgImpl *PostgresDbImplementation) GetMessages(args []int) []map[string]any {
	var messages []model.Message
	res := pgImpl.db.Where("flagged = false").Order("pub_date DESC").Limit(args[0]).Find(&messages)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("GetMessages", "read", "fail").Inc()
		lg.Error("Error getting messages: ", res.Error)
		return []map[string]any{}
	}

	var Messages []map[string]any
	for _, msg := range messages {
		var user model.User
		pgImpl.db.First(&user, msg.AuthorID)

		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	readWritesDatabase.WithLabelValues("GetMessages", "read", "success").Inc()
	lg.Info("Messages retrieved successfully")
	return Messages
}

func (pgImpl *PostgresDbImplementation) GetMessagesForUser(args []int) []map[string]any {
	var messages []model.Message
	res := pgImpl.db.Where("flagged = false AND author_id = ?", args[0]).Order("pub_date DESC").Limit(args[1]).Find(&messages)
	if res.Error != nil {
		readWritesDatabase.WithLabelValues("GetMessagesForUser", "read", "fail").Inc()
		lg.Error("Error getting messages for user: ", res.Error)
	}

	var Messages []map[string]any

	for _, msg := range messages {
		var user model.User
		pgImpl.db.First(&user, msg.AuthorID)

		message := make(map[string]any)
		message["content"] = msg.Text
		message["pub_date"] = msg.PubDate
		message["user"] = user.Username

		Messages = append(Messages, message)
	}
	readWritesDatabase.WithLabelValues("GetMessagesForUser", "read", "success").Inc()
	lg.Info("Messages for user retrieved successfully")
	return Messages
}

func (pgImpl *PostgresDbImplementation) GetFollowees(args []int) []string {
	var followees []string
	res := pgImpl.db.Model(model.User{}).
		Select("username").
		Joins("inner join follower ON follower.whom_id = user_id").
		Where("follower.who_id = ?", args[0]).
		Limit(args[1]).
		Scan(&followees)

	if res.Error != nil {
		readWritesDatabase.WithLabelValues("GetFollowees", "read", "fail").Inc()
		lg.Error("Error getting followees: ", res.Error)
	}
	readWritesDatabase.WithLabelValues("GetFollowees", "read", "success").Inc()
	lg.Info("Followees retrieved successfully")
	return followees
}

func (pgImpl *PostgresDbImplementation) Get_user_id(username string) (int, error) {
	var user model.User
	res := pgImpl.db.Where("username = ?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			readWritesDatabase.WithLabelValues("Get_user_id", "read", "fail").Inc()
			return 0, fmt.Errorf("user with username '%s' not found", username)

		}
		readWritesDatabase.WithLabelValues("Get_user_id", "read", "fail").Inc()
		lg.Error("Error querying database: ", res.Error)
		return 0, fmt.Errorf("error querying database: %v", res.Error)
	}
	readWritesDatabase.WithLabelValues("Get_user_id", "read", "success").Inc()
	lg.Info("User found: ", username)
	return user.UserID, nil
}

func (pgImpl *PostgresDbImplementation) GetAllUsers() []model.User {
	var users []model.User
	pgImpl.db.Find(users)
	return users
}

func (pgImpl *PostgresDbImplementation) CreateUsers(users *[]model.User) error {
	res := pgImpl.db.CreateInBatches(&users, 100)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (pgImpl *PostgresDbImplementation) GetAllMessages() []model.Message {
	var messages []model.Message
	pgImpl.db.Find(messages)
	return messages
}

func (pgImpl *PostgresDbImplementation) CreateMessages(messages *[]model.Message) error {
	res := pgImpl.db.CreateInBatches(&messages, 100)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (pgImpl *PostgresDbImplementation) GetAllFollowers() []model.Follower {
	var followers []model.Follower
	pgImpl.db.Find(followers)
	return followers
}

func (pgImpl *PostgresDbImplementation) CreateFollowers(followers *[]model.Follower) error {
	res := pgImpl.db.CreateInBatches(&followers, 100)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (pgImpl *PostgresDbImplementation) GetCount(key string) int {
	var sim model.Count
	pgImpl.db.Where("key = ?", key).First(&sim)

	return sim.Value
}

func (pgImpl *PostgresDbImplementation) SetCount(key string, value int) error {
	// Upsert operation
	upsert := pgImpl.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},              // Unique columns
		DoUpdates: clause.AssignmentColumns([]string{"value"}), // Columns to update
	}).Create(&model.Count{Key: key, Value: value})

	if upsert.Error != nil {
		log.Fatalf("failed to upsert record: %v", upsert.Error)
		return upsert.Error
	}
	return nil
}

func (pgImpl *PostgresDbImplementation) IsNil(i interface{}) bool {
	if i == nil || i == interface{}(nil) {
		return true
	} else {
		return false
	}
}

func (pgImpl *PostgresDbImplementation) IsZero(i int) bool {
	if i == 0 {
		return true
	} else {
		return false
	}
}
