package database

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
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
	// Load the .env file during package initialization. A missing .env is not
	// fatal: configuration may be provided directly via environment variables
	// (e.g. in CI or production). Other load failures (parse or permission
	// errors) are surfaced as warnings so a broken .env is not mistaken for
	// an absent one.
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			log.Printf("No .env file found; using environment variables")
		} else {
			log.Printf("Warning: failed to load .env file: %v; using environment variables", err)
		}
	}
}

// buildDSN assembles the PostgreSQL connection URL from the POSTGRES_* env
// vars. Values are whitespace-trimmed (stray spaces in .env files are
// common and invalid in URLs) and credentials use userinfo escaping via
// url.UserPassword, so passwords containing special characters (+, @, /,
// spaces) are transmitted correctly.
func buildDSN() string {
	env := func(key string) string {
		return strings.TrimSpace(os.Getenv(key))
	}
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(env("POSTGRES_USER"), env("POSTGRES_PASS")),
		Host:     env("POSTGRES_HOST") + ":" + env("POSTGRES_PORT"),
		Path:     env("POSTGRES_DB"),
		RawQuery: "sslmode=disable",
	}
	return dsn.String()
}

// redactCredentials removes the database password from an error message.
// Driver parse errors echo the full DSN, so surfacing them verbatim would
// leak credentials into logs. Both the raw and userinfo-escaped forms are
// replaced.
func redactCredentials(err error) string {
	msg := err.Error()
	pass := strings.TrimSpace(os.Getenv("POSTGRES_PASS"))
	if pass == "" {
		return msg
	}
	escaped := strings.TrimPrefix(url.UserPassword("u", pass).String(), "u:")
	for _, needle := range []string{escaped, pass} {
		msg = strings.ReplaceAll(msg, needle, "[REDACTED]")
	}
	return msg
}

// NewDatabaseService initializes and returns a new DatabaseService instance
func NewDatabaseService() (*DatabaseService, error) {
	dsn := buildDSN()

	// Never log the DSN itself - it contains the database password. Log only
	// the non-secret connection coordinates for troubleshooting.
	log.Printf("Connecting to PostgreSQL host=%s dbname=%s port=%s",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PORT"))

	// Configure Gorm with logger
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Adjust log level as needed
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %s", redactCredentials(err))
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
