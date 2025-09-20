package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	dataFile = "qso_log.json"
	version  = "1.0.0"
)

// Contact represents an amateur radio QSO (contact)
type Contact struct {
	Callsign    string    `json:"callsign"`
	Date        time.Time `json:"date"`
	TimeOn      string    `json:"time_on"`
	TimeOff     string    `json:"time_off"`
	Frequency   float64   `json:"frequency"`    // MHz
	Band        string    `json:"band"`         // e.g., "20m", "40m"
	Mode        string    `json:"mode"`         // e.g., "SSB", "CW", "FT8"
	RSTSent     string    `json:"rst_sent"`     // Signal report sent
	RSTReceived string    `json:"rst_received"` // Signal report received
	Name        string    `json:"name"`         // Operator name
	QTH         string    `json:"qth"`          // Location
	Country     string    `json:"country"`
	Grid        string    `json:"grid"`  // Maidenhead grid
	Power       int       `json:"power"` // Watts
	Comment     string    `json:"comment"`
	Confirmed   bool      `json:"confirmed"` // QSL confirmed
}

// QSOLogger manages the collection of amateur radio contacts
type QSOLogger struct {
	Contacts []Contact `json:"contacts"`
}

// NewQSOLogger creates a new QSO logger instance
func NewQSOLogger() *QSOLogger {
	logger := &QSOLogger{
		Contacts: make([]Contact, 0),
	}
	logger.LoadContacts()
	return logger
}

// LoadContacts loads QSO data from JSON file
func (q *QSOLogger) LoadContacts() {
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return // File doesn't exist, start with empty log
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}

	err = json.Unmarshal(data, q)
	if err != nil {
		fmt.Printf("Error parsing log file: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d QSOs from log file.\n", len(q.Contacts))
}

// SaveContacts saves QSO data to JSON file
func (q *QSOLogger) SaveContacts() {
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling data: %v\n", err)
		return
	}

	err = os.WriteFile(dataFile, data, 0644)
	if err != nil {
		fmt.Printf("Error saving log file: %v\n", err)
		return
	}
}

// AddContact prompts user for QSO details and adds to log
func (q *QSOLogger) AddContact() {
	fmt.Println("\n=== ADD NEW QSO ===")

	contact := Contact{
		Date: time.Now(),
	}

	// Required fields
	contact.Callsign = strings.ToUpper(getUserInput("Callsign: "))
	if contact.Callsign == "" {
		fmt.Println("Callsign is required!")
		return
	}

	// Frequency and band
	freqStr := getUserInput("Frequency (MHz): ")
	if freq, err := strconv.ParseFloat(freqStr, 64); err == nil {
		contact.Frequency = freq
		contact.Band = frequencyToBand(freq)
	}

	// Mode
	contact.Mode = strings.ToUpper(getUserInput("Mode (SSB/CW/FT8/FT4/PSK31/RTTY): "))
	if contact.Mode == "" {
		contact.Mode = "SSB"
	}

	// Time
	contact.TimeOn = getUserInput("Time ON (HHMM UTC, press Enter for current): ")
	if contact.TimeOn == "" {
		contact.TimeOn = time.Now().UTC().Format("1504")
	}

	contact.TimeOff = getUserInput("Time OFF (HHMM UTC, press Enter for current): ")
	if contact.TimeOff == "" {
		contact.TimeOff = time.Now().UTC().Format("1504")
	}

	// Signal reports
	contact.RSTSent = getUserInput("RST Sent (e.g., 599): ")
	contact.RSTReceived = getUserInput("RST Received (e.g., 599): ")

	// Optional fields
	contact.Name = getUserInput("Name: ")
	contact.QTH = getUserInput("QTH (Location): ")
	contact.Country = getUserInput("Country: ")
	contact.Grid = strings.ToUpper(getUserInput("Grid Square: "))

	powerStr := getUserInput("Power (Watts): ")
	if power, err := strconv.Atoi(powerStr); err == nil {
		contact.Power = power
	}

	contact.Comment = getUserInput("Comment: ")

	q.Contacts = append(q.Contacts, contact)
	q.SaveContacts()

	fmt.Printf("\nQSO with %s added successfully!\n", contact.Callsign)
}

