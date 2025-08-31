package common

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
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

// SerializeBatchBets converts a slice of Bet structs into the batch format
// Format: "agency,count\n" followed by bet lines
// Returns the serialized batch string and an error if serialization fails
func SerializeBatchBets(bets []Bet, agency string) (string, error) {
	if len(bets) == 0 {
		return "", errors.New("cannot serialize empty batch")
	}

	var buf bytes.Buffer

	// Write header line with agency ID and bet count
	header := fmt.Sprintf("%s,%d\n", agency, len(bets))
	buf.WriteString(header)

	// Write each bet in the format: "FirstName,LastName,Document,Birthdate,Number\n"
	for _, bet := range bets {
		betLine := fmt.Sprintf("%s,%s,%s,%s,%d\n",
			bet.FirstName,
			bet.LastName,
			bet.Document,
			bet.Birthdate,
			bet.Number)
		buf.WriteString(betLine)
	}

	buf.WriteByte(0) // Null byte to indicate end of batch
	return buf.String(), nil
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
