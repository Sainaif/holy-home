package sqlite

import (
	"encoding/hex"

	"github.com/google/uuid"
)

// bytesToHex converts byte slice to hex string for storage
func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// hexToBytes converts hex string back to bytes
func hexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// generateID creates a new UUID-based ID
func generateID() string {
	return uuid.New().String()
}

// NullString helper for nullable strings
type NullString struct {
	Value string
	Valid bool
}

func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.Value, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	ns.Value = value.(string)
	return nil
}

// StringPtr returns a pointer to the string, or nil if empty
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// PtrString returns the string value or empty string if nil
func PtrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
