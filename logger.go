package main

import (
	"database/sql"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	version = "1.0.0"
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

// Statistics represents QSO statistics
type Statistics struct {
	TotalQSOs       int            `json:"total_qsos"`
	UniqueCallsigns int            `json:"unique_callsigns"`
	UniqueCountries int            `json:"unique_countries"`
	ConfirmedQSOs   int            `json:"confirmed_qsos"`
	QSOsByBand      map[string]int `json:"qsos_by_band"`
	QSOsByMode      map[string]int `json:"qsos_by_mode"`
	QSOsByCountry   map[string]int `json:"qsos_by_country"`
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

// API Methods for HTTP server

// GetAllContacts returns all contacts from the database
func (q *QSOLogger) GetAllContacts() ([]Contact, error) {
	return q.LoadContacts()
}

// AddContactStruct adds a contact using a Contact struct
func (q *QSOLogger) AddContactStruct(contact Contact) error {
	return q.SaveContact(&contact)
}

// DeleteContact deletes a contact by ID
func (q *QSOLogger) DeleteContact(id int) error {
	query := `DELETE FROM contacts WHERE id = $1`
	result, err := q.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contact with ID %d not found", id)
	}

	return nil
}

// GetContactByID retrieves a contact by its ID
func (q *QSOLogger) GetContactByID(id int) (*Contact, error) {
	query := `
		SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode,
		       rst_sent, rst_received, operator_name, qth, country, grid_square,
		       power_watts, comment, confirmed, created_at, updated_at
		FROM contacts
		WHERE id = $1
	`

	var contact Contact
	err := q.db.QueryRow(query, id).Scan(
		&contact.ID,
		&contact.Callsign,
		&contact.Date,
		&contact.TimeOn,
		&contact.TimeOff,
		&contact.Frequency,
		&contact.Band,
		&contact.Mode,
		&contact.RSTSent,
		&contact.RSTReceived,
		&contact.Name,
		&contact.QTH,
		&contact.Country,
		&contact.Grid,
		&contact.Power,
		&contact.Comment,
		&contact.Confirmed,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contact with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get contact: %w", err)
	}

	return &contact, nil
}

// UpdateContact updates an existing contact
func (q *QSOLogger) UpdateContact(contact Contact) error {
	query := `
		UPDATE contacts 
		SET callsign = $1, contact_date = $2, time_on = $3, time_off = $4, frequency = $5,
		    band = $6, mode = $7, rst_sent = $8, rst_received = $9, operator_name = $10,
		    qth = $11, country = $12, grid_square = $13, power_watts = $14, comment = $15,
		    confirmed = $16, updated_at = $17
		WHERE id = $18
	`

	result, err := q.db.Exec(query,
		contact.Callsign,
		contact.Date,
		contact.TimeOn,
		contact.TimeOff,
		contact.Frequency,
		contact.Band,
		contact.Mode,
		contact.RSTSent,
		contact.RSTReceived,
		contact.Name,
		contact.QTH,
		contact.Country,
		contact.Grid,
		contact.Power,
		contact.Comment,
		contact.Confirmed,
		contact.UpdatedAt,
		contact.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update contact: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("contact with ID %d not found", contact.ID)
	}

	return nil
}

// SearchContactsAPI performs search with API filters
func (q *QSOLogger) SearchContactsAPI(filters SearchRequest) ([]Contact, error) {
	query := `
		SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode,
		       rst_sent, rst_received, operator_name, qth, country, grid_square,
		       power_watts, comment, confirmed, created_at, updated_at
		FROM contacts
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	// Build dynamic WHERE clause
	if filters.Search != "" {
		argCount++
		query += fmt.Sprintf(" AND (LOWER(callsign) LIKE LOWER($%d) OR LOWER(operator_name) LIKE LOWER($%d) OR LOWER(qth) LIKE LOWER($%d) OR LOWER(country) LIKE LOWER($%d))", argCount, argCount, argCount, argCount)
		args = append(args, "%"+filters.Search+"%")
	}

	if filters.DateFrom != "" {
		argCount++
		query += fmt.Sprintf(" AND contact_date >= $%d", argCount)
		args = append(args, filters.DateFrom)
	}

	if filters.DateTo != "" {
		argCount++
		query += fmt.Sprintf(" AND contact_date <= $%d", argCount)
		args = append(args, filters.DateTo)
	}

	if filters.Band != "" {
		argCount++
		query += fmt.Sprintf(" AND band = $%d", argCount)
		args = append(args, filters.Band)
	}

	if filters.Mode != "" {
		argCount++
		query += fmt.Sprintf(" AND mode = $%d", argCount)
		args = append(args, filters.Mode)
	}

	if filters.Country != "" {
		argCount++
		query += fmt.Sprintf(" AND LOWER(country) LIKE LOWER($%d)", argCount)
		args = append(args, "%"+filters.Country+"%")
	}

	if filters.FreqMin > 0 {
		argCount++
		query += fmt.Sprintf(" AND frequency >= $%d", argCount)
		args = append(args, filters.FreqMin)
	}

	if filters.FreqMax > 0 {
		argCount++
		query += fmt.Sprintf(" AND frequency <= $%d", argCount)
		args = append(args, filters.FreqMax)
	}

	if filters.Confirmed {
		query += " AND confirmed = true"
	}

	query += " ORDER BY contact_date DESC, time_on DESC"

	rows, err := q.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var contact Contact
		err := rows.Scan(
			&contact.ID, &contact.Callsign, &contact.Date, &contact.TimeOn,
			&contact.TimeOff, &contact.Frequency, &contact.Band, &contact.Mode,
			&contact.RSTSent, &contact.RSTReceived, &contact.Name, &contact.QTH,
			&contact.Country, &contact.Grid, &contact.Power, &contact.Comment,
			&contact.Confirmed, &contact.CreatedAt, &contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contact: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

// GetStatistics returns QSO statistics
func (q *QSOLogger) GetStatistics() (*Statistics, error) {
	stats := &Statistics{
		QSOsByBand:    make(map[string]int),
		QSOsByMode:    make(map[string]int),
		QSOsByCountry: make(map[string]int),
	}

	// Get basic counts
	err := q.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(DISTINCT callsign) as unique_callsigns,
			COUNT(DISTINCT country) as unique_countries,
			COUNT(CASE WHEN confirmed = true THEN 1 END) as confirmed
		FROM contacts
	`).Scan(&stats.TotalQSOs, &stats.UniqueCallsigns, &stats.UniqueCountries, &stats.ConfirmedQSOs)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic statistics: %w", err)
	}

	// Get QSOs by band
	rows, err := q.db.Query("SELECT band, COUNT(*) FROM contacts GROUP BY band ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get band statistics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var band string
		var count int
		if err := rows.Scan(&band, &count); err != nil {
			return nil, fmt.Errorf("failed to scan band statistics: %w", err)
		}
		stats.QSOsByBand[band] = count
	}

	// Get QSOs by mode
	rows, err = q.db.Query("SELECT mode, COUNT(*) FROM contacts GROUP BY mode ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get mode statistics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var mode string
		var count int
		if err := rows.Scan(&mode, &count); err != nil {
			return nil, fmt.Errorf("failed to scan mode statistics: %w", err)
		}
		stats.QSOsByMode[mode] = count
	}

	// Get QSOs by country
	rows, err = q.db.Query("SELECT country, COUNT(*) FROM contacts WHERE country != '' GROUP BY country ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to get country statistics: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var country string
		var count int
		if err := rows.Scan(&country, &count); err != nil {
			return nil, fmt.Errorf("failed to scan country statistics: %w", err)
		}
		stats.QSOsByCountry[country] = count
	}

	return stats, nil
}

