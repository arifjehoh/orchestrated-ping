package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP request duration in seconds
    HttpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "Duration of HTTP requests in seconds",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint", "status"})

    // HTTP requests total
    HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "endpoint", "status"})

    // Application uptime
    AppUptime = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "app_uptime_seconds",
        Help: "Application uptime in seconds",
    })
)