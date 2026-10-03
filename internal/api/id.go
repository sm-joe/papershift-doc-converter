package api

import (
	"crypto/rand"
	"encoding/hex"
)

func newJobID() (string, error) {
	buffer := make([]byte, 16)

	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}
