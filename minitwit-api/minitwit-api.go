package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"minitwit-api/api"
	"minitwit-api/db"
	"minitwit-api/db/postgres"
	sqlite "minitwit-api/db/sqlitedb"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"go.uber.org/zap"
)

var (
	responseCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "minitwit_http_responses_total",
			Help: "The count of HTTP responses sent.",
		},
		[]string{"handler", "status", "method"},
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "minitwit_request_duration_milliseconds",
			Help: "Request duration distribution.",
		},
		[]string{"handler"},
	)
	cpuGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_cpu_load_percent",
			Help: "Current CPU load as a percentage",
		},
	)
)

func prometheusMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w}

		cpuLoad, err := getCPULoad()
		if err != nil {
			log.Printf("Error getting CPU load: %v", err)
		} else {
			cpuGauge.Set(cpuLoad)
		}

		next.ServeHTTP(rw, r)

		var handlerLabel string
		route := mux.CurrentRoute(r)
		if route != nil {
			name := route.GetName()
			if name != "" {
				handlerLabel = name
			}
		}
		responseCounter.WithLabelValues(handlerLabel, strconv.Itoa(rw.status), r.Method).Inc()

		timer := prometheus.NewTimer(requestDuration.WithLabelValues(handlerLabel))
		defer timer.ObserveDuration()
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func getCPULoad() (float64, error) {
	percentages, err := cpu.Percent(0, false)
	if err != nil || len(percentages) == 0 {
		return 0, err
	}
	return percentages[0], nil
}

func main() {

	pgImpl := &postgres.PostgresDbImplementation{}
	sqliteImpl := &sqlite.SqliteDbImplementation{}
	pgImpl.Connect_db()
	sqliteImpl.Connect_db()

	//logger set up
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"./logs/logs.log"}
	logger, err := config.Build()
	if err != nil {
		log.Fatal(err)
	}
	lg := logger.Sugar()

	dbType := os.Getenv("DBTYPE")

	if dbType == "postgres" {
		db.SetDb(pgImpl)
		//fmt.Println("Using postgres as main db")
		lg.Info("Using postgress as main DB.")
	} else {
		db.SetDb(sqliteImpl)
		//fmt.Println("Using sqlite as main db")
		lg.Info("Using SQLite as main DB.")

	}

	r := mux.NewRouter()

	r.Use(prometheusMiddleware)

	r.HandleFunc("/register", api.Register).Name("Register")
	r.HandleFunc("/msgs", api.Messages).Methods("GET").Name("Messages")
	r.HandleFunc("/msgs/{username}", api.Messages_per_user).Methods("GET", "POST").Name("Messages_per_user")
	r.HandleFunc("/fllws/{username}", api.Follow).Name("Follow")
	r.HandleFunc("/latest", api.Get_latest).Methods("GET").Name("Get_latest")
	r.HandleFunc("/cleandb", api.Cleandb).Name("Cleandb")
	r.HandleFunc("/delete", api.Delete).Name("Delete")

	r.Handle("/metrics", promhttp.Handler()).Name("Metrics")

	//fmt.Println("Listening on port 15001...")
	lg.Info("Listening on port 15001...")
	err = http.ListenAndServe(":15001", r)
	if err != nil {
		//log.Fatalf("Failed to start server: %v", err)
		lg.Fatal("Failed to start server: %v", err)
	}
}
