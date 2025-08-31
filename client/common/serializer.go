package common

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strconv"
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

// ValidateServerResponse parses a JSON string response from the server
// returns true if the response indicates success, false otherwise
func ValidateServerResponse(response string) error {
	var parsedResponse map[string]interface{}
	err := json.Unmarshal([]byte(response), &parsedResponse)
	if err != nil {
		return err
	}

	if parsedResponse["status"] != "OK" {
		if reason, ok := parsedResponse["reason"].(string); ok {
			return errors.New(reason)
		}
		return errors.New("unknown error")
	}

	return nil
}
