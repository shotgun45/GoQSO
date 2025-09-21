package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoTWClient handles communication with ARRL's Logbook of the World
// SECURITY NOTE: LoTW API requires credentials to be sent as URL parameters
// which means they will appear in browser network tabs, server logs, and proxy logs.
// This is a limitation of the LoTW API design, not our implementation.
type LoTWClient struct {
	client   *http.Client
	baseURL  string
	username string
	password string
}

// NewLoTWClient creates a new LoTW client
func NewLoTWClient(username, password string) *LoTWClient {
	return &LoTWClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  "https://lotw.arrl.org",
		username: username,
		password: password,
	}
}

// downloadQSOData downloads confirmed QSO data from LoTW in ADIF format
func (c *LoTWClient) downloadQSOData(startDate, endDate string) (string, error) {
	// LoTW actually uses a direct download URL that doesn't require web session authentication
	// The download URL accepts username/password as parameters
	downloadURL := c.baseURL + "/lotwuser/lotwreport.adi"

	// Prepare query parameters for the download
	params := url.Values{}
	params.Set("login", c.username)
	params.Set("password", c.password)
	params.Set("qso_query", "1")

	// Try both confirmed QSOs and all QSOs to see what's available
	params.Set("qso_qsl", "yes")       // Only confirmed QSOs initially
	params.Set("qso_qsldetail", "yes") // Include QSL details
	params.Set("qso_withown", "yes")   // Include own QSOs
	params.Set("qso_mode", "")         // All modes
	params.Set("qso_band", "")         // All bands
	params.Set("qso_dxcc", "")         // All countries
	params.Set("qso_owncall", "")      // All own callsigns

	// If no start date provided, use a date far in the past to get all records
	if startDate == "" {
		startDate = "1945-01-01" // Start from beginning of amateur radio era
	}
	params.Set("qso_qslsince", startDate) // Start date

	if endDate != "" {
		params.Set("qso_enddate", endDate)
	}
	// Note: Not setting qso_enddate will default to current date

	// Make the request without requiring prior authentication
	fullURL := downloadURL + "?" + params.Encode()

	// Create sanitized logging - LoTW API inherently requires credentials in URL
	// Note: Credentials will still appear in browser network tab due to LoTW API design
	fmt.Printf("DEBUG: LoTW Request initiated for user: %s\n", c.username)

	resp, err := c.client.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to download QSO data: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("DEBUG: LoTW Response Status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: LoTW Response Headers: %+v\n", resp.Header)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Read the ADIF data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read download response: %v", err)
	}

	adifData := string(body)

	// Debug: Log what we received (first 1000 chars to see more)
	debugData := adifData
	if len(debugData) > 1000 {
		debugData = debugData[:1000] + "..."
	}
	fmt.Printf("DEBUG: LoTW Response (first 1000 chars): %s\n", debugData)

	// Also log the full response to see everything LoTW sent
	fmt.Printf("DEBUG: Full LoTW Response:\n%s\n", adifData)

	// Check if we got valid ADIF data - look for ADIF header marker
	if !strings.Contains(strings.ToUpper(adifData), "<EOH>") {
		// Check if we got an error page or login page
		if strings.Contains(strings.ToLower(adifData), "login") || strings.Contains(strings.ToLower(adifData), "password") {
			return "", fmt.Errorf("authentication failed - received login page instead of ADIF data")
		}
		if strings.Contains(strings.ToLower(adifData), "error") {
			return "", fmt.Errorf("LoTW returned an error: %s", debugData)
		}
		return "", fmt.Errorf("invalid ADIF data received - missing <EOH> header. Got: %s", debugData)
	}

	// Log the full length for debugging
	fmt.Printf("DEBUG: LoTW Response total length: %d bytes\n", len(adifData))

	return adifData, nil
}

