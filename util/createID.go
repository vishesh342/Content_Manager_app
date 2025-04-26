package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GeneratePostID generates a unique post ID based on the username, account ID, and scheduled time.
func GeneratePostID(accountID string, scheduledTime time.Time) string {
	// Create a delimiter-separated string
	input := fmt.Sprintf("%s|%s", accountID, scheduledTime.UTC().Format(time.RFC3339Nano))
	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])[:16]
}