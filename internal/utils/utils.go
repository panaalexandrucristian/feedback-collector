package utils

import (
	"regexp"
	"strings"
)

// SanitizeInput removes potentially harmful characters from a string
func SanitizeInput(input string) string {
	// Basic sanitization - remove HTML tags
	re := regexp.MustCompile("<[^>]*>")
	sanitized := re.ReplaceAllString(input, "")

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// ValidateRoomID checks if a room ID follows the expected format
func ValidateRoomID(roomID string) bool {
	// Room IDs are 6 characters long, alphanumeric
	match, _ := regexp.MatchString("^[a-zA-Z0-9]{6}$", roomID)
	return match
}