// ExportADIFToWriter exports all contacts to ADIF format to a writer
func (q *QSOLogger) ExportADIFToWriter(w io.Writer) error {
	contacts, err := q.LoadContacts()
	if err != nil {
		return fmt.Errorf("failed to load contacts: %w", err)
	}

	// Write ADIF header
	adifHeader := fmt.Sprintf("Generated by GoQSO v%s on %s\n\n<ADIF_VER:5>3.1.0\n<PROGRAMID:5>GoQSO\n<PROGRAMVERSION:%d>%s\n<EOH>\n\n",
		version, time.Now().Format("2006-01-02 15:04:05"), len(version), version)

	if _, err := w.Write([]byte(adifHeader)); err != nil {
		return fmt.Errorf("failed to write ADIF header: %w", err)
	}

	// Write each contact
	for _, contact := range contacts {
		adifRecord := fmt.Sprintf("<CALL:%d>%s ", len(contact.Callsign), contact.Callsign)
		adifRecord += fmt.Sprintf("<QSO_DATE:8>%s ", contact.Date.Format("20060102"))
		adifRecord += fmt.Sprintf("<TIME_ON:6>%s ", strings.ReplaceAll(contact.TimeOn, ":", ""))
		adifRecord += fmt.Sprintf("<TIME_OFF:6>%s ", strings.ReplaceAll(contact.TimeOff, ":", ""))
		adifRecord += fmt.Sprintf("<FREQ:%d>%s ", len(fmt.Sprintf("%.3f", contact.Frequency)), fmt.Sprintf("%.3f", contact.Frequency))
		adifRecord += fmt.Sprintf("<BAND:%d>%s ", len(contact.Band), contact.Band)
		adifRecord += fmt.Sprintf("<MODE:%d>%s ", len(contact.Mode), contact.Mode)
		adifRecord += fmt.Sprintf("<RST_SENT:%d>%s ", len(contact.RSTSent), contact.RSTSent)
		adifRecord += fmt.Sprintf("<RST_RCVD:%d>%s ", len(contact.RSTReceived), contact.RSTReceived)

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
		if _, err := w.Write([]byte(adifRecord)); err != nil {
			return fmt.Errorf("failed to write contact record: %w", err)
		}
	}

	return nil
}

