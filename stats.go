package main

import "github.com/prometheus/client_golang/prometheus"

var (
	wsStat = prometheus.NewGauge(prometheus.GaugeOpts{
		Subsystem: "websocket",
		Name:      "conn_active",
		Help:      "Total number of active websocket connections",
	})

	httpDurations = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_req_latency_sec",
		Help:    "Http latency distributions.",
		Buckets: prometheus.LinearBuckets(300, 500, 3),
	}, []string{"path"})

	reqCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Subsystem: "http",
		Name:      "req_total",
		Help:      "Total number of requests received",
	})
)

func init() {
	prometheus.MustRegister(httpDurations)
	prometheus.MustRegister(reqCounter)
	prometheus.MustRegister(wsStat)
}
