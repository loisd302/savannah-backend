package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"backend/pkg/config"
	"backend/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	// Define command line flags
	var (
		action = flag.String("action", "up", "Migration action: up, down, status")
		help   = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	if err := database.InitDatabase(cfg); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.CloseDatabase()

	// Execute migration action
	switch *action {
	case "up":
		if err := database.Migrate(); err != nil {
			log.Fatal("Migration failed:", err)
		}
		fmt.Println("✅ Migrations completed successfully!")

	case "down":
		if err := database.RollbackMigration(); err != nil {
			log.Fatal("Migration rollback failed:", err)
		}
		fmt.Println("✅ Migration rolled back successfully!")

	case "status":
		if err := database.MigrationStatus(); err != nil {
			log.Fatal("Failed to get migration status:", err)
		}

	default:
		fmt.Printf("Unknown action: %s\n", *action)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Migration Tool")
	fmt.Println("==============")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/migrate.go -action=<action>")
	fmt.Println()
	fmt.Println("Actions:")
	fmt.Println("  up     - Run all pending migrations (default)")
	fmt.Println("  down   - Rollback the last migration")
	fmt.Println("  status - Show migration status")
	fmt.Println("  help   - Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/migrate.go -action=up")
	fmt.Println("  go run cmd/migrate.go -action=status")
	fmt.Println("  go run cmd/migrate.go -action=down")
}