// PaginationResult represents paginated query results
type PaginationResult struct {
	Contacts   []Contact `json:"contacts"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalItems int       `json:"total_items"`
	TotalPages int       `json:"total_pages"`
}

// GetContactsPaginated returns paginated contacts
func (q *QSOLogger) GetContactsPaginated(page, pageSize int) (*PaginationResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize

	// Get total count
	var totalItems int
	countQuery := "SELECT COUNT(*) FROM contacts"
	err := q.db.QueryRow(countQuery).Scan(&totalItems)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get paginated contacts
	query := `
		SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode,
		       rst_sent, rst_received, operator_name, qth, country, grid_square,
		       power_watts, comment, confirmed, created_at, updated_at
		FROM contacts
		ORDER BY contact_date DESC, time_on DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := q.db.Query(query, pageSize, offset)
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

	totalPages := (totalItems + pageSize - 1) / pageSize

	return &PaginationResult{
		Contacts:   contacts,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}

// SearchContactsPaginated performs search with API filters and pagination
func (q *QSOLogger) SearchContactsPaginated(filters SearchRequest) (*PaginationResult, error) {
	page := filters.Page
	pageSize := filters.PageSize

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 20 // Default page size
	}

	offset := (page - 1) * pageSize

	// Build base query with WHERE conditions
	whereConditions := []string{"1=1"}
	args := []interface{}{}

	// Build dynamic WHERE clause
	if filters.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(LOWER(callsign) LIKE LOWER($%d) OR LOWER(operator_name) LIKE LOWER($%d) OR LOWER(qth) LIKE LOWER($%d) OR LOWER(country) LIKE LOWER($%d))", len(args)+1, len(args)+1, len(args)+1, len(args)+1))
		args = append(args, "%"+filters.Search+"%")
	}

	if filters.DateFrom != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("contact_date >= $%d", len(args)+1))
		args = append(args, filters.DateFrom)
	}

	if filters.DateTo != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("contact_date <= $%d", len(args)+1))
		args = append(args, filters.DateTo)
	}

	if filters.Band != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(band) = LOWER($%d)", len(args)+1))
		args = append(args, filters.Band)
	}

	if filters.Mode != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(mode) = LOWER($%d)", len(args)+1))
		args = append(args, filters.Mode)
	}

	if filters.Country != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(country) LIKE LOWER($%d)", len(args)+1))
		args = append(args, "%"+filters.Country+"%")
	}

	if filters.FreqMin > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("frequency >= $%d", len(args)+1))
		args = append(args, filters.FreqMin)
	}

	if filters.FreqMax > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("frequency <= $%d", len(args)+1))
		args = append(args, filters.FreqMax)
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Get total count
	var totalItems int
	countQuery := "SELECT COUNT(*) FROM contacts WHERE " + whereClause
	err := q.db.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Build the main query with proper parameter placeholders
	limitPlaceholder := len(args) + 1
	offsetPlaceholder := len(args) + 2

	// #nosec G202 - This is safe because we only concatenate static SQL parts and parameterized placeholders, no user input
	query := "SELECT id, callsign, contact_date, time_on, time_off, frequency, band, mode, " +
		"rst_sent, rst_received, operator_name, qth, country, grid_square, " +
		"power_watts, comment, confirmed, created_at, updated_at " +
		"FROM contacts WHERE " + whereClause + " " +
		"ORDER BY contact_date DESC, time_on DESC " +
		"LIMIT $" + fmt.Sprintf("%d", limitPlaceholder) + " OFFSET $" + fmt.Sprintf("%d", offsetPlaceholder)

	args = append(args, pageSize, offset)

	rows, err := q.db.Query(query, args...)
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

	totalPages := (totalItems + pageSize - 1) / pageSize

	return &PaginationResult{
		Contacts:   contacts,
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
}
