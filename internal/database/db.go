package database

// Database connection and initialization will be here
// This will handle GORM setup and migrations

import (
	"deploy-platform/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection and runs migrations
// If databaseURL is empty, uses SQLite for development
// Otherwise, uses PostgreSQL (format: "postgres://user:password@host/dbname?sslmode=disable")
func InitDB(databaseURL string) error {
	var err error
	var dialector gorm.Dialector

	// Use SQLite for development, PostgreSQL for production
	if databaseURL == "" {
		databaseURL = "deployments.db"
		dialector = sqlite.Open(databaseURL)
		log.Println("Using SQLite database:", databaseURL)
	} else {
		dialector = postgres.Open(databaseURL)
		log.Println("Using PostgreSQL database")
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return err
	}

	// Auto-migrate all models
	// This will create tables, add missing columns, and create indexes
	err = DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Deployment{},
		&models.Build{},
		&models.Environment{},
		&models.Hostname{},
	)

	if err != nil {
		return err
	}

	log.Println("Database connected and migrated successfully")
	return nil
}
