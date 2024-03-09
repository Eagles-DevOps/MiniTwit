package main

import (
	"fmt"
	"log"
	"net/http"

	"minitwit-api/api"

	"github.com/gorilla/mux"

	"minitwit-api/db"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cpuGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_cpu_load_percent",
			Help: "Current load of the CPU in percent.",
		},
	)
	responseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "minitwit_http_responses_total",
			Help: "The count of HTTP responses sent.",
		},
		[]string{"status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "minitwit_request_duration_milliseconds",
			Help: "Request duration distribution.",
		},
		[]string{"endpoint"},
	)
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "myapp_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})
)

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func init() {
	prometheus.MustRegister(cpuGauge)
	prometheus.MustRegister(responseCounter)
	prometheus.MustRegister(requestDuration)
}

func main() {

	db.Connect_db()
	r := mux.NewRouter()

	r.Use(prometheusMiddleware)

	r.HandleFunc("/register", api.Register)
	r.HandleFunc("/msgs", api.Messages)
	r.HandleFunc("/msgs/{username}", api.Messages_per_user).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", api.Follow)
	r.HandleFunc("/latest", api.Get_latest).Methods("GET")
	r.HandleFunc("/cleandb", api.Cleandb)
	r.HandleFunc("/delete", api.Delete)

	http.Handle("/metrics", promhttp.Handler())
	r.Path("/obj/{id}").HandlerFunc( // Delete maybe?
		func(w http.ResponseWriter, r *http.Request) {})
	// err := http.ListenAndServe(":15002", nil)

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
