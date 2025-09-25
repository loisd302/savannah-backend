package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/auth"
	"backend/internal/docs"
	"backend/internal/jobs"
	"backend/internal/services"
	"backend/pkg/config"
	"backend/pkg/database"
	"backend/pkg/middleware"
	"backend/pkg/routes"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	log.Println("Connecting to database...")
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Run database migrations
	log.Println("Running database migrations...")
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// Initialize Redis client
	log.Println("Connecting to Redis...")
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("SMS job processing will be disabled")
	} else {
		log.Println("Redis connected successfully")
	}

	// Initialize job queue
	jobQueue := jobs.NewRedisJobQueue(redisClient)

	// Initialize SMS service
	smsConfig := &services.SMSConfig{
		Username:   cfg.SMS.Username,
		APIKey:     cfg.SMS.APIKey,
		Shortcode:  cfg.SMS.Shortcode,
		BaseURL:    cfg.SMS.BaseURL,
		IsSandbox:  cfg.SMS.IsSandbox,
		RetryLimit: cfg.SMS.RetryLimit,
		RetryDelay: 30 * time.Second,
	}
	smsService := services.NewSMSService(smsConfig, jobQueue)

	// Initialize OIDC provider (if configured)
	var oidcProvider *auth.OIDCProvider
	if cfg.OIDC.IssuerURL != "" && cfg.OIDC.ClientID != "" {
		log.Println("Initializing OIDC provider...")
		oidcConfig := &auth.OIDCConfig{
			IssuerURL:    cfg.OIDC.IssuerURL,
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			RedirectURL:  cfg.OIDC.RedirectURL,
			Scopes:       cfg.OIDC.Scopes,
		}
		var err error
		oidcProvider, err = auth.NewOIDCProvider(oidcConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize OIDC provider: %v", err)
			log.Println("Authentication will be disabled")
		}
	} else {
		log.Println("OIDC configuration not provided, authentication disabled")
	}

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Setup Swagger documentation routes
	docs.SetupSwaggerRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check database connection
		sqlDB, err := database.GetDB().DB()
		dbStatus := "ok"
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "error"
		}

		// Check Redis connection
		redisStatus := "ok"
		if redisClient.Ping(ctx).Err() != nil {
			redisStatus = "error"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "OK",
			"message":   "Server is running",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
			"services": gin.H{
				"database":     dbStatus,
				"redis":        redisStatus,
				"sms_service":  "ok",
				"auth_enabled": oidcProvider != nil,
			},
		})
	})

	// Setup routes (with conditional auth)
	if oidcProvider != nil {
		routes.SetupRoutes(router, database.GetDB(), oidcProvider, smsService)
	} else {
		// Setup routes without authentication for development
		router.GET("/api/v1/*path", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "authentication_required",
				"message": "OIDC provider not configured. Please set OIDC environment variables.",
			})
		})
	}

	// Start SMS job processor in background
	if redisClient.Ping(ctx).Err() == nil {
		go func() {
			log.Println("Starting SMS job processor...")
			if err := smsService.ProcessSMSJobs(ctx); err != nil {
				log.Printf("SMS job processor stopped: %v", err)
			}
		}()
	}

	// Setup graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("📖 API Documentation: http://localhost:%s/docs", cfg.Port)
		log.Printf("💚 Health Check: http://localhost:%s/health", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close database connection
	if err := database.CloseDatabase(); err != nil {
		log.Println("Error closing database:", err)
	}

	log.Println("✅ Server exited")
}
