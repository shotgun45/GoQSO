package main

import (
	"database/sql"
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
	ID          int       `db:"id"`
	Callsign    string    `db:"callsign"`
	Date        time.Time `db:"contact_date"`
	TimeOn      string    `db:"time_on"`
	TimeOff     string    `db:"time_off"`
	Frequency   float64   `db:"frequency"`     // MHz
	Band        string    `db:"band"`          // e.g., "20m", "40m"
	Mode        string    `db:"mode"`          // e.g., "SSB", "CW", "FT8"
	RSTSent     string    `db:"rst_sent"`      // Signal report sent
	RSTReceived string    `db:"rst_received"`  // Signal report received
	Name        string    `db:"operator_name"` // Operator name
	QTH         string    `db:"qth"`           // Location
	Country     string    `db:"country"`
	Grid        string    `db:"grid_square"` // Maidenhead grid
	Power       int       `db:"power_watts"` // Watts
	Comment     string    `db:"comment"`
	Confirmed   bool      `db:"confirmed"` // QSL confirmed
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// QSOLogger manages the collection of amateur radio contacts using PostgreSQL
type QSOLogger struct {
	db *sql.DB
}

// NewQSOLogger creates a new QSO logger instance with database connection
func NewQSOLogger() (*QSOLogger, error) {
	db, err := InitializeDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger := &QSOLogger{
		db: db,
	}

	return logger, nil
}

// Close closes the database connection
func (q *QSOLogger) Close() error {
	if q.db != nil {
		return q.db.Close()
	}
	return nil
}

// LoadContacts loads QSO data from PostgreSQL database
func (q *QSOLogger) LoadContacts() ([]Contact, error) {
	query := `
		SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode,
		       rst_sent, rst_received, operator_name, qth, country, grid_square,
		       power_watts, comment, confirmed, created_at, updated_at
		FROM contacts
		ORDER BY contact_date DESC, time_on DESC
	`

	rows, err := q.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(
			&contact.ID, &contact.Callsign, &contact.Date, &contact.TimeOn, &contact.TimeOff,
			&contact.Frequency, &contact.Band, &contact.Mode, &contact.RSTSent, &contact.RSTReceived,
			&contact.Name, &contact.QTH, &contact.Country, &contact.Grid, &contact.Power,
			&contact.Comment, &contact.Confirmed, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contacts: %w", err)
	}

	return contacts, nil
}

// SaveContact saves a QSO contact to PostgreSQL database
func (q *QSOLogger) SaveContact(contact *Contact) error {
	query := `
		INSERT INTO contacts (
			callsign, contact_date, time_on, time_off, frequency, band, mode,
			rst_sent, rst_received, operator_name, qth, country, grid_square,
			power_watts, comment, confirmed
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at, updated_at
	`

	err := q.db.QueryRow(
		query,
		contact.Callsign, contact.Date, contact.TimeOn, contact.TimeOff,
		contact.Frequency, contact.Band, contact.Mode, contact.RSTSent, contact.RSTReceived,
		contact.Name, contact.QTH, contact.Country, contact.Grid, contact.Power,
		contact.Comment, contact.Confirmed,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to save contact: %w", err)
	}

	return nil
}

// AddContact prompts user for QSO details and adds to database
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

	// Save to database
	if err := q.SaveContact(&contact); err != nil {
		fmt.Printf("Error saving contact: %v\n", err)
		return
	}

	fmt.Printf("\nQSO with %s added successfully! (ID: %d)\n", contact.Callsign, contact.ID)
}

// ListContacts displays all QSOs from the database
func (q *QSOLogger) ListContacts() {
	contacts, err := q.LoadContacts()
	if err != nil {
		fmt.Printf("Error loading contacts: %v\n", err)
		return
	}

	if len(contacts) == 0 {
		fmt.Println("\nNo QSOs found in database.")
		return
	}

	fmt.Printf("\n=== QSO LOG (%d contacts) ===\n", len(contacts))
	fmt.Println("Date       Time  Callsign     Freq     Band  Mode  RST S/R  Name")
	fmt.Println("--------------------------------------------------------------------")

	// Contacts are already sorted by date DESC from LoadContacts query
	for _, contact := range contacts {
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

// SearchContacts allows searching for specific QSOs by callsign
func (q *QSOLogger) SearchContacts() {
	fmt.Println("\n=== SEARCH QSOs ===")
	searchTerm := strings.ToUpper(getUserInput("Enter callsign or partial callsign: "))

	if searchTerm == "" {
		return
	}

	query := `
		SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode,
		       rst_sent, rst_received, operator_name, qth, country, grid_square,
		       power_watts, comment, confirmed, created_at, updated_at
		FROM contacts
		WHERE UPPER(callsign) LIKE $1
		ORDER BY contact_date DESC, time_on DESC
	`

	rows, err := q.db.Query(query, "%"+searchTerm+"%")
	if err != nil {
		fmt.Printf("Error searching contacts: %v\n", err)
		return
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(
			&contact.ID, &contact.Callsign, &contact.Date, &contact.TimeOn, &contact.TimeOff,
			&contact.Frequency, &contact.Band, &contact.Mode, &contact.RSTSent, &contact.RSTReceived,
			&contact.Name, &contact.QTH, &contact.Country, &contact.Grid, &contact.Power,
			&contact.Comment, &contact.Confirmed, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			fmt.Printf("Error scanning contact: %v\n", err)
			continue
		}
		contacts = append(contacts, contact)
	}

	if len(contacts) == 0 {
		fmt.Printf("No QSOs found matching '%s'\n", searchTerm)
		return
	}

	fmt.Printf("\n=== SEARCH RESULTS for '%s' (%d found) ===\n", searchTerm, len(contacts))
	fmt.Println("Date       Time  Callsign     Freq     Band  Mode  RST S/R  Name      QTH")
	fmt.Println("-------------------------------------------------------------------------------")

	for _, contact := range contacts {
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

// ShowStatistics displays various statistics about the QSO log from database
func (q *QSOLogger) ShowStatistics() {
	contacts, err := q.LoadContacts()
	if err != nil {
		fmt.Printf("Error loading contacts: %v\n", err)
		return
	}

	if len(contacts) == 0 {
		fmt.Println("\nNo QSOs found in database.")
		return
	}

	fmt.Println("\n=== QSO LOG STATISTICS ===")

	// Basic counts
	fmt.Printf("Total QSOs: %d\n", len(contacts))

	// Band statistics
	bandCounts := make(map[string]int)
	modeCounts := make(map[string]int)
	countryCounts := make(map[string]int)
	confirmedCount := 0

	for _, contact := range contacts {
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
		float64(confirmedCount)/float64(len(contacts))*100)

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
	if len(contacts) > 0 {
		earliest := contacts[0].Date
		latest := contacts[0].Date

		for _, contact := range contacts {
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
	contacts, err := q.LoadContacts()
	if err != nil {
		fmt.Printf("Error loading contacts: %v\n", err)
		return
	}

	if len(contacts) == 0 {
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
	header += fmt.Sprintf("Total QSOs: %d\n", len(contacts))
	header += "<EOH>\n\n"

	file.WriteString(header)

	// Export each contact
	for _, contact := range contacts {
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

	fmt.Printf("\nSuccessfully exported %d QSOs to %s\n", len(contacts), filename)
}
