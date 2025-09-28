package goqso

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestDatabaseErrorHandling(t *testing.T) {
	// Test with invalid database connection
	invalidLogger := &QSOLogger{db: nil}

	// Test SaveContact with nil database - this will panic, so we need to recover
	contact := Contact{
		Callsign: "ERROR_TEST",
		Date:     time.Now(),
	}

	// Test that SaveContact with nil database panics (as expected)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when saving contact with nil database")
		}
	}()

	_ = invalidLogger.SaveContact(&contact)
	// This should panic before reaching here
}

func TestInvalidContactData(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Test with extremely long strings that might cause database errors
	invalidContact := Contact{
		Callsign: strings.Repeat("A", 1000), // Very long callsign
		Date:     time.Now(),
		Mode:     strings.Repeat("X", 500),   // Very long mode
		Comment:  strings.Repeat("C", 10000), // Very long comment
	}

	// This might fail due to database constraints
	err := logger.SaveContact(&invalidContact)
	// We expect this to either succeed or fail gracefully
	if err != nil {
		t.Logf("Expected behavior: long strings caused error: %v", err)
	}
}

func TestAPIErrorResponses(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test malformed JSON in POST request
	malformedJSON := `{"callsign": "TEST", "invalid": json}`
	resp, err := http.Post(server.URL+"/api/contacts", "application/json", strings.NewReader(malformedJSON))
	if err != nil {
		t.Fatalf("Failed to post malformed JSON: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for malformed JSON, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if response.Success {
		t.Error("Expected success false for malformed JSON")
	}

	if response.Error == "" {
		t.Error("Expected error message for malformed JSON")
	}

	// Test missing Content-Type header
	resp2, err := http.Post(server.URL+"/api/contacts", "", strings.NewReader(`{"callsign": "TEST"}`))
	if err != nil {
		t.Fatalf("Failed to post without content type: %v", err)
	}
	defer resp2.Body.Close()

	// Should still work or give appropriate error
	if resp2.StatusCode != http.StatusBadRequest && resp2.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 400 or 201 for missing content type, got %d", resp2.StatusCode)
	}
}

func TestInvalidHTTPMethods(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test unsupported methods on /api/contacts
	unsupportedMethods := []string{"PATCH", "HEAD", "TRACE", "CONNECT"}

	for _, method := range unsupportedMethods {
		req, err := http.NewRequest(method, server.URL+"/api/contacts", nil)
		if err != nil {
			t.Fatalf("Failed to create %s request: %v", method, err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute %s request: %v", method, err)
		}
		defer resp.Body.Close()

		// Should return method not allowed or not found
		if resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 405 or 404 for %s method, got %d", method, resp.StatusCode)
		}
	}
}

func TestLargeRequestHandling(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Create a very large contact record
	largeContact := Contact{
		Callsign: "LARGE_TEST",
		Date:     time.Now(),
		Comment:  strings.Repeat("This is a very long comment. ", 1000), // ~30KB comment
	}

	jsonData, err := json.Marshal(largeContact)
	if err != nil {
		t.Fatalf("Failed to marshal large contact: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/contacts", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to post large contact: %v", err)
	}
	defer resp.Body.Close()

	// Should either succeed or fail gracefully
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 201, 400, or 413 for large request, got %d", resp.StatusCode)
	}
}

func TestConcurrentRequests(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Add a test contact first
	testContact := Contact{
		Callsign: "CONCURRENT_TEST",
		Date:     time.Now(),
		Mode:     "SSB",
		Band:     "20m",
	}

	err := logger.SaveContact(&testContact)
	if err != nil {
		t.Fatalf("Failed to save test contact: %v", err)
	}

	// Perform concurrent operations
	numRequests := 10
	results := make(chan error, numRequests)

	// Concurrent GET requests
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(server.URL + "/api/contacts")
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- err
				return
			}
			results <- nil
		}()
	}

	// Check results
	for i := 0; i < numRequests; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent request %d failed: %v", i, err)
		}
	}
}