// parseADIFData parses ADIF data into LoTWQSO structs
func (c *LoTWClient) parseADIFData(adifData string) ([]LoTWQSO, error) {
	var qsos []LoTWQSO

	// Create an ADIF parser and parse the data
	parser := NewADIFParser()
	reader := strings.NewReader(adifData)
	records, err := parser.ParseADIF(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ADIF data: %v", err)
	}

	fmt.Printf("DEBUG: ADIF parser found %d records\n", len(records))

	// Convert ADIF records to LoTWQSO structs
	for i, record := range records {
		fmt.Printf("DEBUG: Processing record %d: callsign=%s, date=%s, band=%s, mode=%s\n",
			i+1, record.Callsign, record.Date, record.Band, record.Mode)

		qso := LoTWQSO{
			Call:        record.Callsign,
			Band:        strings.ToUpper(record.Band),
			Mode:        record.Mode,
			QSODate:     record.Date,
			TimeOn:      record.TimeOn,
			Country:     record.Country,
			State:       record.QTH, // QTH field often contains state
			GridSquare:  record.Grid,
			Frequency:   fmt.Sprintf("%.3f", record.Frequency),
			StationCall: c.username,
			MyGridSq:    "",  // LoTW doesn't always provide this
			QSLRcvd:     "Y", // All LoTW data is confirmed
		}
		qsos = append(qsos, qso)
	}

	fmt.Printf("DEBUG: Converted %d LoTWQSO records\n", len(qsos))
	return qsos, nil
}

// LoTWQSO represents a QSO record from LoTW
type LoTWQSO struct {
	Call        string `json:"call"`
	Band        string `json:"band"`
	Mode        string `json:"mode"`
	QSODate     string `json:"qso_date"`
	TimeOn      string `json:"time_on"`
	Country     string `json:"country"`
	State       string `json:"state"`
	GridSquare  string `json:"gridsquare"`
	Frequency   string `json:"freq"`
	StationCall string `json:"station_callsign"`
	MyGridSq    string `json:"my_gridsquare"`
	QSLRcvd     string `json:"qsl_rcvd"`
}

// GetQSOs retrieves QSO data from LoTW by downloading ADIF data directly
func (c *LoTWClient) GetQSOs(startDate, endDate string) ([]LoTWQSO, error) {
	// LoTW download URL accepts credentials directly, no web session needed
	adifData, err := c.downloadQSOData(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to download QSO data: %v", err)
	}

	// Parse ADIF data into LoTWQSO structs
	qsos, err := c.parseADIFData(adifData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ADIF data: %v", err)
	}

	return qsos, nil
}

// ConvertToADIFRecord converts a LoTWQSO to an ADIFRecord
func (q *LoTWQSO) ConvertToADIFRecord() ADIFRecord {
	// Parse frequency
	var freq float64
	if q.Frequency != "" {
		if f, err := parseFloat(q.Frequency); err == nil {
			freq = f
		}
	}

	// Format date (LoTW uses YYYY-MM-DD, we want the same)
	date := q.QSODate
	if len(date) == 10 && strings.Contains(date, "-") {
		// Already in correct format
	} else if len(date) == 8 {
		// Convert YYYYMMDD to YYYY-MM-DD
		date = fmt.Sprintf("%s-%s-%s", date[:4], date[4:6], date[6:8])
	}

	// Format time (ensure HH:MM:SS format)
	timeOn := q.TimeOn
	if len(timeOn) == 4 {
		timeOn = fmt.Sprintf("%s:%s:00", timeOn[:2], timeOn[2:4])
	} else if len(timeOn) == 6 {
		timeOn = fmt.Sprintf("%s:%s:%s", timeOn[:2], timeOn[2:4], timeOn[4:6])
	}

	return ADIFRecord{
		Callsign:    q.Call,
		Date:        date,
		TimeOn:      timeOn,
		TimeOff:     timeOn, // LoTW typically doesn't provide time_off
		Frequency:   freq,
		Band:        strings.ToLower(q.Band), // Convert to lowercase (20m format)
		Mode:        q.Mode,
		RSTSent:     "59", // LoTW doesn't always provide RST
		RSTReceived: "59",
		Name:        "", // LoTW doesn't provide operator names
		QTH:         q.State,
		Country:     q.Country,
		Grid:        q.GridSquare,
		Power:       100, // Default power since LoTW doesn't provide this
		Comment:     "Imported from LoTW",
		Confirmed:   q.QSLRcvd == "Y",
	}
}

