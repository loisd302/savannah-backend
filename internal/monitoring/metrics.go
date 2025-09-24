package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Database metrics
	dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Redis metrics
	redisConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)

	redisQueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_queue_depth",
			Help: "Number of jobs in Redis queue",
		},
		[]string{"queue_name"},
	)

	redisJobsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_jobs_processed_total",
			Help: "Total number of Redis jobs processed",
		},
		[]string{"queue_name", "status"},
	)

	redisJobProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_job_processing_duration_seconds",
			Help:    "Duration of Redis job processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue_name"},
	)

	// Business metrics
	customersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "customers_total",
			Help: "Total number of customers",
		},
	)

	ordersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "orders_total",
			Help: "Total number of orders",
		},
	)

	smsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sms_sent_total",
			Help: "Total number of SMS messages sent",
		},
		[]string{"status"},
	)

	// Application metrics
	appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "environment", "build_date"},
	)

	appUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)
)

// Metrics holds all the metric collectors
type Metrics struct {
	startTime time.Time
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	m := &Metrics{
		startTime: time.Now(),
	}

	// Register all metrics
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		dbConnectionsActive,
		dbConnectionsIdle,
		dbQueriesTotal,
		dbQueryDuration,
		redisConnectionsActive,
		redisQueueDepth,
		redisJobsProcessed,
		redisJobProcessingDuration,
		customersTotal,
		ordersTotal,
		smsTotal,
		appInfo,
		appUptime,
	)

	return m
}

// HTTPMiddleware creates a Gin middleware for HTTP metrics
func (m *Metrics) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), statusCode).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

// Handler returns the Prometheus metrics handler
func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

// UpdateAppInfo sets application information metrics
func (m *Metrics) UpdateAppInfo(version, environment, buildDate string) {
	appInfo.WithLabelValues(version, environment, buildDate).Set(1)
}

// UpdateUptime updates the application uptime metric
func (m *Metrics) UpdateUptime() {
	uptime := time.Since(m.startTime).Seconds()
	appUptime.Set(uptime)
}

// Database Metrics Methods
func (m *Metrics) SetDBConnectionsActive(count float64) {
	dbConnectionsActive.Set(count)
}

func (m *Metrics) SetDBConnectionsIdle(count float64) {
	dbConnectionsIdle.Set(count)
}

func (m *Metrics) IncDBQueries(operation, table string) {
	dbQueriesTotal.WithLabelValues(operation, table).Inc()
}

func (m *Metrics) ObserveDBQueryDuration(operation, table string, duration float64) {
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// Redis Metrics Methods
func (m *Metrics) SetRedisConnectionsActive(count float64) {
	redisConnectionsActive.Set(count)
}

func (m *Metrics) SetRedisQueueDepth(queueName string, depth float64) {
	redisQueueDepth.WithLabelValues(queueName).Set(depth)
}

func (m *Metrics) IncRedisJobsProcessed(queueName, status string) {
	redisJobsProcessed.WithLabelValues(queueName, status).Inc()
}

func (m *Metrics) ObserveRedisJobProcessingDuration(queueName string, duration float64) {
	redisJobProcessingDuration.WithLabelValues(queueName).Observe(duration)
}

// Business Metrics Methods
func (m *Metrics) SetCustomersTotal(count float64) {
	customersTotal.Set(count)
}

func (m *Metrics) SetOrdersTotal(count float64) {
	ordersTotal.Set(count)
}

func (m *Metrics) IncSMSSent(status string) {
	smsTotal.WithLabelValues(status).Inc()
}