// ListContacts displays all QSOs in the log
func (q *QSOLogger) ListContacts() {
	if len(q.Contacts) == 0 {
		fmt.Println("\nNo QSOs found in log.")
		return
	}

	fmt.Printf("\n=== QSO LOG (%d contacts) ===\n", len(q.Contacts))
	fmt.Println("Date       Time  Callsign     Freq     Band  Mode  RST S/R  Name")
	fmt.Println("--------------------------------------------------------------------")

	// Sort by date (newest first)
	sort.Slice(q.Contacts, func(i, j int) bool {
		return q.Contacts[i].Date.After(q.Contacts[j].Date)
	})

	for _, contact := range q.Contacts {
		fmt.Printf("%-10s %-5s %-12s %8.3f %-5s %-5s %-7s %s\n",
			contact.Date.Format("2006-01-02"),
			contact.TimeOn,
			contact.Callsign,
			contact.Frequency,
			contact.Band,
			contact.Mode,
			contact.RSTSent+"/"+contact.RSTReceived,
			contact.Name)
	}
}

// SearchContacts allows searching for specific QSOs
func (q *QSOLogger) SearchContacts() {
	if len(q.Contacts) == 0 {
		fmt.Println("\nNo QSOs found in log.")
		return
	}

	fmt.Println("\n=== SEARCH QSOs ===")
	searchTerm := strings.ToUpper(getUserInput("Enter callsign or partial callsign: "))

	if searchTerm == "" {
		return
	}

	found := false
	fmt.Printf("\n=== SEARCH RESULTS for '%s' ===\n", searchTerm)

	for _, contact := range q.Contacts {
		if strings.Contains(contact.Callsign, searchTerm) {
			if !found {
				fmt.Println("Date       Time  Callsign     Freq     Band  Mode  RST S/R  Name      QTH")
				fmt.Println("-------------------------------------------------------------------------------")
				found = true
			}
			fmt.Printf("%-10s %-5s %-12s %8.3f %-5s %-5s %-7s %-9s %s\n",
				contact.Date.Format("2006-01-02"),
				contact.TimeOn,
				contact.Callsign,
				contact.Frequency,
				contact.Band,
				contact.Mode,
				contact.RSTSent+"/"+contact.RSTReceived,
				contact.Name,
				contact.QTH)
		}
	}

	if !found {
		fmt.Printf("No QSOs found matching '%s'\n", searchTerm)
	}
}

// ShowStatistics displays various statistics about the QSO log
func (q *QSOLogger) ShowStatistics() {
	if len(q.Contacts) == 0 {
		fmt.Println("\nNo QSOs found in log.")
		return
	}

	fmt.Println("\n=== QSO LOG STATISTICS ===")

	// Basic counts
	fmt.Printf("Total QSOs: %d\n", len(q.Contacts))

	// Band statistics
	bandCounts := make(map[string]int)
	modeCounts := make(map[string]int)
	countryCounts := make(map[string]int)
	confirmedCount := 0

	for _, contact := range q.Contacts {
		if contact.Band != "" {
			bandCounts[contact.Band]++
		}
		if contact.Mode != "" {
			modeCounts[contact.Mode]++
		}
		if contact.Country != "" {
			countryCounts[contact.Country]++
		}
		if contact.Confirmed {
			confirmedCount++
		}
	}

	fmt.Printf("Confirmed QSOs: %d (%.1f%%)\n", confirmedCount,
		float64(confirmedCount)/float64(len(q.Contacts))*100)

	// Bands worked
	fmt.Printf("\nBands worked (%d):\n", len(bandCounts))
	bandNames := make([]string, 0, len(bandCounts))
	for band := range bandCounts {
		bandNames = append(bandNames, band)
	}
	sort.Strings(bandNames)

	for _, band := range bandNames {
		fmt.Printf("  %-8s: %d QSOs\n", band, bandCounts[band])
	}

	// Modes used
	fmt.Printf("\nModes used (%d):\n", len(modeCounts))
	modeNames := make([]string, 0, len(modeCounts))
	for mode := range modeCounts {
		modeNames = append(modeNames, mode)
	}
	sort.Strings(modeNames)

	for _, mode := range modeNames {
		fmt.Printf("  %-8s: %d QSOs\n", mode, modeCounts[mode])
	}

	// Countries worked
	fmt.Printf("\nCountries worked (%d):\n", len(countryCounts))
	if len(countryCounts) <= 10 {
		// Show all if 10 or fewer
		countryNames := make([]string, 0, len(countryCounts))
		for country := range countryCounts {
			countryNames = append(countryNames, country)
		}
		sort.Strings(countryNames)

		for _, country := range countryNames {
			if country != "" {
				fmt.Printf("  %-20s: %d QSOs\n", country, countryCounts[country])
			}
		}
	} else {
		// Show top 10
		type countryCount struct {
			name  string
			count int
		}

		var counts []countryCount
		for country, count := range countryCounts {
			if country != "" {
				counts = append(counts, countryCount{country, count})
			}
		}

		sort.Slice(counts, func(i, j int) bool {
			return counts[i].count > counts[j].count
		})

		fmt.Println("  Top 10:")
		for i, cc := range counts {
			if i >= 10 {
				break
			}
			fmt.Printf("  %-20s: %d QSOs\n", cc.name, cc.count)
		}
	}

	// Date range
	if len(q.Contacts) > 0 {
		earliest := q.Contacts[0].Date
		latest := q.Contacts[0].Date

		for _, contact := range q.Contacts {
			if contact.Date.Before(earliest) {
				earliest = contact.Date
			}
			if contact.Date.After(latest) {
				latest = contact.Date
			}
		}

		fmt.Printf("\nDate range: %s to %s\n",
			earliest.Format("2006-01-02"),
			latest.Format("2006-01-02"))
	}
}

