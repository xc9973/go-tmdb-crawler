package main

import (
	"fmt"
	"log"

	"github.com/xc9973/go-tmdb-crawler/config"
	"github.com/xc9973/go-tmdb-crawler/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Open database connection
	dbPath := cfg.GetSQLitePath()
	fmt.Printf("Applying migration to database: %s\n", dbPath)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Apply migration for Show model (which includes correction fields)
	fmt.Println("Migrating Show model with correction fields...")
	if err := db.AutoMigrate(&models.Show{}); err != nil {
		log.Fatalf("Failed to migrate Show model: %v", err)
	}

	// Verify the columns exist
	fmt.Println("\nVerifying correction fields in shows table...")
	if !db.Migrator().HasColumn(&models.Show{}, "refresh_threshold") {
		log.Fatal("Column 'refresh_threshold' not found")
	}
	fmt.Println("✓ Column 'refresh_threshold' exists")

	if !db.Migrator().HasColumn(&models.Show{}, "stale_detected_at") {
		log.Fatal("Column 'stale_detected_at' not found")
	}
	fmt.Println("✓ Column 'stale_detected_at' exists")

	if !db.Migrator().HasColumn(&models.Show{}, "last_correction_result") {
		log.Fatal("Column 'last_correction_result' not found")
	}
	fmt.Println("✓ Column 'last_correction_result' exists")

	// Verify indexes
	fmt.Println("\nVerifying indexes...")
	if db.Migrator().HasIndex(&models.Show{}, "idx_shows_stale_detected_at") {
		fmt.Println("✓ Index 'idx_shows_stale_detected_at' exists")
	} else {
		fmt.Println("ℹ Index 'idx_shows_stale_detected_at' will be created by GORM")
	}

	fmt.Println("\n✅ Migration 006 completed successfully!")
	fmt.Println("Correction fields have been added to the shows table:")
	fmt.Println("  - refresh_threshold: Custom refresh threshold in days")
	fmt.Println("  - stale_detected_at: Timestamp when show was detected as stale")
	fmt.Println("  - last_correction_result: Result of last correction attempt")
}
