package utils

import "strings"

func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts[0]) <= 2 {
		return "***@" + parts[1]
	}
	return parts[0][:2] + strings.Repeat("*", len(parts[0])-2) + "@" + parts[1]
}
