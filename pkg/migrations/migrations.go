package migrations

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Version     string    `json:"version" gorm:"unique;not null"`
	Description string    `json:"description" gorm:"not null"`
	AppliedAt   time.Time `json:"applied_at" gorm:"autoCreateTime"`
}

// MigrationFunc represents a migration function
type MigrationFunc func(*gorm.DB) error

// MigrationItem represents a migration with its up and down functions
type MigrationItem struct {
	Version     string
	Description string
	Up          MigrationFunc
	Down        MigrationFunc
}

// Migrator handles database migrations
type Migrator struct {
	db         *gorm.DB
	migrations []MigrationItem
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: getAllMigrations(),
	}
}

// Run executes all pending migrations
func (m *Migrator) Run() error {
	// Create migrations table if it doesn't exist
	if err := m.db.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	var appliedMigrations []Migration
	if err := m.db.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %w", err)
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	// Apply pending migrations
	for _, migration := range m.migrations {
		if !appliedVersions[migration.Version] {
			log.Printf("Running migration: %s - %s", migration.Version, migration.Description)
			
			if err := migration.Up(m.db); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migration.Version, err)
			}

			// Record migration as applied
			migrationRecord := Migration{
				Version:     migration.Version,
				Description: migration.Description,
			}
			if err := m.db.Create(&migrationRecord).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}

			log.Printf("Migration completed: %s", migration.Version)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

// Rollback rolls back the last applied migration
func (m *Migrator) Rollback() error {
	var lastMigration Migration
	if err := m.db.Last(&lastMigration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("no migrations to rollback")
		}
		return fmt.Errorf("failed to fetch last migration: %w", err)
	}

	// Find migration definition
	var migrationItem *MigrationItem
	for _, migration := range m.migrations {
		if migration.Version == lastMigration.Version {
			migrationItem = &migration
			break
		}
	}

	if migrationItem == nil {
		return fmt.Errorf("migration definition not found for version: %s", lastMigration.Version)
	}

	if migrationItem.Down == nil {
		return fmt.Errorf("no rollback function defined for migration: %s", lastMigration.Version)
	}

	log.Printf("Rolling back migration: %s - %s", migrationItem.Version, migrationItem.Description)

	if err := migrationItem.Down(m.db); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", migrationItem.Version, err)
	}

	// Remove migration record
	if err := m.db.Delete(&lastMigration).Error; err != nil {
		return fmt.Errorf("failed to remove migration record %s: %w", migrationItem.Version, err)
	}

	log.Printf("Migration rolled back: %s", migrationItem.Version)
	return nil
}

// Status shows the current migration status
func (m *Migrator) Status() error {
	var appliedMigrations []Migration
	if err := m.db.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %w", err)
	}

	appliedVersions := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedVersions[migration.Version] = true
	}

	log.Println("Migration Status:")
	log.Println("=================")
	
	for _, migration := range m.migrations {
		status := "PENDING"
		if appliedVersions[migration.Version] {
			status = "APPLIED"
		}
		log.Printf("[%s] %s - %s", status, migration.Version, migration.Description)
	}

	return nil
}