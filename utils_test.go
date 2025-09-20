package main

import (
	"testing"
	"time"
)

func TestFrequencyToBand(t *testing.T) {
	tests := []struct {
		name      string
		frequency float64
		expected  string
	}{
		{"160m band", 1.9, "160m"},
		{"80m band", 3.6, "80m"},
		{"60m band", 5.35, "60m"},
		{"40m band", 7.1, "40m"},
		{"30m band", 10.12, "30m"},
		{"20m band", 14.2, "20m"},
		{"17m band", 18.1, "17m"},
		{"15m band", 21.2, "15m"},
		{"12m band", 24.9, "12m"},
		{"10m band", 28.5, "10m"},
		{"6m band", 52.0, "6m"},
		{"2m band", 146.0, "2m"},
		{"70cm band", 435.0, "70cm"},
		{"Unknown frequency", 123.45, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := frequencyToBand(tt.frequency)
			if result != tt.expected {
				t.Errorf("frequencyToBand(%.2f) = %s; want %s", tt.frequency, result, tt.expected)
			}
		})
	}
}

func TestQSOLoggerCreation(t *testing.T) {
	// Test creating a QSOLogger struct
	logger := &QSOLogger{
		Contacts: make([]Contact, 0),
	}

	if logger.Contacts == nil {
		t.Error("Contacts slice should be initialized")
	}

	if len(logger.Contacts) != 0 {
		t.Errorf("Expected empty contacts slice, got length %d", len(logger.Contacts))
	}
}

func TestContactStructure(t *testing.T) {
	// Test creating a Contact with all fields
	testTime := time.Date(2025, 9, 20, 14, 30, 0, 0, time.UTC)
	contact := Contact{
		Callsign:    "W1AW",
		Date:        testTime,
		TimeOn:      "1430",
		TimeOff:     "1445",
		Frequency:   14.205,
		Band:        "20m",
		Mode:        "SSB",
		RSTSent:     "599",
		RSTReceived: "589",
		Name:        "John",
		QTH:         "Newington, CT",
		Country:     "USA",
		Grid:        "FN31",
		Power:       100,
		Comment:     "Test contact",
		Confirmed:   false,
	}

	// Test all field assignments
	if contact.Callsign != "W1AW" {
		t.Errorf("Expected callsign W1AW, got %s", contact.Callsign)
	}

	if contact.Band != "20m" {
		t.Errorf("Expected band 20m, got %s", contact.Band)
	}

	if contact.Mode != "SSB" {
		t.Errorf("Expected mode SSB, got %s", contact.Mode)
	}

	if contact.Power != 100 {
		t.Errorf("Expected power 100, got %d", contact.Power)
	}

	if contact.Frequency != 14.205 {
		t.Errorf("Expected frequency 14.205, got %.3f", contact.Frequency)
	}

	if contact.RSTSent != "599" {
		t.Errorf("Expected RST sent 599, got %s", contact.RSTSent)
	}

	if contact.RSTReceived != "589" {
		t.Errorf("Expected RST received 589, got %s", contact.RSTReceived)
	}

	if contact.Grid != "FN31" {
		t.Errorf("Expected grid FN31, got %s", contact.Grid)
	}

	if contact.Name != "John" {
		t.Errorf("Expected name John, got %s", contact.Name)
	}

	if contact.QTH != "Newington, CT" {
		t.Errorf("Expected QTH 'Newington, CT', got %s", contact.QTH)
	}

	if contact.Country != "USA" {
		t.Errorf("Expected country USA, got %s", contact.Country)
	}

	if contact.Comment != "Test contact" {
		t.Errorf("Expected comment 'Test contact', got %s", contact.Comment)
	}

	if contact.TimeOn != "1430" {
		t.Errorf("Expected time on 1430, got %s", contact.TimeOn)
	}

	if contact.TimeOff != "1445" {
		t.Errorf("Expected time off 1445, got %s", contact.TimeOff)
	}

	if contact.Confirmed != false {
		t.Errorf("Expected confirmed false, got %t", contact.Confirmed)
	}
}

