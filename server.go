package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// API response types
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ContactRequest struct {
	Callsign     string  `json:"callsign"`
	OperatorName string  `json:"operator_name"`
	ContactDate  string  `json:"contact_date"`
	TimeOn       string  `json:"time_on"`
	TimeOff      string  `json:"time_off"`
	Frequency    float64 `json:"frequency"`
	Band         string  `json:"band"`
	Mode         string  `json:"mode"`
	PowerWatts   int     `json:"power_watts"`
	RSTSent      string  `json:"rst_sent"`
	RSTReceived  string  `json:"rst_received"`
	QTH          string  `json:"qth"`
	Country      string  `json:"country"`
	GridSquare   string  `json:"grid_square"`
	Comment      string  `json:"comment"`
	Confirmed    bool    `json:"confirmed"`
}

type SearchRequest struct {
	Search    string  `json:"search"`
	DateFrom  string  `json:"date_from"`
	DateTo    string  `json:"date_to"`
	Band      string  `json:"band"`
	Mode      string  `json:"mode"`
	Country   string  `json:"country"`
	FreqMin   float64 `json:"freq_min"`
	FreqMax   float64 `json:"freq_max"`
	Confirmed bool    `json:"confirmed"`
}

func enableCORS(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	return c.Handler(next)
}

func setupRoutes(logger *QSOLogger) *mux.Router {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Contacts endpoints
	api.HandleFunc("/contacts", handleGetContacts(logger)).Methods("GET")
	api.HandleFunc("/contacts", handleCreateContact(logger)).Methods("POST")
	api.HandleFunc("/contacts/{id}", handleUpdateContact(logger)).Methods("PUT")
	api.HandleFunc("/contacts/{id}", handleDeleteContact(logger)).Methods("DELETE")
	api.HandleFunc("/contacts/search", handleSearchContacts(logger)).Methods("POST")
	api.HandleFunc("/contacts/export", handleExportContacts(logger)).Methods("GET")

	// Statistics endpoint
	api.HandleFunc("/statistics", handleGetStatistics(logger)).Methods("GET")

	// Health check
	api.HandleFunc("/health", handleHealthCheck).Methods("GET")

	return r
}

func handleGetContacts(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contacts, err := logger.GetAllContacts()
		if err != nil {
			sendError(w, "Failed to retrieve contacts", http.StatusInternalServerError)
			return
		}

		sendSuccess(w, contacts)
	}
}

