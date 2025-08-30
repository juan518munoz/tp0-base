package common

import (
	"encoding/json"
)

// SerializeBet converts a Bet struct into a JSON string
// Returns the JSON string representation of the bet and an error if serialization fails
func SerializeBet(b Bet) (string, error) {
	betMap := map[string]interface{}{
		"name":      b.Name,
		"surname":   b.Surname,
		"id":        b.Id,
		"birthdate": b.Birthdate,
		"number":    b.Number,
	}

	jsonBytes, err := json.Marshal(betMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
