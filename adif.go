package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ADIFRecord represents a single QSO record from an ADIF file
type ADIFRecord struct {
	Callsign    string
	Date        string
	TimeOn      string
	TimeOff     string
	Frequency   float64
	Band        string
	Mode        string
	RSTSent     string
	RSTReceived string
	Name        string
	QTH         string
	Country     string
	Grid        string
	Power       int
	Comment     string
	Confirmed   bool
}

// ADIFParser handles parsing of ADIF files
type ADIFParser struct {
	fieldRegex *regexp.Regexp
}

// NewADIFParser creates a new ADIF parser
func NewADIFParser() *ADIFParser {
	// ADIF field format: <FIELD_NAME:LENGTH:TYPE>DATA
	// We'll use a simpler regex that captures field name, length, and data
	fieldRegex := regexp.MustCompile(`<([^:>]+):(\d+)(?::[^>]*)?>([^<]*)`)

	return &ADIFParser{
		fieldRegex: fieldRegex,
	}
}

// ParseADIF parses an ADIF file and returns a slice of QSO records
func (p *ADIFParser) ParseADIF(reader io.Reader) ([]ADIFRecord, error) {
	var records []ADIFRecord
	scanner := bufio.NewScanner(reader)

	var content strings.Builder

	// Read the entire file content
	for scanner.Scan() {
		line := scanner.Text()
		// Skip header comments and metadata
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		content.WriteString(line)
		content.WriteString(" ")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ADIF file: %v", err)
	}

	// Parse records
	adifContent := content.String()

	// Split by <EOR> or <eor> (End of Record) markers - case insensitive
	// Use a simpler approach: find all <eor> positions (case insensitive)
	adifLower := strings.ToLower(adifContent)

	// Split by <eor> (case insensitive)
	recordStrings := strings.Split(adifLower, "<eor>")

	// Get the original case content for each record
	var originalRecords []string
	currentPos := 0

	for i := range recordStrings {
		if i == len(recordStrings)-1 {
			// Last segment after final <eor>
			break
		}

		// Find the next <eor> position in original content
		nextEorPos := strings.Index(strings.ToLower(adifContent[currentPos:]), "<eor>")
		if nextEorPos == -1 {
			break
		}

		// Extract the record content in original case
		recordEnd := currentPos + nextEorPos
		recordContent := strings.TrimSpace(adifContent[currentPos:recordEnd])

		if recordContent != "" {
			originalRecords = append(originalRecords, recordContent)
		}

		// Move past this <eor>
		currentPos = recordEnd + 5
	}

	// Use originalRecords instead of recordStrings
	recordStrings = originalRecords

	fmt.Printf("DEBUG: Found %d record sections in ADIF data\n", len(recordStrings))

	for _, recordStr := range recordStrings {
		recordStr = strings.TrimSpace(recordStr)
		if recordStr == "" {
			continue
		}

		record, err := p.parseRecord(recordStr)
		if err != nil {
			// Log the error but continue parsing other records
			fmt.Printf("Warning: Failed to parse record: %v\n", err)
			continue
		}

		records = append(records, record)
	}

	return records, nil
}

// parseRecord parses a single ADIF record string
func (p *ADIFParser) parseRecord(recordStr string) (ADIFRecord, error) {
	record := ADIFRecord{}

	// Find all field matches
	matches := p.fieldRegex.FindAllStringSubmatch(recordStr, -1)

	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		fieldName := strings.ToUpper(match[1])
		lengthStr := match[2]
		data := match[3]

		// Parse field length and extract the correct amount of data
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			continue
		}

		// Ensure we don't go beyond the available data
		if len(data) < length {
			length = len(data)
		}

		fieldValue := data[:length]

		// Map ADIF fields to our record structure
		switch fieldName {
		case "CALL":
			record.Callsign = fieldValue
		case "QSO_DATE":
			record.Date = p.formatDate(fieldValue)
		case "TIME_ON":
			record.TimeOn = p.formatTime(fieldValue)
		case "TIME_OFF":
			record.TimeOff = p.formatTime(fieldValue)
		case "FREQ":
			if freq, err := strconv.ParseFloat(fieldValue, 64); err == nil {
				record.Frequency = freq
			}
		case "BAND":
			record.Band = fieldValue
		case "MODE":
			record.Mode = fieldValue
		case "RST_SENT":
			record.RSTSent = fieldValue
		case "RST_RCVD":
			record.RSTReceived = fieldValue
		case "NAME":
			record.Name = fieldValue
		case "QTH":
			record.QTH = fieldValue
		case "COUNTRY":
			record.Country = fieldValue
		case "GRIDSQUARE":
			record.Grid = fieldValue
		case "TX_PWR":
			if power, err := strconv.Atoi(fieldValue); err == nil {
				record.Power = power
			}
		case "COMMENT":
			record.Comment = fieldValue
		case "QSL_RCVD":
			record.Confirmed = strings.ToUpper(fieldValue) == "Y"
		}
	}

	// Validate required fields
	if record.Callsign == "" {
		return record, fmt.Errorf("missing required field: CALL")
	}

	// Auto-detect band from frequency if band is not specified
	if record.Band == "" && record.Frequency > 0 {
		record.Band = frequencyToBand(record.Frequency)
	}

	// Set default values for missing fields
	if record.Date == "" {
		record.Date = time.Now().Format("2006-01-02")
	}
	if record.TimeOn == "" {
		record.TimeOn = time.Now().Format("15:04:05")
	}
	if record.TimeOff == "" {
		record.TimeOff = record.TimeOn
	}
	if record.Mode == "" {
		record.Mode = "SSB"
	}
	if record.RSTSent == "" {
		record.RSTSent = "59"
	}
	if record.RSTReceived == "" {
		record.RSTReceived = "59"
	}

	return record, nil
}

// formatDate converts ADIF date format (YYYYMMDD) to ISO format (YYYY-MM-DD)
func (p *ADIFParser) formatDate(adifDate string) string {
	if len(adifDate) == 8 {
		return fmt.Sprintf("%s-%s-%s", adifDate[:4], adifDate[4:6], adifDate[6:8])
	}
	return adifDate
}

// formatTime converts ADIF time format (HHMMSS or HHMM) to HH:MM:SS format
func (p *ADIFParser) formatTime(adifTime string) string {
	switch len(adifTime) {
	case 6: // HHMMSS
		return fmt.Sprintf("%s:%s:%s", adifTime[:2], adifTime[2:4], adifTime[4:6])
	case 4: // HHMM
		return fmt.Sprintf("%s:%s:00", adifTime[:2], adifTime[2:4])
	default:
		return adifTime
	}
}

// ConvertToContactRequest converts an ADIFRecord to a ContactRequest
func (r *ADIFRecord) ConvertToContactRequest() ContactRequest {
	return ContactRequest{
		Callsign:     r.Callsign,
		OperatorName: r.Name,
		ContactDate:  r.Date,
		TimeOn:       r.TimeOn,
		TimeOff:      r.TimeOff,
		Frequency:    r.Frequency,
		Band:         r.Band,
		Mode:         r.Mode,
		PowerWatts:   r.Power,
		RSTSent:      r.RSTSent,
		RSTReceived:  r.RSTReceived,
		QTH:          r.QTH,
		Country:      r.Country,
		GridSquare:   r.Grid,
		Comment:      r.Comment,
		Confirmed:    r.Confirmed,
	}
}
