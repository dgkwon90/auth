// Package utils provides utility functions for the authentication service.
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString generates a random string of length n.
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)[:n]
}