func TestQSOLoggerOperations(t *testing.T) {
	logger := &QSOLogger{
		Contacts: make([]Contact, 0),
	}

	// Test empty logger
	if len(logger.Contacts) != 0 {
		t.Errorf("Expected 0 contacts, got %d", len(logger.Contacts))
	}

	// Add a test contact
	testContact := Contact{
		Callsign:    "TEST",
		Date:        time.Date(2025, 9, 20, 15, 0, 0, 0, time.UTC),
		Frequency:   14.205,
		Band:        "20m",
		Mode:        "SSB",
		RSTSent:     "599",
		RSTReceived: "579",
	}

	logger.Contacts = append(logger.Contacts, testContact)

	// Test after adding contact
	if len(logger.Contacts) != 1 {
		t.Errorf("Expected 1 contact after adding, got %d", len(logger.Contacts))
	}

	if logger.Contacts[0].Callsign != "TEST" {
		t.Errorf("Expected callsign TEST, got %s", logger.Contacts[0].Callsign)
	}

	if logger.Contacts[0].Mode != "SSB" {
		t.Errorf("Expected mode SSB, got %s", logger.Contacts[0].Mode)
	}

	// Add another contact
	secondContact := Contact{
		Callsign:  "TEST2",
		Date:      time.Date(2025, 9, 20, 16, 0, 0, 0, time.UTC),
		Frequency: 7.125,
		Band:      "40m",
		Mode:      "CW",
	}

	logger.Contacts = append(logger.Contacts, secondContact)

	// Test multiple contacts
	if len(logger.Contacts) != 2 {
		t.Errorf("Expected 2 contacts after adding second, got %d", len(logger.Contacts))
	}

	// Verify both contacts are present
	callsigns := make(map[string]bool)
	for _, contact := range logger.Contacts {
		callsigns[contact.Callsign] = true
	}

	if !callsigns["TEST"] {
		t.Error("First contact (TEST) not found in logger")
	}

	if !callsigns["TEST2"] {
		t.Error("Second contact (TEST2) not found in logger")
	}
}

// TestFrequencyToBandEdgeCases tests edge cases for frequency to band conversion
func TestFrequencyToBandEdgeCases(t *testing.T) {
	edgeCases := []struct {
		name      string
		frequency float64
		expected  string
	}{
		// Test exact boundaries
		{"160m lower bound", 1.8, "160m"},
		{"160m upper bound", 2.0, "160m"},
		{"80m lower bound", 3.5, "80m"},
		{"80m upper bound", 4.0, "80m"},
		{"Below amateur bands", 1.0, "Unknown"},
		{"Between 160m and 80m", 2.5, "Unknown"},
		{"Very high frequency", 1000.0, "Unknown"},
		{"Zero frequency", 0.0, "Unknown"},
		{"Negative frequency", -1.0, "Unknown"},
	}

	for _, tt := range edgeCases {
		t.Run(tt.name, func(t *testing.T) {
			result := frequencyToBand(tt.frequency)
			if result != tt.expected {
				t.Errorf("frequencyToBand(%.2f) = %s; want %s", tt.frequency, result, tt.expected)
			}
		})
	}
}

// TestContactValidation tests various contact field validations
func TestContactValidation(t *testing.T) {
	tests := []struct {
		name     string
		contact  Contact
		field    string
		expected interface{}
	}{
		{
			name: "Empty callsign",
			contact: Contact{
				Callsign: "",
				Mode:     "SSB",
			},
			field:    "Callsign",
			expected: "",
		},
		{
			name: "Valid callsign with suffix",
			contact: Contact{
				Callsign: "W1AW/P",
				Mode:     "CW",
			},
			field:    "Callsign",
			expected: "W1AW/P",
		},
		{
			name: "Power zero",
			contact: Contact{
				Power: 0,
			},
			field:    "Power",
			expected: 0,
		},
		{
			name: "High power",
			contact: Contact{
				Power: 1500,
			},
			field:    "Power",
			expected: 1500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.field {
			case "Callsign":
				if tt.contact.Callsign != tt.expected.(string) {
					t.Errorf("Expected %s, got %s", tt.expected.(string), tt.contact.Callsign)
				}
			case "Power":
				if tt.contact.Power != tt.expected.(int) {
					t.Errorf("Expected %d, got %d", tt.expected.(int), tt.contact.Power)
				}
			}
		})
	}
}
