package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/prometheus/client_golang/prometheus"
)

func connLimitMiddleware(size int) mux.MiddlewareFunc {
	budget := make(chan int, size)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case budget <- 0:
				next.ServeHTTP(w, r)
				<-budget
			default:
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("Connection limit reached"))
			}
		})
	}
}

func metricMiddleware(ignorePaths []string) mux.MiddlewareFunc {
	skipMap := make(map[string]bool, len(ignorePaths))
	for _, p := range ignorePaths {
		skipMap[p] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skipMap[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			reqCounter.Inc()
			start := time.Now()
			next.ServeHTTP(w, r)
			httpDurations.With(prometheus.Labels{
				"path": r.URL.Path,
			}).Observe(float64(time.Since(start).Seconds()) / 1e3)
		})
	}
}
