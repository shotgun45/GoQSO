package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use test environment variables or defaults
	config := &DatabaseConfig{
		Host:     getEnvOrDefault("TEST_POSTGRES_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_POSTGRES_PORT", "5432"),
		User:     getEnvOrDefault("TEST_POSTGRES_USER", "postgres"),
		Password: getEnvOrDefault("TEST_POSTGRES_PASSWORD", ""),
		DBName:   getEnvOrDefault("TEST_POSTGRES_DB", "goqso_test"),
		SSLMode:  getEnvOrDefault("TEST_POSTGRES_SSLMODE", "disable"),
	}

	db, err := sql.Open("postgres", config.ConnectionString())
	if err != nil {
		t.Skipf("Skipping database tests - PostgreSQL not available: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping database tests - cannot connect to PostgreSQL: %v", err)
	}

	// Clean up any existing test data
	cleanupTestDB(t, db)

	// Run migrations
	if err := runMigrations(db); err != nil {
		t.Fatalf("Failed to run test migrations: %v", err)
	}

	return db
}

// cleanupTestDB removes all test data
func cleanupTestDB(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS contacts CASCADE")
	if err != nil {
		t.Logf("Warning: failed to clean up test database: %v", err)
	}
}

// teardownTestDB closes database connection and cleans up
func teardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		cleanupTestDB(t, db)
		db.Close()
	}
}

func TestContactDatabaseOperations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Test data
	contact := Contact{
		Callsign:    "W1AW",
		Date:        time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC),
		TimeOn:      "1200",
		TimeOff:     "1210",
		Frequency:   14.205,
		Band:        "20m",
		Mode:        "SSB",
		RSTSent:     "599",
		RSTReceived: "589",
		Name:        "John",
		QTH:         "Connecticut",
		Country:     "USA",
		Grid:        "FN31",
		Power:       100,
		Comment:     "Test QSO",
		Confirmed:   false,
	}

	// Test saving contact
	err := logger.SaveContact(&contact)
	if err != nil {
		t.Fatalf("Failed to save contact: %v", err)
	}

	// Verify contact was assigned an ID
	if contact.ID == 0 {
		t.Error("Contact ID should be assigned after saving")
	}

	// Test loading contacts
	contacts, err := logger.LoadContacts()
	if err != nil {
		t.Fatalf("Failed to load contacts: %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}

	loadedContact := contacts[0]

	// Verify key fields
	if loadedContact.Callsign != contact.Callsign {
		t.Errorf("Callsign mismatch: got %s, want %s", loadedContact.Callsign, contact.Callsign)
	}

	if loadedContact.Frequency != contact.Frequency {
		t.Errorf("Frequency mismatch: got %.3f, want %.3f", loadedContact.Frequency, contact.Frequency)
	}

	if loadedContact.Mode != contact.Mode {
		t.Errorf("Mode mismatch: got %s, want %s", loadedContact.Mode, contact.Mode)
	}

	if loadedContact.Band != contact.Band {
		t.Errorf("Band mismatch: got %s, want %s", loadedContact.Band, contact.Band)
	}

	if loadedContact.Name != contact.Name {
		t.Errorf("Name mismatch: got %s, want %s", loadedContact.Name, contact.Name)
	}

	if loadedContact.Country != contact.Country {
		t.Errorf("Country mismatch: got %s, want %s", loadedContact.Country, contact.Country)
	}
}

func TestMultipleContactsDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Create multiple test contacts
	contacts := []Contact{
		{
			Callsign:  "TEST1",
			Date:      time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC),
			Frequency: 14.205,
			Band:      "20m",
			Mode:      "SSB",
			Country:   "USA",
		},
		{
			Callsign:  "TEST2",
			Date:      time.Date(2025, 9, 20, 13, 0, 0, 0, time.UTC),
			Frequency: 7.125,
			Band:      "40m",
			Mode:      "CW",
			Country:   "Canada",
		},
		{
			Callsign:  "TEST3",
			Date:      time.Date(2025, 9, 20, 14, 0, 0, 0, time.UTC),
			Frequency: 21.205,
			Band:      "15m",
			Mode:      "FT8",
			Country:   "Germany",
		},
	}

	// Save all contacts
	for i := range contacts {
		err := logger.SaveContact(&contacts[i])
		if err != nil {
			t.Fatalf("Failed to save contact %d: %v", i, err)
		}
	}

	// Load all contacts
	loadedContacts, err := logger.LoadContacts()
	if err != nil {
		t.Fatalf("Failed to load contacts: %v", err)
	}

	if len(loadedContacts) != 3 {
		t.Errorf("Expected 3 contacts, got %d", len(loadedContacts))
	}

	// Verify contacts are sorted by date DESC (newest first)
	if len(loadedContacts) >= 2 {
		if loadedContacts[0].Date.Before(loadedContacts[1].Date) {
			t.Error("Contacts should be sorted by date DESC")
		}
	}

	// Verify specific contact data
	callsigns := make(map[string]bool)
	for _, contact := range loadedContacts {
		callsigns[contact.Callsign] = true
	}

	expectedCallsigns := []string{"TEST1", "TEST2", "TEST3"}
	for _, callsign := range expectedCallsigns {
		if !callsigns[callsign] {
			t.Errorf("Missing expected callsign: %s", callsign)
		}
	}
}

func TestDatabaseConnection(t *testing.T) {
	// Test database configuration
	config := NewDatabaseConfig()

	// Override with test values if provided
	if os.Getenv("TEST_POSTGRES_HOST") != "" {
		config.Host = os.Getenv("TEST_POSTGRES_HOST")
	}
	if os.Getenv("TEST_POSTGRES_DB") != "" {
		config.DBName = os.Getenv("TEST_POSTGRES_DB")
	}

	// Test connection string generation
	connStr := config.ConnectionString()
	if connStr == "" {
		t.Error("Connection string should not be empty")
	}

	// Test if connection string contains expected elements
	expectedElements := []string{
		fmt.Sprintf("host=%s", config.Host),
		fmt.Sprintf("port=%s", config.Port),
		fmt.Sprintf("dbname=%s", config.DBName),
	}

	for _, element := range expectedElements {
		if !contains(connStr, element) {
			t.Errorf("Connection string missing element: %s", element)
		}
	}
}

func TestDatabaseMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Verify that contacts table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'contacts'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}

	if !exists {
		t.Error("Contacts table should exist after migrations")
	}

	// Verify that required indexes exist
	expectedIndexes := []string{
		"idx_contacts_callsign",
		"idx_contacts_date",
		"idx_contacts_band",
		"idx_contacts_mode",
	}

	for _, indexName := range expectedIndexes {
		var indexExists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM pg_indexes 
				WHERE tablename = 'contacts' 
				AND indexname = $1
			)
		`, indexName).Scan(&indexExists)

		if err != nil {
			t.Fatalf("Failed to check index %s: %v", indexName, err)
		}

		if !indexExists {
			t.Errorf("Index %s should exist after migrations", indexName)
		}
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
