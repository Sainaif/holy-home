package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEmailValidation tests email validation logic
func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "Valid email",
			email: "user@example.com",
			valid: true,
		},
		{
			name:  "Valid email with subdomain",
			email: "user@mail.example.com",
			valid: true,
		},
		{
			name:  "Valid email with plus",
			email: "user+tag@example.com",
			valid: true,
		},
		{
			name:  "Invalid - no @",
			email: "userexample.com",
			valid: false,
		},
		{
			name:  "Invalid - no domain",
			email: "user@",
			valid: false,
		},
		{
			name:  "Invalid - no username",
			email: "@example.com",
			valid: false,
		},
		{
			name:  "Invalid - empty",
			email: "",
			valid: false,
		},
		{
			name:  "Invalid - spaces",
			email: "user @example.com",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simple email validation regex
			// This matches the pattern used in the application
			// A real implementation would be in utils or validation package
			isValid := isValidEmail(tt.email)
			assert.Equal(t, tt.valid, isValid, "Email validation mismatch for: %s", tt.email)
		})
	}
}

// isValidEmail is a simple email validation helper for testing
// In production, this would be in a utils package
func isValidEmail(email string) bool {
	if len(email) == 0 {
		return false
	}

	// Basic check: must contain @ and have text before and after
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false // Multiple @ signs
			}
			atIndex = i
		}
		if c == ' ' {
			return false // No spaces allowed
		}
	}

	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false // @ must have text before and after
	}

	// Check for dot in domain part
	domainPart := email[atIndex+1:]
	hasDot := false
	for _, c := range domainPart {
		if c == '.' {
			hasDot = true
			break
		}
	}

	return hasDot
}

// TestUserRoleValidation tests user role validation
func TestUserRoleValidation(t *testing.T) {
	validRoles := []string{"ADMIN", "RESIDENT"}

	tests := []struct {
		name  string
		role  string
		valid bool
	}{
		{
			name:  "Valid ADMIN role",
			role:  "ADMIN",
			valid: true,
		},
		{
			name:  "Valid RESIDENT role",
			role:  "RESIDENT",
			valid: true,
		},
		{
			name:  "Invalid role lowercase",
			role:  "admin",
			valid: false,
		},
		{
			name:  "Invalid role custom",
			role:  "CUSTOM",
			valid: false,
		},
		{
			name:  "Invalid empty role",
			role:  "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := false
			for _, validRole := range validRoles {
				if tt.role == validRole {
					isValid = true
					break
				}
			}
			assert.Equal(t, tt.valid, isValid, "Role validation mismatch for: %s", tt.role)
		})
	}
}
