package observability

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// =========================
	// HTTP Metrics
	// =========================

	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors.",
		},
		[]string{"method", "path"},
	)

	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// =========================
	// Database Metrics
	// =========================

	DBOpenConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_open_connections",
			Help: "Current number of open database connections.",
		},
	)

	DBInUseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_in_use_connections",
			Help: "Current number of database connections in use.",
		},
	)

	DBIdleConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_idle_connections",
			Help: "Current number of idle database connections.",
		},
	)

	DBWaitCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_wait_count_total",
			Help: "Total number of waits for a database connection.",
		},
	)

	DBWaitDuration = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "db_wait_duration_seconds_total",
			Help: "Total time spent waiting for database connections.",
		},
	)

	DBMaxOpenConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_max_open_connections",
			Help: "Maximum number of open database connections.",
		},
	)

	// =========================
	// Redis Metrics
	// =========================

	RedisOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations.",
		},
		[]string{"operation"},
	)

	RedisErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_errors_total",
			Help: "Total number of Redis errors.",
		},
		[]string{"operation"},
	)

	RedisOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// =========================
	// Worker Metrics
	// =========================

	WorkerEventsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_events_processed_total",
			Help: "Total number of worker events processed successfully.",
		},
	)

	WorkerErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_errors_total",
			Help: "Total number of worker processing errors.",
		},
	)

	WorkerRetries = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_retries_total",
			Help: "Total number of worker retries.",
		},
	)

	WorkerDLQ = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "worker_dlq_total",
			Help: "Total number of messages moved to the dead letter queue.",
		},
	)

	WorkerProcessingDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "worker_processing_duration_seconds",
			Help:    "Worker message processing duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func Register() {
	prometheus.MustRegister(
		// HTTP
		HTTPRequests,
		HTTPErrors,
		HTTPDuration,

		// Database
		DBOpenConnections,
		DBInUseConnections,
		DBIdleConnections,
		DBWaitCount,
		DBWaitDuration,
		DBMaxOpenConnections,

		// Redis
		RedisOperations,
		RedisErrors,
		RedisOperationDuration,

		// Worker
		WorkerEventsProcessed,
		WorkerErrors,
		WorkerRetries,
		WorkerDLQ,
		WorkerProcessingDuration,
	)
}

func RecordHTTP(
	method string,
	path string,
	status int,
	duration time.Duration,
) {
	HTTPRequests.WithLabelValues(
		method,
		path,
		strconv.Itoa(status),
	).Inc()

	if status >= 400 {
		HTTPErrors.WithLabelValues(
			method,
			path,
		).Inc()
	}

	HTTPDuration.WithLabelValues(
		method,
		path,
	).Observe(duration.Seconds())
}

func RecordDBStats(stats sql.DBStats) {
	DBOpenConnections.Set(
		float64(stats.OpenConnections),
	)

	DBInUseConnections.Set(
		float64(stats.InUse),
	)

	DBIdleConnections.Set(
		float64(stats.Idle),
	)

	DBWaitCount.Add(
		float64(stats.WaitCount),
	)

	DBWaitDuration.Add(
		stats.WaitDuration.Seconds(),
	)

	DBMaxOpenConnections.Set(
		float64(stats.MaxOpenConnections),
	)
}
