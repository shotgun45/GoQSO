package main

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestContactJSONMarshaling(t *testing.T) {
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

	// Test JSON marshaling
	data, err := json.Marshal(contact)
	if err != nil {
		t.Fatalf("Failed to marshal contact: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaledContact Contact
	err = json.Unmarshal(data, &unmarshaledContact)
	if err != nil {
		t.Fatalf("Failed to unmarshal contact: %v", err)
	}

	// Verify key fields
	if unmarshaledContact.Callsign != contact.Callsign {
		t.Errorf("Callsign mismatch: got %s, want %s", unmarshaledContact.Callsign, contact.Callsign)
	}

	if unmarshaledContact.Frequency != contact.Frequency {
		t.Errorf("Frequency mismatch: got %.3f, want %.3f", unmarshaledContact.Frequency, contact.Frequency)
	}

	if unmarshaledContact.Mode != contact.Mode {
		t.Errorf("Mode mismatch: got %s, want %s", unmarshaledContact.Mode, contact.Mode)
	}
}

func TestQSOLoggerJSONOperations(t *testing.T) {
	logger := &QSOLogger{
		Contacts: []Contact{
			{
				Callsign:  "TEST1",
				Date:      time.Now(),
				Frequency: 14.205,
				Band:      "20m",
				Mode:      "SSB",
			},
			{
				Callsign:  "TEST2",
				Date:      time.Now(),
				Frequency: 7.125,
				Band:      "40m",
				Mode:      "CW",
			},
		},
	}

	// Test JSON marshaling of logger
	data, err := json.Marshal(logger)
	if err != nil {
		t.Fatalf("Failed to marshal logger: %v", err)
	}

	// Test JSON unmarshaling of logger
	var unmarshaledLogger QSOLogger
	err = json.Unmarshal(data, &unmarshaledLogger)
	if err != nil {
		t.Fatalf("Failed to unmarshal logger: %v", err)
	}

	if len(unmarshaledLogger.Contacts) != 2 {
		t.Errorf("Expected 2 contacts, got %d", len(unmarshaledLogger.Contacts))
	}

	if unmarshaledLogger.Contacts[0].Callsign != "TEST1" {
		t.Errorf("First contact callsign mismatch: got %s, want TEST1", unmarshaledLogger.Contacts[0].Callsign)
	}

	if unmarshaledLogger.Contacts[1].Callsign != "TEST2" {
		t.Errorf("Second contact callsign mismatch: got %s, want TEST2", unmarshaledLogger.Contacts[1].Callsign)
	}
}

func TestSaveAndLoadContacts(t *testing.T) {
	// Create a temporary test file
	testFile := "test_qso_log.json"
	defer os.Remove(testFile) // Clean up after test

	// Create a logger with test data
	originalLogger := &QSOLogger{
		Contacts: []Contact{
			{
				Callsign:  "TESTQSO",
				Date:      time.Date(2025, 9, 20, 12, 0, 0, 0, time.UTC),
				Frequency: 21.205,
				Band:      "15m",
				Mode:      "FT8",
			},
		},
	}

	// Save to test file
	data, err := json.MarshalIndent(originalLogger, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	err = os.WriteFile(testFile, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load from test file
	loadedData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	var loadedLogger QSOLogger
	err = json.Unmarshal(loadedData, &loadedLogger)
	if err != nil {
		t.Fatalf("Failed to unmarshal loaded data: %v", err)
	}

	// Verify loaded data
	if len(loadedLogger.Contacts) != 1 {
		t.Errorf("Expected 1 contact, got %d", len(loadedLogger.Contacts))
	}

	if loadedLogger.Contacts[0].Callsign != "TESTQSO" {
		t.Errorf("Callsign mismatch: got %s, want TESTQSO", loadedLogger.Contacts[0].Callsign)
	}

	if loadedLogger.Contacts[0].Band != "15m" {
		t.Errorf("Band mismatch: got %s, want 15m", loadedLogger.Contacts[0].Band)
	}
}
