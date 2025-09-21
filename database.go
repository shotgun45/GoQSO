package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed sql/schema/*.sql
var embedMigrations embed.FS

// DatabaseConfig holds PostgreSQL connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabaseConfig creates database configuration from environment variables
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		User:     getEnvOrDefault("POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", ""),
		DBName:   getEnvOrDefault("POSTGRES_DB", "goqso"),
		SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ConnectionString builds PostgreSQL connection string
func (cfg *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

// InitializeDatabase creates the database connection and sets up tables
func InitializeDatabase() (*sql.DB, error) {
	config := NewDatabaseConfig()

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations uses Goose to run database migrations
func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "sql/schema"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Database migrations completed successfully")
	return nil
}

// CheckDatabaseConnection tests if the database is accessible
func CheckDatabaseConnection() error {
	config := NewDatabaseConfig()

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Database connection successful")
	return nil
}

// MigrateUp runs all pending migrations
func MigrateUp() error {
	config := NewDatabaseConfig()

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return runMigrations(db)
}

// MigrateDown rolls back the last migration
func MigrateDown() error {
	config := NewDatabaseConfig()

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Down(db, "sql/schema"); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	fmt.Println("Migration rolled back successfully")
	return nil
}

// MigrateStatus shows the status of all migrations
func MigrateStatus() error {
	config := NewDatabaseConfig()

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Status(db, "sql/schema"); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// Helper functions for import operations

// findExistingContact searches for an existing contact by callsign, date, and time
func findExistingContact(logger *QSOLogger, callsign, date, timeOn string) (*Contact, error) {
	query := `
		SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode,
			   rst_sent, rst_received, operator_name, qth, country, grid_square,
			   power_watts, comment, confirmed, created_at, updated_at
		FROM contacts
		WHERE callsign = $1 AND contact_date = $2 AND time_on = $3
		LIMIT 1
	`

	row := logger.db.QueryRow(query, callsign, date, timeOn)

	var contact Contact
	err := row.Scan(
		&contact.ID, &contact.Callsign, &contact.Date, &contact.TimeOn, &contact.TimeOff,
		&contact.Frequency, &contact.Band, &contact.Mode, &contact.RSTSent, &contact.RSTReceived,
		&contact.Name, &contact.QTH, &contact.Country, &contact.Grid, &contact.Power,
		&contact.Comment, &contact.Confirmed, &contact.CreatedAt, &contact.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No existing contact found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find existing contact: %w", err)
	}

	return &contact, nil
}

// createContact creates a new contact from a ContactRequest
func createContact(logger *QSOLogger, contactReq ContactRequest) (*Contact, error) {
	query := `
		INSERT INTO contacts (
			callsign, contact_date, time_on, time_off, frequency, band, mode,
			rst_sent, rst_received, operator_name, qth, country, grid_square,
			power_watts, comment, confirmed
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at, updated_at
	`

	// Parse date string to time.Time
	contactDate, err := time.Parse("2006-01-02", contactReq.ContactDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	var contact Contact
	contact.Callsign = contactReq.Callsign
	contact.Date = contactDate
	contact.TimeOn = contactReq.TimeOn
	contact.TimeOff = contactReq.TimeOff
	contact.Frequency = contactReq.Frequency
	contact.Band = contactReq.Band
	contact.Mode = contactReq.Mode
	contact.RSTSent = contactReq.RSTSent
	contact.RSTReceived = contactReq.RSTReceived
	contact.Name = contactReq.OperatorName
	contact.QTH = contactReq.QTH
	contact.Country = contactReq.Country
	contact.Grid = contactReq.GridSquare
	contact.Power = contactReq.PowerWatts
	contact.Comment = contactReq.Comment
	contact.Confirmed = contactReq.Confirmed

	err = logger.db.QueryRow(
		query,
		contact.Callsign, contact.Date, contact.TimeOn, contact.TimeOff,
		contact.Frequency, contact.Band, contact.Mode, contact.RSTSent, contact.RSTReceived,
		contact.Name, contact.QTH, contact.Country, contact.Grid, contact.Power,
		contact.Comment, contact.Confirmed,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	return &contact, nil
}

// updateContact updates an existing contact with new data
func updateContact(logger *QSOLogger, id int, contactReq ContactRequest) error {
	query := `
		UPDATE contacts SET
			callsign = $1, contact_date = $2, time_on = $3, time_off = $4,
			frequency = $5, band = $6, mode = $7, rst_sent = $8, rst_received = $9,
			operator_name = $10, qth = $11, country = $12, grid_square = $13,
			power_watts = $14, comment = $15, confirmed = $16, updated_at = NOW()
		WHERE id = $17
	`

	_, err := logger.db.Exec(
		query,
		contactReq.Callsign, contactReq.ContactDate, contactReq.TimeOn, contactReq.TimeOff,
		contactReq.Frequency, contactReq.Band, contactReq.Mode, contactReq.RSTSent, contactReq.RSTReceived,
		contactReq.OperatorName, contactReq.QTH, contactReq.Country, contactReq.GridSquare, contactReq.PowerWatts,
		contactReq.Comment, contactReq.Confirmed, id,
	)

	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	return nil
}
