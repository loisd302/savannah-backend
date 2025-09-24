package monitoring

import (
	"context"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional functionality
type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new structured logger
func NewLogger(environment string) *Logger {
	logger := logrus.New()

	// Set output format based on environment
	if environment == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
			ForceColors:     true,
		})
	}

	// Set log level based on environment
	switch environment {
	case "development":
		logger.SetLevel(logrus.DebugLevel)
	case "staging":
		logger.SetLevel(logrus.InfoLevel)
	case "production":
		logger.SetLevel(logrus.WarnLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.SetOutput(os.Stdout)

	return &Logger{
		Logger: logger,
	}
}

// WithContext adds context information to log entry
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.Logger.WithContext(ctx)

	// Add correlation ID if present
	if correlationID := ctx.Value("correlation_id"); correlationID != nil {
		entry = entry.WithField("correlation_id", correlationID)
	}

	// Add user ID if present
	if userID := ctx.Value("user_id"); userID != nil {
		entry = entry.WithField("user_id", userID)
	}

	// Add request ID if present
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}

	return entry
}

// HTTPMiddleware creates a Gin middleware for request logging
func (l *Logger) HTTPMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Generate correlation ID for request tracing
		correlationID := uuid.New().String()

		// Create structured log entry
		entry := l.WithFields(logrus.Fields{
			"correlation_id": correlationID,
			"method":         param.Method,
			"path":           param.Path,
			"status_code":    param.StatusCode,
			"latency":        param.Latency.String(),
			"client_ip":      param.ClientIP,
			"user_agent":     param.Request.UserAgent(),
			"response_size":  param.BodySize,
		})

		// Log based on status code
		if param.StatusCode >= 500 {
			entry.Error("HTTP request completed with server error")
		} else if param.StatusCode >= 400 {
			entry.Warn("HTTP request completed with client error")
		} else {
			entry.Info("HTTP request completed successfully")
		}

		// Add correlation ID to response headers for tracing
		param.Keys["correlation_id"] = correlationID

		return ""
	})
}

// LogError logs an error with context and stack trace
func (l *Logger) LogError(ctx context.Context, err error, message string, fields logrus.Fields) {
	entry := l.WithContext(ctx)
	
	if fields != nil {
		entry = entry.WithFields(fields)
	}
	
	entry.WithError(err).Error(message)
}

// LogInfo logs an info message with context
func (l *Logger) LogInfo(ctx context.Context, message string, fields logrus.Fields) {
	entry := l.WithContext(ctx)
	
	if fields != nil {
		entry = entry.WithFields(fields)
	}
	
	entry.Info(message)
}

// LogDebug logs a debug message with context
func (l *Logger) LogDebug(ctx context.Context, message string, fields logrus.Fields) {
	entry := l.WithContext(ctx)
	
	if fields != nil {
		entry = entry.WithFields(fields)
	}
	
	entry.Debug(message)
}

// LogWarn logs a warning message with context
func (l *Logger) LogWarn(ctx context.Context, message string, fields logrus.Fields) {
	entry := l.WithContext(ctx)
	
	if fields != nil {
		entry = entry.WithFields(fields)
	}
	
	entry.Warn(message)
}

// LogDatabaseOperation logs database operations for monitoring
func (l *Logger) LogDatabaseOperation(ctx context.Context, operation, table string, duration time.Duration, err error) {
	fields := logrus.Fields{
		"operation": operation,
		"table":     table,
		"duration":  duration.String(),
	}

	if err != nil {
		l.LogError(ctx, err, "Database operation failed", fields)
	} else {
		l.LogDebug(ctx, "Database operation completed", fields)
	}
}

// LogSMSOperation logs SMS operations for monitoring
func (l *Logger) LogSMSOperation(ctx context.Context, phoneNumber, message string, status string, err error) {
	fields := logrus.Fields{
		"phone_number": phoneNumber,
		"message_id":   uuid.New().String(), // Generate message ID for tracking
		"status":       status,
	}

	// Don't log the actual message content for privacy
	if err != nil {
		l.LogError(ctx, err, "SMS sending failed", fields)
	} else {
		l.LogInfo(ctx, "SMS sent successfully", fields)
	}
}

// LogJobProcessing logs background job processing
func (l *Logger) LogJobProcessing(ctx context.Context, jobType, jobID string, duration time.Duration, err error) {
	fields := logrus.Fields{
		"job_type": jobType,
		"job_id":   jobID,
		"duration": duration.String(),
	}

	if err != nil {
		l.LogError(ctx, err, "Job processing failed", fields)
	} else {
		l.LogInfo(ctx, "Job processed successfully", fields)
	}
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(ctx context.Context, eventType, description string, fields logrus.Fields) {
	securityFields := logrus.Fields{
		"event_type":  eventType,
		"description": description,
		"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Merge additional fields
	if fields != nil {
		for k, v := range fields {
			securityFields[k] = v
		}
	}

	entry := l.WithContext(ctx).WithFields(securityFields)
	entry.Warn("Security event detected")
}

// LogBusinessEvent logs important business events
func (l *Logger) LogBusinessEvent(ctx context.Context, eventType string, entityID string, fields logrus.Fields) {
	businessFields := logrus.Fields{
		"event_type": eventType,
		"entity_id":  entityID,
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
	}

	// Merge additional fields
	if fields != nil {
		for k, v := range fields {
			businessFields[k] = v
		}
	}

	entry := l.WithContext(ctx).WithFields(businessFields)
	entry.Info("Business event occurred")
}