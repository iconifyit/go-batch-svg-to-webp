package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseService struct {
	DB *gorm.DB
}

// Config represents the database configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string // Options: "disable", "require", etc.
}

func init() {
	// Load the .env file during package initialization
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// NewDatabaseService initializes and returns a new DatabaseService instance
func NewDatabaseService() (*DatabaseService, error) {
	// Load environment variables (use godotenv if needed)
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASS"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	fmt.Println(dsn)

	// Configure Gorm with logger
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Adjust log level as needed
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %v", err)
	}

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from Gorm: %v", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(10)                  // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                   // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(30 * time.Minute) // Maximum lifetime of a connection

	return &DatabaseService{DB: db}, nil
}

// Close closes the database connection
func (svc *DatabaseService) Close() error {
	sqlDB, err := svc.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from Gorm: %v", err)
	}
	return sqlDB.Close()
}

func WhereNot(column string, value interface{}) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Not(map[string]interface{}{column: value})
	}
}

func WhereIn(column string, values []interface{}) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(column+" IN ?", values)
	}
}

func WhereLike(column string, pattern string) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(column+" LIKE ?", pattern)
	}
}

func WhereCustom(customCondition string, args ...interface{}) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(customCondition, args...)
	}
}

func Where(column string, value interface{}) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		return tx.Where(column+" = ?", value)
	}
}