// ExportADIF exports QSOs to Amateur Data Interchange Format
func (q *QSOLogger) ExportADIF() {
	if len(q.Contacts) == 0 {
		fmt.Println("\nNo QSOs to export.")
		return
	}

	filename := fmt.Sprintf("goqso_export_%s.adi", time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating ADIF file: %v\n", err)
		return
	}
	defer file.Close()

	// ADIF header
	header := fmt.Sprintf("Amateur Data Interchange Format export from GoQSO v%s\n", version)
	header += fmt.Sprintf("Export date: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	header += fmt.Sprintf("Total QSOs: %d\n", len(q.Contacts))
	header += "<EOH>\n\n"

	file.WriteString(header)

	// Export each contact
	for _, contact := range q.Contacts {
		adifRecord := ""

		if contact.Callsign != "" {
			adifRecord += fmt.Sprintf("<CALL:%d>%s ", len(contact.Callsign), contact.Callsign)
		}

		dateStr := contact.Date.Format("20060102")
		adifRecord += fmt.Sprintf("<QSO_DATE:%d>%s ", len(dateStr), dateStr)

		if contact.TimeOn != "" {
			adifRecord += fmt.Sprintf("<TIME_ON:%d>%s ", len(contact.TimeOn), contact.TimeOn)
		}

		if contact.TimeOff != "" {
			adifRecord += fmt.Sprintf("<TIME_OFF:%d>%s ", len(contact.TimeOff), contact.TimeOff)
		}

		if contact.Frequency > 0 {
			freqStr := fmt.Sprintf("%.3f", contact.Frequency)
			adifRecord += fmt.Sprintf("<FREQ:%d>%s ", len(freqStr), freqStr)
		}

		if contact.Band != "" {
			adifRecord += fmt.Sprintf("<BAND:%d>%s ", len(contact.Band), contact.Band)
		}

		if contact.Mode != "" {
			adifRecord += fmt.Sprintf("<MODE:%d>%s ", len(contact.Mode), contact.Mode)
		}

		if contact.RSTSent != "" {
			adifRecord += fmt.Sprintf("<RST_SENT:%d>%s ", len(contact.RSTSent), contact.RSTSent)
		}

		if contact.RSTReceived != "" {
			adifRecord += fmt.Sprintf("<RST_RCVD:%d>%s ", len(contact.RSTReceived), contact.RSTReceived)
		}

		if contact.Name != "" {
			adifRecord += fmt.Sprintf("<NAME:%d>%s ", len(contact.Name), contact.Name)
		}

		if contact.QTH != "" {
			adifRecord += fmt.Sprintf("<QTH:%d>%s ", len(contact.QTH), contact.QTH)
		}

		if contact.Country != "" {
			adifRecord += fmt.Sprintf("<COUNTRY:%d>%s ", len(contact.Country), contact.Country)
		}

		if contact.Grid != "" {
			adifRecord += fmt.Sprintf("<GRIDSQUARE:%d>%s ", len(contact.Grid), contact.Grid)
		}

		if contact.Power > 0 {
			powerStr := fmt.Sprintf("%d", contact.Power)
			adifRecord += fmt.Sprintf("<TX_PWR:%d>%s ", len(powerStr), powerStr)
		}

		if contact.Comment != "" {
			adifRecord += fmt.Sprintf("<COMMENT:%d>%s ", len(contact.Comment), contact.Comment)
		}

		adifRecord += "<EOR>\n"
		file.WriteString(adifRecord)
	}

	fmt.Printf("\nSuccessfully exported %d QSOs to %s\n", len(q.Contacts), filename)
}
