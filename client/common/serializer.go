package common

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// SerializeBatchBets converts a slice of Bet structs into the batch format
// Format: "agency,count\n" followed by bet lines
// Returns the serialized batch string and an error if serialization fails
func SerializeBatchBets(bets []Bet, agency string) (string, error) {
	if len(bets) == 0 {
		return "", errors.New("cannot serialize empty batch")
	}

	var buf bytes.Buffer

	// Write header line with agency ID and bet count
	header := fmt.Sprintf("BETS,%s,%d\n", agency, len(bets))
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

// ValidateBetsServerResponse parses a CSV string response from the server
// returns nil if the response indicates success, error otherwise
func ValidateBetsServerResponse(response string) error {
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

// ValidateFinishedNotificationServerResponse parses a CSV string
// response from the server
// returns nil if the response indicates success, error otherwise
func ValidateFinishedNotificationServerResponse(response string) error {
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

// GetResultsServerResponse parses a CSV string response from the server
// if there are agencies still pending to send their results, returns false
// otherwise returns true and the amount of bets won
func GetResultsServerResponse(response string) (bool, int, error) {
	// OK,<count>
	// FAIL,NOT_READY
	// FAIL,<other_error>
	reader := csv.NewReader(strings.NewReader(response))
	record, err := reader.Read()
	if err != nil {
		return false, 0, err
	}

	if len(record) < 1 {
		return false, 0, errors.New("invalid response from server")
	}

	if record[0] == "FAIL" {
		if len(record) >= 2 {
			if record[1] == "NOT_READY" {
				return false, 0, nil
			}
			return false, 0, errors.New(record[1])
		}
		return false, 0, errors.New("unknown error from server")
	}

	if record[0] != "OK" || len(record) < 2 {
		return false, 0, errors.New("invalid response from server")
	}

	count, err := strconv.Atoi(record[1])
	if err != nil {
		return false, 0, errors.New("invalid count in server response")
	}
	return true, count, nil
}
