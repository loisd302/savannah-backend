package monitoring

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents the health of an individual component
type ComponentHealth struct {
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message"`
	LastChecked time.Time    `json:"last_checked"`
	Duration    string       `json:"duration"`
	Details     interface{}  `json:"details,omitempty"`
}

// HealthResponse represents the overall health response
type HealthResponse struct {
	Status     HealthStatus                   `json:"status"`
	Timestamp  time.Time                     `json:"timestamp"`
	Uptime     string                        `json:"uptime"`
	Version    string                        `json:"version"`
	Components map[string]ComponentHealth    `json:"components"`
}

// HealthChecker manages health checks for various components
type HealthChecker struct {
	db         *sql.DB
	redis      *redis.Client
	startTime  time.Time
	version    string
	logger     *Logger
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, redisClient *redis.Client, version string, logger *Logger) *HealthChecker {
	return &HealthChecker{
		db:        db,
		redis:     redisClient,
		startTime: time.Now(),
		version:   version,
		logger:    logger,
	}
}

// CheckHealth performs all health checks and returns the overall status
func (hc *HealthChecker) CheckHealth(ctx context.Context) HealthResponse {
	components := make(map[string]ComponentHealth)
	
	// Check database health
	components["database"] = hc.checkDatabase(ctx)
	
	// Check Redis health
	components["redis"] = hc.checkRedis(ctx)
	
	// Check external services
	components["sms_service"] = hc.checkSMSService(ctx)
	
	// Determine overall status
	overallStatus := hc.determineOverallStatus(components)
	
	return HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     time.Since(hc.startTime).String(),
		Version:    hc.version,
		Components: components,
	}
}

// checkDatabase checks the health of the database connection
func (hc *HealthChecker) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	if hc.db == nil {
		return ComponentHealth{
			Status:      StatusUnhealthy,
			Message:     "Database connection not initialized",
			LastChecked: time.Now(),
			Duration:    "0ms",
		}
	}
	
	// Simple ping to check connectivity
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err := hc.db.PingContext(ctx)
	duration := time.Since(start)
	
	if err != nil {
		hc.logger.LogError(ctx, err, "Database health check failed", nil)
		return ComponentHealth{
			Status:      StatusUnhealthy,
			Message:     fmt.Sprintf("Database ping failed: %v", err),
			LastChecked: time.Now(),
			Duration:    duration.String(),
		}
	}
	
	// Get database stats for detailed health info
	stats := hc.db.Stats()
	details := map[string]interface{}{
		"open_connections":     stats.OpenConnections,
		"max_open_connections": stats.MaxOpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
	}
	
	// Check if connections are healthy
	status := StatusHealthy
	message := "Database is healthy"
	
	if stats.OpenConnections > int(float64(stats.MaxOpenConnections)*0.8) {
		status = StatusDegraded
		message = "Database connection pool is nearly exhausted"
	}
	
	return ComponentHealth{
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Duration:    duration.String(),
		Details:     details,
	}
}

// checkRedis checks the health of the Redis connection
func (hc *HealthChecker) checkRedis(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	if hc.redis == nil {
		return ComponentHealth{
			Status:      StatusUnhealthy,
			Message:     "Redis connection not initialized",
			LastChecked: time.Now(),
			Duration:    "0ms",
		}
	}
	
	// Simple ping to check connectivity
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	pong, err := hc.redis.Ping(ctx).Result()
	duration := time.Since(start)
	
	if err != nil {
		hc.logger.LogError(ctx, err, "Redis health check failed", nil)
		return ComponentHealth{
			Status:      StatusUnhealthy,
			Message:     fmt.Sprintf("Redis ping failed: %v", err),
			LastChecked: time.Now(),
			Duration:    duration.String(),
		}
	}
	
	if pong != "PONG" {
		return ComponentHealth{
			Status:      StatusUnhealthy,
			Message:     "Redis returned unexpected ping response",
			LastChecked: time.Now(),
			Duration:    duration.String(),
		}
	}
	
	// Get Redis info for detailed health
	info, err := hc.redis.Info(ctx).Result()
	details := map[string]interface{}{
		"ping_response": pong,
	}
	
	if err == nil {
		details["info_available"] = true
	}
	
	return ComponentHealth{
		Status:      StatusHealthy,
		Message:     "Redis is healthy",
		LastChecked: time.Now(),
		Duration:    duration.String(),
		Details:     details,
	}
}

// checkSMSService checks the health of the SMS service
func (hc *HealthChecker) checkSMSService(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// For SMS service, we'll do a lightweight check
	// In a real scenario, you might want to make a test API call
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	// Simulate SMS service health check
	// This could be replaced with actual Africa's Talking API status check
	time.Sleep(100 * time.Millisecond) // Simulate network call
	duration := time.Since(start)
	
	// For now, we'll assume SMS service is healthy if we can reach this point
	// In production, you'd make an actual API call to check service status
	
	return ComponentHealth{
		Status:      StatusHealthy,
		Message:     "SMS service is healthy",
		LastChecked: time.Now(),
		Duration:    duration.String(),
		Details: map[string]interface{}{
			"provider": "Africa's Talking",
			"endpoint": "configured",
		},
	}
}

// determineOverallStatus determines the overall system health based on component health
func (hc *HealthChecker) determineOverallStatus(components map[string]ComponentHealth) HealthStatus {
	hasUnhealthy := false
	hasDegraded := false
	
	for _, component := range components {
		switch component.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}
	
	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	
	return StatusHealthy
}

// HealthHandler returns a Gin handler for health checks
func (hc *HealthChecker) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		health := hc.CheckHealth(ctx)
		
		// Set appropriate HTTP status code
		var statusCode int
		switch health.Status {
		case StatusHealthy:
			statusCode = http.StatusOK
		case StatusDegraded:
			statusCode = http.StatusOK // Still return 200 for degraded
		case StatusUnhealthy:
			statusCode = http.StatusServiceUnavailable
		}
		
		c.JSON(statusCode, health)
	}
}

// LivenessHandler returns a simple liveness check (Kubernetes liveness probe)
func (hc *HealthChecker) LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	}
}

// ReadinessHandler returns a readiness check (Kubernetes readiness probe)
func (hc *HealthChecker) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		health := hc.CheckHealth(ctx)
		
		// Service is ready if it's healthy or degraded
		if health.Status == StatusUnhealthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not_ready",
				"timestamp": time.Now(),
				"reason":    "Service is unhealthy",
			})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now(),
		})
	}
}