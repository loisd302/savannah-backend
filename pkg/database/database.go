package database

import (
	"fmt"
	"log"
	"time"

	"backend/pkg/config"
	"backend/pkg/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Africa/Nairobi",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Set log level based on environment
	if cfg.Environment == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Error)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to database")
	return nil
}

// Migrate runs database migrations using explicit migration files
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	migrator := migrations.NewMigrator(DB)
	if err := migrator.Run(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigration rolls back the last applied migration
func RollbackMigration() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	migrator := migrations.NewMigrator(DB)
	return migrator.Rollback()
}

// MigrationStatus shows the current migration status
func MigrationStatus() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	migrator := migrations.NewMigrator(DB)
	return migrator.Status()
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}