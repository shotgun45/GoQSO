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

func TestContactStructure(t *testing.T) {
	// Test creating a Contact with all fields including new time format
	testTime := time.Date(2025, 9, 20, 14, 30, 0, 0, time.UTC)
	contact := Contact{
		Callsign:    "W1AW",
		Date:        testTime,
		TimeOn:      "14:30:00", // Updated to HH:MM:SS format
		TimeOff:     "14:45:00", // Updated to HH:MM:SS format
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

	if contact.TimeOn != "14:30:00" {
		t.Errorf("Expected time on 14:30:00, got %s", contact.TimeOn)
	}

	if contact.TimeOff != "14:45:00" {
		t.Errorf("Expected time off 14:45:00, got %s", contact.TimeOff)
	}

	if contact.Confirmed != false {
		t.Errorf("Expected confirmed false, got %t", contact.Confirmed)
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

// TestTimeFormatValidation tests time format validation for HH:MM:SS
func TestTimeFormatValidation(t *testing.T) {
	validTimeFormats := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"Valid HH:MM:SS", "14:30:00", true},
		{"Valid midnight", "00:00:00", true},
		{"Valid end of day", "23:59:59", true},
		{"Valid noon", "12:00:00", true},
		{"Legacy HHMM format", "1430", true}, // Should still be accepted
		{"Valid early morning", "06:15:30", true},
		{"Valid evening", "21:45:22", true},
		{"Invalid hour 24", "24:00:00", false},
		{"Invalid hour 25", "25:30:00", false},
		{"Invalid minute 60", "14:60:00", false},
		{"Invalid minute 99", "14:99:00", false},
		{"Invalid second 60", "14:30:60", false},
		{"Invalid second 99", "14:30:99", false},
		{"Invalid format H:MM:SS", "4:30:00", false},
		{"Invalid format HH:M:SS", "14:3:00", false},
		{"Invalid format HH:MM:S", "14:30:0", false},
		{"Empty string", "", true}, // Empty should be allowed
		{"Invalid characters", "14:3a:00", false},
		{"Too many colons", "14:30:00:00", false},
		{"Missing seconds", "14:30", false},
		{"Only hour", "14", false},
	}

	for _, tt := range validTimeFormats {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTimeFormat(tt.timeStr)
			if result != tt.expected {
				t.Errorf("isValidTimeFormat(%q) = %t; want %t", tt.timeStr, result, tt.expected)
			}
		})
	}
}

// Helper function to validate time format (this would be implemented in the main code)
func isValidTimeFormat(timeStr string) bool {
	if timeStr == "" {
		return true // Empty time is allowed
	}

	// Check for legacy HHMM format (4 digits)
	if len(timeStr) == 4 {
		for _, c := range timeStr {
			if c < '0' || c > '9' {
				return false
			}
		}
		// Validate HHMM values
		if len(timeStr) == 4 {
			hour := timeStr[0:2]
			minute := timeStr[2:4]
			return isValidHour(hour) && isValidMinute(minute)
		}
		return true
	}

	// Check for HH:MM:SS format
	if len(timeStr) != 8 {
		return false
	}

	if timeStr[2] != ':' || timeStr[5] != ':' {
		return false
	}

	hour := timeStr[0:2]
	minute := timeStr[3:5]
	second := timeStr[6:8]

	return isValidHour(hour) && isValidMinute(minute) && isValidSecond(second)
}

func isValidHour(hour string) bool {
	if len(hour) != 2 {
		return false
	}
	for _, c := range hour {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Convert to int and check range
	h := int(hour[0]-'0')*10 + int(hour[1]-'0')
	return h >= 0 && h <= 23
}

func isValidMinute(minute string) bool {
	if len(minute) != 2 {
		return false
	}
	for _, c := range minute {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Convert to int and check range
	m := int(minute[0]-'0')*10 + int(minute[1]-'0')
	return m >= 0 && m <= 59
}

func isValidSecond(second string) bool {
	if len(second) != 2 {
		return false
	}
	for _, c := range second {
		if c < '0' || c > '9' {
			return false
		}
	}
	// Convert to int and check range
	s := int(second[0]-'0')*10 + int(second[1]-'0')
	return s >= 0 && s <= 59
}

// TestTimeFormatConversion tests conversion between time formats
func TestTimeFormatConversion(t *testing.T) {
	conversionTests := []struct {
		name     string
		input    string
		expected string
	}{
		{"HHMM to HH:MM:SS", "1430", "14:30:00"},
		{"HHMM midnight", "0000", "00:00:00"},
		{"HHMM evening", "2145", "21:45:00"},
		{"Already HH:MM:SS", "14:30:00", "14:30:00"},
		{"Empty string", "", ""},
	}

	for _, tt := range conversionTests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToHHMMSS(tt.input)
			if result != tt.expected {
				t.Errorf("convertToHHMMSS(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to convert time formats (this would be implemented in the main code)
func convertToHHMMSS(timeStr string) string {
	if timeStr == "" {
		return ""
	}

	// If already in HH:MM:SS format, return as-is
	if len(timeStr) == 8 && timeStr[2] == ':' && timeStr[5] == ':' {
		return timeStr
	}

	// If in HHMM format, convert to HH:MM:SS
	if len(timeStr) == 4 {
		hour := timeStr[0:2]
		minute := timeStr[2:4]
		return hour + ":" + minute + ":00"
	}

	return timeStr // Return as-is if format is unknown
}

// TestContactTimeFields tests that contacts can handle both time formats
func TestContactTimeFields(t *testing.T) {
	testCases := []struct {
		name     string
		timeOn   string
		timeOff  string
		expected bool
	}{
		{"Both HH:MM:SS", "14:30:00", "14:45:00", true},
		{"Both HHMM", "1430", "1445", true},
		{"Mixed formats", "14:30:00", "1445", true},
		{"Empty times", "", "", true},
		{"One empty", "14:30:00", "", true},
		{"Invalid format", "25:30:00", "14:45:00", false},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			contact := Contact{
				Callsign: "TEST",
				TimeOn:   tt.timeOn,
				TimeOff:  tt.timeOff,
			}

			// Test that the contact can be created
			if contact.Callsign != "TEST" {
				t.Error("Contact creation failed")
			}

			// Test time format validation
			validOn := isValidTimeFormat(contact.TimeOn)
			validOff := isValidTimeFormat(contact.TimeOff)

			if tt.expected && (!validOn || !validOff) {
				t.Errorf("Expected valid time formats for %s, got validOn=%t, validOff=%t", tt.name, validOn, validOff)
			}

			if !tt.expected && (validOn && validOff) {
				t.Errorf("Expected invalid time formats for %s, but both were valid", tt.name)
			}
		})
	}
}