func handleCreateContact(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ContactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Parse date
		contactDate, err := time.Parse("2006-01-02", req.ContactDate)
		if err != nil {
			sendError(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		contact := Contact{
			Callsign:    strings.ToUpper(strings.TrimSpace(req.Callsign)),
			Name:        strings.TrimSpace(req.OperatorName),
			Date:        contactDate,
			TimeOn:      req.TimeOn,
			TimeOff:     req.TimeOff,
			Frequency:   req.Frequency,
			Band:        req.Band,
			Mode:        strings.ToUpper(req.Mode),
			Power:       req.PowerWatts,
			RSTSent:     req.RSTSent,
			RSTReceived: req.RSTReceived,
			QTH:         strings.TrimSpace(req.QTH),
			Country:     strings.TrimSpace(req.Country),
			Grid:        strings.ToUpper(strings.TrimSpace(req.GridSquare)),
			Comment:     strings.TrimSpace(req.Comment),
			Confirmed:   req.Confirmed,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := logger.AddContactStruct(contact); err != nil {
			sendError(w, fmt.Sprintf("Failed to add contact: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the created contact to return it (find by callsign and date since we don't have the ID)
		contacts, err := logger.GetAllContacts()
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to retrieve created contact: %v", err), http.StatusInternalServerError)
			return
		}

		// Find the most recently created contact that matches our criteria
		var createdContact *Contact
		for i := len(contacts) - 1; i >= 0; i-- {
			c := &contacts[i]
			if c.Callsign == contact.Callsign &&
				c.Date.Format("2006-01-02") == contact.Date.Format("2006-01-02") &&
				c.TimeOn == contact.TimeOn {
				createdContact = c
				break
			}
		}

		if createdContact == nil {
			sendError(w, "Failed to retrieve created contact", http.StatusInternalServerError)
			return
		}

		sendSuccess(w, createdContact)
	}
}

func handleUpdateContact(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]

		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendError(w, "Invalid contact ID", http.StatusBadRequest)
			return
		}

		var req ContactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Parse date
		contactDate, err := time.Parse("2006-01-02", req.ContactDate)
		if err != nil {
			sendError(w, "Invalid date format", http.StatusBadRequest)
			return
		}

		contact := Contact{
			ID:          id,
			Callsign:    strings.ToUpper(strings.TrimSpace(req.Callsign)),
			Name:        strings.TrimSpace(req.OperatorName),
			Date:        contactDate,
			TimeOn:      req.TimeOn,
			TimeOff:     req.TimeOff,
			Frequency:   req.Frequency,
			Band:        req.Band,
			Mode:        strings.ToUpper(req.Mode),
			Power:       req.PowerWatts,
			RSTSent:     req.RSTSent,
			RSTReceived: req.RSTReceived,
			QTH:         strings.TrimSpace(req.QTH),
			Country:     strings.TrimSpace(req.Country),
			Grid:        strings.ToUpper(strings.TrimSpace(req.GridSquare)),
			Comment:     strings.TrimSpace(req.Comment),
			Confirmed:   req.Confirmed,
			UpdatedAt:   time.Now(),
		}

		if err := logger.UpdateContact(contact); err != nil {
			sendError(w, fmt.Sprintf("Failed to update contact: %v", err), http.StatusInternalServerError)
			return
		}

		// Get the updated contact to return it
		updatedContact, err := logger.GetContactByID(id)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to retrieve updated contact: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, updatedContact)
	}
}

func handleDeleteContact(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]

		id, err := strconv.Atoi(idStr)
		if err != nil {
			sendError(w, "Invalid contact ID", http.StatusBadRequest)
			return
		}

		if err := logger.DeleteContact(id); err != nil {
			sendError(w, fmt.Sprintf("Failed to delete contact: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, map[string]string{"message": "Contact deleted successfully"})
	}
}

func handleSearchContacts(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		contacts, err := logger.SearchContactsAPI(req)
		if err != nil {
			sendError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, contacts)
	}
}

func handleExportContacts(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := fmt.Sprintf("goqso_export_%s.adi", time.Now().Format("20060102_150405"))

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		if err := logger.ExportADIFToWriter(w); err != nil {
			sendError(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func handleGetStatistics(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := logger.GetStatistics()
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get statistics: %v", err), http.StatusInternalServerError)
			return
		}

		sendSuccess(w, stats)
	}
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	sendSuccess(w, map[string]string{
		"status":  "healthy",
		"version": version,
		"time":    time.Now().Format(time.RFC3339),
	})
}

func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	}); err != nil {
		log.Printf("Failed to encode success response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   message,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
		// Can't call http.Error here since WriteHeader was already called
		// Just log the error - the client will get the status code at least
	}
}

func startServer() {
	logger, err := NewQSOLogger()
	if err != nil {
		log.Fatalf("Failed to initialize QSO logger: %v", err)
	}
	defer logger.Close()

	router := setupRoutes(logger)
	handler := enableCORS(router)

	port := ":8080"
	fmt.Printf("Starting GoQSO API server on port %s\n", port)
	fmt.Printf("Frontend should be accessible at: http://localhost:3000\n")
	fmt.Printf("API endpoints available at: http://localhost:8080/api\n")

	// Configure server with security timeouts to prevent attacks like Slowloris
	server := &http.Server{
		Addr:              port,
		Handler:           handler,
		ReadTimeout:       15 * time.Second, // Maximum duration for reading the entire request
		WriteTimeout:      15 * time.Second, // Maximum duration before timing out writes
		IdleTimeout:       60 * time.Second, // Maximum amount of time to wait for the next request
		ReadHeaderTimeout: 5 * time.Second,  // Amount of time allowed to read request headers
	}

	log.Fatal(server.ListenAndServe())
}
