package common

import (
	"encoding/json"
	"errors"
)

// SerializeBet converts a Bet struct into a JSON string
// Returns the JSON string representation of the bet and an error if serialization fails
func SerializeBet(b Bet, agency string) (string, error) {
	betMap := map[string]interface{}{
		"agency":    agency,
		"firstName": b.FirstName,
		"lastName":  b.LastName,
		"document":  b.Document,
		"birthdate": b.Birthdate,
		"number":    b.Number,
	}

	jsonBytes, err := json.Marshal(betMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
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
