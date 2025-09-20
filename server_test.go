package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// setupTestServer creates a test server with database
func setupTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	// Setup test database
	db := setupTestDB(t)

	logger := &QSOLogger{db: db}

	// Create router manually for testing
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	// Add the routes using the actual handler functions
	api.HandleFunc("/contacts", handleGetContacts(logger)).Methods("GET")
	api.HandleFunc("/contacts", handleCreateContact(logger)).Methods("POST")
	api.HandleFunc("/contacts/{id}", handleUpdateContact(logger)).Methods("PUT")
	api.HandleFunc("/contacts/{id}", handleDeleteContact(logger)).Methods("DELETE")

	// Add CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Create server with test routes
	server := httptest.NewServer(c.Handler(router))

	return server, db
}

func TestGetContactsEndpoint(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test empty contacts list
	resp, err := http.Get(server.URL + "/api/contacts")
	if err != nil {
		t.Fatalf("Failed to get contacts: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("Expected success true")
	}

	contacts, ok := response.Data.([]interface{})
	if !ok {
		t.Error("Expected data to be array")
	}

	if len(contacts) != 0 {
		t.Errorf("Expected 0 contacts, got %d", len(contacts))
	}
}

func TestPostContactEndpoint(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test adding a new contact
	newContact := Contact{
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
		Comment:     "Test QSO via API",
		Confirmed:   false,
	}

	jsonData, err := json.Marshal(newContact)
	if err != nil {
		t.Fatalf("Failed to marshal contact: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/contacts", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to post contact: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %t. Error: %s", response.Success, response.Error)
	}

	// Verify contact was saved by getting all contacts
	resp2, err := http.Get(server.URL + "/api/contacts")
	if err != nil {
		t.Fatalf("Failed to get contacts: %v", err)
	}
	defer resp2.Body.Close()

	var response2 APIResponse
	err = json.NewDecoder(resp2.Body).Decode(&response2)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	contacts, ok := response2.Data.([]interface{})
	if !ok {
		t.Error("Expected data to be array")
	}

	if len(contacts) != 1 {
		t.Errorf("Expected 1 contact after POST, got %d", len(contacts))
	}
}

func TestPostContactInvalidData(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test with invalid JSON
	invalidJSON := `{"callsign": "TEST", "invalid_json"`

	resp, err := http.Post(server.URL+"/api/contacts", "application/json", strings.NewReader(invalidJSON))
	if err != nil {
		t.Fatalf("Failed to post invalid contact: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if response.Success {
		t.Error("Expected success false for invalid JSON")
	}
}

func TestPutContactEndpoint(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// First, add a contact to update
	originalContact := Contact{
		Callsign:    "UPDATE_ME",
		Date:        time.Date(2025, 9, 20, 15, 0, 0, 0, time.UTC),
		TimeOn:      "15:00:00",
		TimeOff:     "15:10:00",
		Frequency:   21.205,
		Band:        "15m",
		Mode:        "SSB",
		RSTSent:     "599",
		RSTReceived: "579",
		Name:        "Original",
		QTH:         "Original QTH",
		Country:     "USA",
		Power:       100,
		Comment:     "Original comment",
		Confirmed:   false,
	}

	err := logger.SaveContact(&originalContact)
	if err != nil {
		t.Fatalf("Failed to save original contact: %v", err)
	}

	// Update the contact
	updatedContact := originalContact
	updatedContact.Name = "Updated Name"
	updatedContact.QTH = "Updated QTH"
	updatedContact.Power = 250
	updatedContact.Comment = "Updated comment"
	updatedContact.Confirmed = true

	jsonData, err := json.Marshal(updatedContact)
	if err != nil {
		t.Fatalf("Failed to marshal updated contact: %v", err)
	}

	// Create PUT request
	url := fmt.Sprintf("%s/api/contacts/%d", server.URL, originalContact.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute PUT request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %t. Error: %s", response.Success, response.Error)
	}

	// Verify the contact was actually updated
	retrieved, err := logger.GetContactByID(originalContact.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated contact: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got %s", retrieved.Name)
	}

	if retrieved.Power != 250 {
		t.Errorf("Expected power 250, got %d", retrieved.Power)
	}

	if !retrieved.Confirmed {
		t.Error("Expected confirmed true")
	}
}

func TestDeleteContactEndpoint(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	logger := &QSOLogger{db: db}

	// Add a contact to delete
	contactToDelete := Contact{
		Callsign: "DELETE_ME",
		Date:     time.Date(2025, 9, 20, 16, 0, 0, 0, time.UTC),
		Mode:     "CW",
		Band:     "40m",
	}

	err := logger.SaveContact(&contactToDelete)
	if err != nil {
		t.Fatalf("Failed to save contact to delete: %v", err)
	}

	// Delete the contact
	url := fmt.Sprintf("%s/api/contacts/%d", server.URL, contactToDelete.ID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success true, got %t. Error: %s", response.Success, response.Error)
	}

	// Verify the contact was actually deleted
	_, err = logger.GetContactByID(contactToDelete.ID)
	if err == nil {
		t.Error("Expected error when getting deleted contact")
	}
}

func TestDeleteNonexistentContact(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Try to delete contact that doesn't exist
	url := fmt.Sprintf("%s/api/contacts/999", server.URL)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute DELETE request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	var response APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("Expected success false for nonexistent contact")
	}
}

func TestCORSHeaders(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test CORS preflight request
	req, err := http.NewRequest("OPTIONS", server.URL+"/api/contacts", nil)
	if err != nil {
		t.Fatalf("Failed to create OPTIONS request: %v", err)
	}
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute OPTIONS request: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS headers
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected Access-Control-Allow-Origin *, got %s", resp.Header.Get("Access-Control-Allow-Origin"))
	}

	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "POST") {
		t.Error("Expected Access-Control-Allow-Methods to include POST")
	}

	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Error("Expected Access-Control-Allow-Headers to include Content-Type")
	}
}

func TestGetContactByIDEndpoint(t *testing.T) {
	// This test is skipped because there's no GET /api/contacts/{id} endpoint
	// The application only supports getting all contacts via GET /api/contacts
	t.Skip("GET by ID endpoint not implemented in current API")
}

func TestInvalidIDParameterEndpoints(t *testing.T) {
	server, db := setupTestServer(t)
	defer server.Close()
	defer teardownTestDB(t, db)

	// Test invalid ID parameter
	invalidIDs := []string{"abc", "0.5", "-1", "999999999999999999999"}

	for _, invalidID := range invalidIDs {
		t.Run("Invalid ID: "+invalidID, func(t *testing.T) {
			// Test GET with invalid ID
			url := fmt.Sprintf("%s/api/contacts/%s", server.URL, invalidID)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to get contact with invalid ID: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400 for invalid ID %s, got %d", invalidID, resp.StatusCode)
			}

			// Test DELETE with invalid ID
			req, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				t.Fatalf("Failed to create DELETE request with invalid ID: %v", err)
			}

			client := &http.Client{}
			resp2, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute DELETE request with invalid ID: %v", err)
			}
			defer resp2.Body.Close()

			if resp2.StatusCode != http.StatusBadRequest {
				t.Errorf("Expected status 400 for DELETE with invalid ID %s, got %d", invalidID, resp2.StatusCode)
			}
		})
	}
}
