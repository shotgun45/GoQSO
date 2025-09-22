package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Global variables
var startTime = time.Now()

// API response types
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Items      interface{} `json:"items"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int         `json:"total_items"`
	TotalPages int         `json:"total_pages"`
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
	Page      int     `json:"page"`      // Current page (1-based)
	PageSize  int     `json:"page_size"` // Items per page
}

type ImportOptions struct {
	FileType        string `json:"file_type"`
	MergeDuplicates bool   `json:"merge_duplicates"`
	UpdateExisting  bool   `json:"update_existing"`
}

type ImportResult struct {
	Success       bool     `json:"success"`
	ImportedCount int      `json:"imported_count"`
	SkippedCount  int      `json:"skipped_count"`
	ErrorCount    int      `json:"error_count"`
	Errors        []string `json:"errors"`
	Message       string   `json:"message"`
}

type LotwCredentials struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type LotwImportRequest struct {
	Credentials LotwCredentials `json:"credentials"`
	Options     ImportOptions   `json:"options"`
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

	// Import endpoints
	api.HandleFunc("/import/adif", handleImportADIF(logger)).Methods("POST")
	api.HandleFunc("/import/lotw", handleImportLoTW(logger)).Methods("POST")

	// Health check
	api.HandleFunc("/health", handleHealthCheck).Methods("GET")

	// Admin endpoints
	api.HandleFunc("/admin/system", handleAdminSystem(logger)).Methods("GET")
	api.HandleFunc("/admin/merge-duplicates", handleMergeDuplicates(logger)).Methods("POST")

	return r
}

func handleGetContacts(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters from query string
		pageStr := r.URL.Query().Get("page")
		pageSizeStr := r.URL.Query().Get("page_size")

		page := 1
		pageSize := 20

		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if pageSizeStr != "" {
			if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 1000 {
				pageSize = ps
			}
		}

		// Get paginated contacts
		result, err := logger.GetContactsPaginated(page, pageSize)
		if err != nil {
			sendError(w, "Failed to retrieve contacts", http.StatusInternalServerError)
			return
		}

		// Return paginated response
		response := PaginatedResponse{
			Items:      result.Contacts,
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: result.TotalItems,
			TotalPages: result.TotalPages,
		}

		sendSuccess(w, response)
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

		// Set default pagination if not provided
		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}

		// Use paginated search
		result, err := logger.SearchContactsPaginated(req)
		if err != nil {
			sendError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Return paginated response
		response := PaginatedResponse{
			Items:      result.Contacts,
			Page:       result.Page,
			PageSize:   result.PageSize,
			TotalItems: result.TotalItems,
			TotalPages: result.TotalPages,
		}

		sendSuccess(w, response)
	}
}

func handleExportContacts(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters for date filtering
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")

		var startDate, endDate *time.Time

		if startDateStr != "" {
			parsed, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				sendError(w, fmt.Sprintf("Invalid start_date format: %v", err), http.StatusBadRequest)
				return
			}
			startDate = &parsed
		}

		if endDateStr != "" {
			parsed, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				sendError(w, fmt.Sprintf("Invalid end_date format: %v", err), http.StatusBadRequest)
				return
			}
			endDate = &parsed
		}

		// Generate filename with date range if specified
		filename := "goqso_export"
		if startDate != nil || endDate != nil {
			if startDate != nil && endDate != nil {
				filename += fmt.Sprintf("_%s_to_%s", startDate.Format("20060102"), endDate.Format("20060102"))
			} else if startDate != nil {
				filename += fmt.Sprintf("_from_%s", startDate.Format("20060102"))
			} else if endDate != nil {
				filename += fmt.Sprintf("_until_%s", endDate.Format("20060102"))
			}
		} else {
			filename += fmt.Sprintf("_%s", time.Now().Format("20060102_150405"))
		}
		filename += ".adi"

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		if err := logger.ExportADIFToWriterFiltered(w, startDate, endDate); err != nil {
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

func handleAdminSystem(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get contact count
		contactCount, err := logger.GetContactCount()
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get contact count: %v", err), http.StatusInternalServerError)
			return
		}

		// Get database size
		dbSize, err := logger.GetDatabaseSize()
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get database size: %v", err), http.StatusInternalServerError)
			return
		}

		// Get duplicate count
		duplicateCount, err := logger.CountDuplicateContacts()
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to get duplicate count: %v", err), http.StatusInternalServerError)
			return
		}

		// Prepare system info
		systemInfo := map[string]interface{}{
			"application": map[string]interface{}{
				"name":       "GoQSO",
				"version":    version,
				"go_version": runtime.Version(),
				"start_time": startTime.Format(time.RFC3339),
				"uptime":     time.Since(startTime).String(),
			},
			"database": map[string]interface{}{
				"contact_count":   contactCount,
				"database_size":   dbSize,
				"duplicate_count": duplicateCount,
			},
		}

		sendSuccess(w, systemInfo)
	}
}

func handleMergeDuplicates(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mergedCount, err := logger.MergeDuplicateContacts()
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to merge duplicate contacts: %v", err), http.StatusInternalServerError)
			return
		}

		result := map[string]interface{}{
			"merged_count": mergedCount,
			"message":      fmt.Sprintf("Successfully merged %d duplicate records", mergedCount),
		}

		sendSuccess(w, result)
	}
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

// handleImportADIF handles ADIF file imports
func handleImportADIF(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form
		err := r.ParseMultipartForm(10 << 20) // 10 MB max
		if err != nil {
			sendError(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get the uploaded file
		file, header, err := r.FormFile("file")
		if err != nil {
			sendError(w, "No file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Parse options
		optionsStr := r.FormValue("options")
		var options ImportOptions
		if optionsStr != "" {
			if err := json.Unmarshal([]byte(optionsStr), &options); err != nil {
				sendError(w, "Invalid options format", http.StatusBadRequest)
				return
			}
		}

		// Parse ADIF file
		parser := NewADIFParser()
		records, err := parser.ParseADIF(file)
		if err != nil {
			sendError(w, fmt.Sprintf("Failed to parse ADIF file: %v", err), http.StatusBadRequest)
			return
		}

		// Import records into database
		result := ImportResult{
			Success:       true,
			ImportedCount: 0,
			SkippedCount:  0,
			ErrorCount:    0,
			Errors:        []string{},
			Message:       fmt.Sprintf("Processing %d records from %s", len(records), header.Filename),
		}

		for _, record := range records {
			contactReq := record.ConvertToContactRequest()

			// Check for duplicates if merge_duplicates OR update_existing is enabled
			if options.MergeDuplicates || options.UpdateExisting {
				existing, err := findExistingContact(logger, contactReq.Callsign, contactReq.ContactDate, contactReq.TimeOn)
				if err != nil {
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("Error checking for duplicate %s: %v", contactReq.Callsign, err))
					continue
				}

				if existing != nil {
					if options.UpdateExisting {
						// Update existing contact
						err = updateContact(logger, existing.ID, contactReq)
						if err != nil {
							result.ErrorCount++
							result.Errors = append(result.Errors, fmt.Sprintf("Error updating %s: %v", contactReq.Callsign, err))
						} else {
							result.ImportedCount++
						}
					} else {
						// MergeDuplicates is enabled but UpdateExisting is not, so skip
						result.SkippedCount++
					}
					continue
				}
			}

			// Create new contact
			_, err := createContact(logger, contactReq)
			if err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("Error creating %s: %v", contactReq.Callsign, err))
			} else {
				result.ImportedCount++
			}
		}

		// Update final message
		if result.ErrorCount == 0 {
			result.Message = fmt.Sprintf("Successfully imported %d contacts from %s", result.ImportedCount, header.Filename)
		} else {
			result.Message = fmt.Sprintf("Imported %d contacts with %d errors from %s", result.ImportedCount, result.ErrorCount, header.Filename)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Failed to encode import result: %v", err)
		}
	}
}

// handleImportLoTW handles Logbook of the World imports
func handleImportLoTW(logger *QSOLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LotwImportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// Validate credentials
		if req.Credentials.Username == "" || req.Credentials.Password == "" {
			sendError(w, "Username and password are required", http.StatusBadRequest)
			return
		}

		// Import from LoTW
		result := ImportFromLoTW(logger, req.Credentials, req.Options)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Failed to encode import result: %v", err)
		}
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
