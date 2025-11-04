package services

import (
	"testing"

	"github.com/sainaif/holy-home/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid password",
			password: "SecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  false, // HashPassword should still work with empty string
		},
		{
			name:     "Long password",
			password: "ThisIsAVeryLongPasswordThatShouldStillBeHashedCorrectly123456789!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "Unicode password",
			password: "Zażółć gęślą jaźń 🔒",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hash the password
			hashedPassword, err := utils.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify the password
			valid, err := utils.VerifyPassword(tt.password, hashedPassword)
			assert.NoError(t, err)
			assert.True(t, valid, "Password verification should succeed for correct password")

			// Verify with wrong password
			valid, err = utils.VerifyPassword("wrongpassword", hashedPassword)
			assert.NoError(t, err)
			assert.False(t, valid, "Password verification should fail for incorrect password")

			// Ensure each hash is unique (salt is working)
			hashedPassword2, err := utils.HashPassword(tt.password)
			assert.NoError(t, err)
			assert.NotEqual(t, hashedPassword, hashedPassword2, "Each hash should be unique due to salt")

			// But both should verify correctly
			valid, err = utils.VerifyPassword(tt.password, hashedPassword2)
			assert.NoError(t, err)
			assert.True(t, valid)
		})
	}
}

func TestPasswordVerificationWithInvalidHash(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "Empty hash",
			hash:    "",
			wantErr: true,
		},
		{
			name:    "Invalid hash format",
			hash:    "not-a-valid-argon2-hash",
			wantErr: true,
		},
		{
			name:    "Malformed hash",
			hash:    "$argon2id$v=19$m=65536,t=3,p=2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := utils.VerifyPassword("password123", tt.hash)

			if tt.wantErr {
				assert.Error(t, err, "Should return error for invalid hash")
			}

			assert.False(t, valid, "Should not validate with invalid hash")
		})
	}
}