// parseFloat safely parses a float value
func parseFloat(s string) (float64, error) {
	// Remove any non-numeric characters except decimal point
	cleaned := ""
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			cleaned += string(r)
		}
	}

	if cleaned == "" {
		return 0, fmt.Errorf("no numeric value found")
	}

	return parseFloatString(cleaned)
}

// parseFloatString parses a clean float string
func parseFloatString(s string) (float64, error) {
	var result float64
	var decimal float64 = 1
	var afterDecimal bool

	for _, r := range s {
		if r == '.' {
			afterDecimal = true
			continue
		}

		digit := float64(r - '0')
		if !afterDecimal {
			result = result*10 + digit
		} else {
			decimal /= 10
			result += digit * decimal
		}
	}

	return result, nil
}

// ImportFromLoTW handles the complete LoTW import process
func ImportFromLoTW(logger *QSOLogger, credentials LotwCredentials, options ImportOptions) ImportResult {
	client := NewLoTWClient(credentials.Username, credentials.Password)

	// Get QSOs from LoTW
	qsos, err := client.GetQSOs(credentials.StartDate, credentials.EndDate)
	if err != nil {
		return ImportResult{
			Success:       false,
			ImportedCount: 0,
			SkippedCount:  0,
			ErrorCount:    1,
			Errors:        []string{fmt.Sprintf("Failed to retrieve data from LoTW: %v", err)},
			Message:       "LoTW import failed",
		}
	}

	fmt.Printf("DEBUG: Retrieved %d QSOs from LoTW for processing\n", len(qsos))

	// Convert LoTW QSOs to ADIF records and import
	result := ImportResult{
		Success:       true,
		ImportedCount: 0,
		SkippedCount:  0,
		ErrorCount:    0,
		Errors:        []string{},
		Message:       fmt.Sprintf("Processing %d confirmed QSOs from LoTW for %s", len(qsos), credentials.Username),
	}

	for i, qso := range qsos {
		fmt.Printf("DEBUG: Processing QSO %d/%d: %s on %s\n", i+1, len(qsos), qso.Call, qso.QSODate)

		adifRecord := qso.ConvertToADIFRecord()
		contactReq := adifRecord.ConvertToContactRequest()

		// Check for duplicates if merge_duplicates is enabled
		if options.MergeDuplicates {
			existing, err := findExistingContact(logger, contactReq.Callsign, contactReq.ContactDate, contactReq.TimeOn)
			if err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("Error checking for duplicate %s: %v", contactReq.Callsign, err))
				continue
			}

			if existing != nil {
				if options.UpdateExisting {
					// Update existing contact with LoTW confirmation
					contactReq.Confirmed = true // LoTW data is always confirmed
					err = updateContact(logger, existing.ID, contactReq)
					if err != nil {
						result.ErrorCount++
						result.Errors = append(result.Errors, fmt.Sprintf("Error updating %s: %v", contactReq.Callsign, err))
					} else {
						result.ImportedCount++
					}
				} else {
					result.SkippedCount++
				}
				continue
			}
		}

		// Create new contact
		contactReq.Confirmed = true // LoTW data is always confirmed
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
		result.Message = fmt.Sprintf("Successfully imported %d confirmed QSOs from LoTW for %s", result.ImportedCount, credentials.Username)
	} else {
		result.Message = fmt.Sprintf("Imported %d QSOs with %d errors from LoTW for %s", result.ImportedCount, result.ErrorCount, credentials.Username)
	}

	return result
}
