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
	"minitwit-api/logger"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
)

var lg = logger.InitializeLogger()

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
	lg.Info("Starting Minitwit API server")

	pgImpl := &postgres.PostgresDbImplementation{}
	sqliteImpl := &sqlite.SqliteDbImplementation{}

	dbType := os.Getenv("DBTYPE")

	if dbType == "postgres" {
		lg.Info("Using Postgress as main DB.")
		pgImpl.Connect_db()
		db.SetDb(pgImpl)
	} else {
		lg.Info("Using SQLite as main DB.")
		sqliteImpl.Connect_db()
		db.SetDb(sqliteImpl)
	}

	r := mux.NewRouter()
	r.Use(prometheusMiddleware)

	r.HandleFunc("/health", api.Health).Name("Health")
	r.HandleFunc("/stress", api.Stress).Name("Stress")
	r.HandleFunc("/register", api.Register).Name("Register")
	r.HandleFunc("/msgs", api.Messages).Methods("GET").Name("Messages")
	r.HandleFunc("/msgs/{username}", api.Messages_per_user).Methods("GET", "POST").Name("Messages_per_user")
	r.HandleFunc("/fllws/{username}", api.Follow).Name("Follow")
	r.HandleFunc("/latest", api.Get_latest).Methods("GET").Name("Get_latest")
	r.HandleFunc("/cleandb", api.Cleandb).Name("Cleandb")
	r.HandleFunc("/delete", api.Delete).Name("Delete")

	r.Handle("/metrics", promhttp.Handler()).Name("Metrics")

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "15001"
	}

	lg.Info("Listening on port:", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		lg.Fatal("Failed to start server: %v", err)
	}
}
