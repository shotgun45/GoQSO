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
		TimeOn:      "12:00:00", // Updated to HH:MM:SS format
		TimeOff:     "12:10:00", // Updated to HH:MM:SS format
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

// Tests for API methods used by HTTP server

func TestGetAllContactsAPI(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Initially should return empty slice
	contacts, err := logger.GetAllContacts()
	if err != nil {
		t.Fatalf("GetAllContacts failed: %v", err)
	}

	if len(contacts) != 0 {
		t.Errorf("Expected 0 contacts, got %d", len(contacts))
	}

	// Add a test contact
	testContact := Contact{
		Callsign:    "W1AW",
		Date:        time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC),
		TimeOn:      "12:00:00",
		TimeOff:     "12:10:00",
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

	err = logger.SaveContact(&testContact)
	if err != nil {
		t.Fatalf("Failed to save test contact: %v", err)
	}

	// Test GetAllContacts returns the contact
	contacts, err = logger.GetAllContacts()
	if err != nil {
		t.Fatalf("GetAllContacts failed: %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}

	if contacts[0].Callsign != "W1AW" {
		t.Errorf("Expected callsign W1AW, got %s", contacts[0].Callsign)
	}
}

func TestAddContactStructAPI(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	testContact := Contact{
		Callsign:    "VE1TEST",
		Date:        time.Date(2025, 9, 20, 15, 30, 0, 0, time.UTC),
		TimeOn:      "15:30:00",
		TimeOff:     "15:45:00",
		Frequency:   7.125,
		Band:        "40m",
		Mode:        "CW",
		RSTSent:     "579",
		RSTReceived: "559",
		Name:        "Jane",
		QTH:         "Halifax, NS",
		Country:     "Canada",
		Grid:        "FN85",
		Power:       50,
		Comment:     "Great CW QSO",
		Confirmed:   true,
	}

	// Test AddContactStruct
	err := logger.AddContactStruct(testContact)
	if err != nil {
		t.Fatalf("AddContactStruct failed: %v", err)
	}

	// Verify contact was saved
	contacts, err := logger.GetAllContacts()
	if err != nil {
		t.Fatalf("GetAllContacts failed: %v", err)
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(contacts))
	}

	saved := contacts[0]
	if saved.Callsign != testContact.Callsign {
		t.Errorf("Expected callsign %s, got %s", testContact.Callsign, saved.Callsign)
	}

	if saved.Mode != testContact.Mode {
		t.Errorf("Expected mode %s, got %s", testContact.Mode, saved.Mode)
	}

	if saved.Country != testContact.Country {
		t.Errorf("Expected country %s, got %s", testContact.Country, saved.Country)
	}

	if saved.Confirmed != testContact.Confirmed {
		t.Errorf("Expected confirmed %t, got %t", testContact.Confirmed, saved.Confirmed)
	}
}

func TestGetContactByIDAPI(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Test getting non-existent contact
	contact, err := logger.GetContactByID(999)
	if err == nil {
		t.Error("Expected error for non-existent contact")
	}
	if contact != nil {
		t.Error("Expected nil contact for non-existent ID")
	}

	// Add a test contact
	testContact := Contact{
		Callsign:    "JA1TEST",
		Date:        time.Date(2025, 9, 20, 10, 0, 0, 0, time.UTC),
		TimeOn:      "10:00:00",
		TimeOff:     "10:05:00",
		Frequency:   21.205,
		Band:        "15m",
		Mode:        "FT8",
		RSTSent:     "-10",
		RSTReceived: "-15",
		Name:        "Taro",
		QTH:         "Tokyo",
		Country:     "Japan",
		Grid:        "PM95",
		Power:       10,
		Comment:     "FT8 digital mode",
		Confirmed:   false,
	}

	err = logger.SaveContact(&testContact)
	if err != nil {
		t.Fatalf("Failed to save test contact: %v", err)
	}

	// Test getting existing contact
	retrieved, err := logger.GetContactByID(testContact.ID)
	if err != nil {
		t.Fatalf("GetContactByID failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected contact, got nil")
	}

	if retrieved.Callsign != testContact.Callsign {
		t.Errorf("Expected callsign %s, got %s", testContact.Callsign, retrieved.Callsign)
	}

	if retrieved.Mode != testContact.Mode {
		t.Errorf("Expected mode %s, got %s", testContact.Mode, retrieved.Mode)
	}

	if retrieved.Grid != testContact.Grid {
		t.Errorf("Expected grid %s, got %s", testContact.Grid, retrieved.Grid)
	}

	if retrieved.ID != testContact.ID {
		t.Errorf("Expected ID %d, got %d", testContact.ID, retrieved.ID)
	}
}

func TestUpdateContactAPI(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Add a test contact to update
	originalContact := Contact{
		Callsign:    "G0TEST",
		Date:        time.Date(2025, 9, 20, 16, 0, 0, 0, time.UTC),
		TimeOn:      "16:00:00",
		TimeOff:     "16:15:00",
		Frequency:   28.405,
		Band:        "10m",
		Mode:        "SSB",
		RSTSent:     "599",
		RSTReceived: "599",
		Name:        "Robert",
		QTH:         "London",
		Country:     "England",
		Grid:        "IO91",
		Power:       100,
		Comment:     "Good signal",
		Confirmed:   false,
	}

	err := logger.SaveContact(&originalContact)
	if err != nil {
		t.Fatalf("Failed to save original contact: %v", err)
	}

	// Update the contact
	updatedContact := originalContact
	updatedContact.Name = "Bob"
	updatedContact.QTH = "Manchester"
	updatedContact.Power = 250
	updatedContact.Confirmed = true
	updatedContact.Comment = "Updated QSO details"
	updatedContact.UpdatedAt = time.Now()

	err = logger.UpdateContact(updatedContact)
	if err != nil {
		t.Fatalf("UpdateContact failed: %v", err)
	}

	// Retrieve and verify updates
	retrieved, err := logger.GetContactByID(originalContact.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated contact: %v", err)
	}

	if retrieved.Name != "Bob" {
		t.Errorf("Expected updated name Bob, got %s", retrieved.Name)
	}

	if retrieved.QTH != "Manchester" {
		t.Errorf("Expected updated QTH Manchester, got %s", retrieved.QTH)
	}

	if retrieved.Power != 250 {
		t.Errorf("Expected updated power 250, got %d", retrieved.Power)
	}

	if retrieved.Confirmed != true {
		t.Errorf("Expected confirmed true, got %t", retrieved.Confirmed)
	}

	if retrieved.Comment != "Updated QSO details" {
		t.Errorf("Expected updated comment, got %s", retrieved.Comment)
	}

	// Test updating non-existent contact
	nonExistentContact := Contact{
		ID:       999,
		Callsign: "FAKE",
	}

	err = logger.UpdateContact(nonExistentContact)
	if err == nil {
		t.Error("Expected error when updating non-existent contact")
	}
}

func TestDeleteContactAPI(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Test deleting non-existent contact
	err := logger.DeleteContact(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent contact")
	}

	// Add test contacts
	contacts := []Contact{
		{
			Callsign: "DELETE1",
			Date:     time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC),
			Mode:     "SSB",
			Band:     "20m",
		},
		{
			Callsign: "DELETE2",
			Date:     time.Date(2025, 9, 20, 13, 0, 0, 0, time.UTC),
			Mode:     "CW",
			Band:     "40m",
		},
	}

	for i := range contacts {
		err := logger.SaveContact(&contacts[i])
		if err != nil {
			t.Fatalf("Failed to save test contact %d: %v", i, err)
		}
	}

	// Verify we have 2 contacts
	allContacts, err := logger.GetAllContacts()
	if err != nil {
		t.Fatalf("GetAllContacts failed: %v", err)
	}
	if len(allContacts) != 2 {
		t.Errorf("Expected 2 contacts before deletion, got %d", len(allContacts))
	}

	// Delete first contact
	err = logger.DeleteContact(contacts[0].ID)
	if err != nil {
		t.Fatalf("DeleteContact failed: %v", err)
	}

	// Verify we now have 1 contact
	allContacts, err = logger.GetAllContacts()
	if err != nil {
		t.Fatalf("GetAllContacts failed: %v", err)
	}
	if len(allContacts) != 1 {
		t.Errorf("Expected 1 contact after deletion, got %d", len(allContacts))
	}

	// Verify the remaining contact is the right one
	if allContacts[0].Callsign != "DELETE2" {
		t.Errorf("Expected remaining contact DELETE2, got %s", allContacts[0].Callsign)
	}

	// Try to get the deleted contact
	deleted, err := logger.GetContactByID(contacts[0].ID)
	if err == nil {
		t.Error("Expected error when getting deleted contact")
	}
	if deleted != nil {
		t.Error("Expected nil when getting deleted contact")
	}
}
