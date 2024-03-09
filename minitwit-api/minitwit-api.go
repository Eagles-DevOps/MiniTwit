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
	cpuGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "minitwit_cpu_load_percent",
			Help: "Current load of the CPU in percent.",
		},
	)
	responseCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "minitwit_http_responses_total",
			Help: "The count of HTTP responses sent.",
		},
		[]string{"status"}, // "status" label required
	)
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "minitwit_request_duration_milliseconds",
			Help: "Request duration distribution.",
		},
		[]string{"path"}, // "path" label required
	)
)

// Middleware handler function for prometheus
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, _ := mux.CurrentRoute(r).GetPathTemplate()
		timer := prometheus.NewTimer(requestDuration.With(prometheus.Labels{"path": path}))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func main() {

	db.Connect_db()
	r := mux.NewRouter()

	// Adds middleware handlers
	r.Use(prometheusMiddleware) // Should come before other handlers

	r.HandleFunc("/register", api.Register)
	r.HandleFunc("/msgs", api.Messages)
	r.HandleFunc("/msgs/{username}", api.Messages_per_user).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", api.Follow)
	r.HandleFunc("/latest", api.Get_latest).Methods("GET")
	r.HandleFunc("/cleandb", api.Cleandb)
	r.HandleFunc("/delete", api.Delete)

	r.Handle("/metrics", promhttp.Handler())

	fmt.Println("Listening on port 15001...")
	err := http.ListenAndServe(":15001", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
