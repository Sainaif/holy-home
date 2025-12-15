package utils

import (
	"testing"
)

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Too short", "Short1!", true},
		{"No uppercase", "longpassword1!", true},
		{"No lowercase", "LONGPASSWORD1!", true},
		{"No number", "LongPassword!", true},
		{"No special", "LongPassword1", true},
		{"Valid password", "LongPassword1!", false},
		{"Valid password with spaces", "Long Password 1!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordStrength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