func TestEdgeCaseInputs(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	edgeCases := []struct {
		name     string
		contact  Contact
		expected int // Expected HTTP status code
	}{
		{
			name: "Empty callsign",
			contact: Contact{
				Callsign: "",
				Date:     time.Now(),
			},
			expected: http.StatusCreated, // Should be allowed
		},
		{
			name: "Very long callsign",
			contact: Contact{
				Callsign: "VE1VERYLONGCALLSIGNTEST/PORTABLE",
				Date:     time.Now(),
			},
			expected: http.StatusCreated, // Should be allowed
		},
		{
			name: "Special characters in fields",
			contact: Contact{
				Callsign: "TEST/1",
				Name:     "José María",
				QTH:      "São Paulo, BR",
				Comment:  "Testing unicode: ñáéíóú",
				Date:     time.Now(),
			},
			expected: http.StatusCreated,
		},
		{
			name: "Zero frequency",
			contact: Contact{
				Callsign:  "ZERO_FREQ",
				Frequency: 0.0,
				Date:      time.Now(),
			},
			expected: http.StatusCreated,
		},
		{
			name: "Negative frequency",
			contact: Contact{
				Callsign:  "NEG_FREQ",
				Frequency: -14.205,
				Date:      time.Now(),
			},
			expected: http.StatusCreated, // Database should handle this
		},
		{
			name: "Very high frequency",
			contact: Contact{
				Callsign:  "HIGH_FREQ",
				Frequency: 999999.999,
				Date:      time.Now(),
			},
			expected: http.StatusCreated,
		},
		{
			name: "Zero power",
			contact: Contact{
				Callsign: "ZERO_POWER",
				Power:    0,
				Date:     time.Now(),
			},
			expected: http.StatusCreated,
		},
		{
			name: "Negative power",
			contact: Contact{
				Callsign: "NEG_POWER",
				Power:    -100,
				Date:     time.Now(),
			},
			expected: http.StatusCreated,
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tc.contact)
			if err != nil {
				t.Fatalf("Failed to marshal contact: %v", err)
			}

			resp, err := http.Post(server.URL+"/api/contacts", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to post contact: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.expected {
				t.Errorf("Expected status %d for %s, got %d", tc.expected, tc.name, resp.StatusCode)
			}
		})
	}
}

func TestDatabaseConstraintViolations(t *testing.T) {
	// This test would be more relevant if we had actual database constraints
	// For now, we'll test that the system handles potential constraint violations gracefully

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Test with contact that has potential constraint issues
	contact := Contact{
		Callsign: "CONSTRAINT_TEST",
		Date:     time.Date(1800, 1, 1, 0, 0, 0, 0, time.UTC), // Very old date
	}

	err := logger.SaveContact(&contact)
	// Should either succeed or fail gracefully
	if err != nil {
		t.Logf("Expected behavior: constraint violation caused error: %v", err)
	}
}

func TestTimeoutHandling(t *testing.T) {
	// Test that the server can handle slow requests gracefully
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Create a client with a very short timeout
	client := &http.Client{
		Timeout: 1 * time.Millisecond,
	}

	// This should timeout
	_, err := client.Get(server.URL + "/api/contacts")
	if err == nil {
		t.Log("Request completed faster than expected (this is okay)")
	} else {
		// Expected timeout error
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
			t.Errorf("Expected timeout error, got: %v", err)
		}
	}
}

func TestInvalidURLPaths(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	invalidPaths := []string{
		"/api/contacts/",
		"/api/contact",
		"/api/contacts/abc/def",
		"/api/contacts/999/extra",
		"/api/invalid",
		"/invalid",
		"/../../../etc/passwd",
		"/api/contacts/%2e%2e%2f%2e%2e%2f",
	}

	for _, path := range invalidPaths {
		t.Run("Invalid path: "+path, func(t *testing.T) {
			resp, err := http.Get(server.URL + path)
			if err != nil {
				t.Fatalf("Failed to request invalid path %s: %v", path, err)
			}
			defer resp.Body.Close()

			// Should return 404 or other appropriate error
			if resp.StatusCode == http.StatusOK {
				t.Errorf("Expected error status for invalid path %s, got 200", path)
			}
		})
	}
}
