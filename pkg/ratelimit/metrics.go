package ratelimit

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// RequestsTotal counts the total number of requests per role
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_requests_total",
			Help: "Total number of requests per role",
		},
		[]string{"role"},
	)

	// BlockedRequests counts the total number of blocked requests due to rate limiting
	BlockedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_blocked_total",
			Help: "Total number of blocked requests due to rate limiting",
		},
		[]string{"role"},
	)

	// RequestDuration measures the request duration per role
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rate_limit_request_duration_seconds",
			Help:    "Histogram of request duration per role",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"role"},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(BlockedRequests)
	prometheus.MustRegister(RequestDuration)
}