package common

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
)

// SerializeBet converts a Bet struct into a CSV string
// Returns the CSV string representation of the bet and an error if serialization fails
// The returned string is finished with a newline character
func SerializeBet(b Bet, agency string) (string, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Create a row with all bet fields
	row := []string{
		agency,
		b.FirstName,
		b.LastName,
		b.Document,
		b.Birthdate,
		strconv.FormatUint(uint64(b.Number), 10),
	}

	// Write the row to the CSV writer
	if err := writer.Write(row); err != nil {
		return "", err
	}

	// Flush the writer to ensure all data is written to the buffer
	writer.Flush()

	csvStr := buf.String()
	return csvStr, nil
}

// ValidateServerResponse parses a CSV string response from the server
// returns nil if the response indicates success, error otherwise
func ValidateServerResponse(response string) error {
	reader := csv.NewReader(strings.NewReader(response))
	record, err := reader.Read()
	if err != nil {
		return err
	}

	if len(record) < 1 {
		return errors.New("invalid response from server")
	}

	if record[0] == "FAIL" {
		if len(record) >= 2 {
			return errors.New(record[1])
		}
		return errors.New("unknown error from server")
	}

	return nil
}
