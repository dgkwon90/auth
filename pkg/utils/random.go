package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(b)[:n]
}
