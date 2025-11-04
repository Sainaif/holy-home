package internal

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockDatabase is a mock MongoDB database for testing
type MockDatabase struct {
	mock.Mock
}

// Collection returns a mock collection
func (m *MockDatabase) Collection(name string, opts ...interface{}) *mongo.Collection {
	args := m.Called(name, opts)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*mongo.Collection)
}

// MockCollection is a mock MongoDB collection for testing
type MockCollection struct {
	mock.Mock
}

// TestObjectID creates a deterministic ObjectID for testing
func TestObjectID(value string) primitive.ObjectID {
	// Create a 24-character hex string padded from the input
	hexStr := value
	for len(hexStr) < 24 {
		hexStr += "0"
	}
	if len(hexStr) > 24 {
		hexStr = hexStr[:24]
	}

	objID, _ := primitive.ObjectIDFromHex(hexStr)
	return objID
}

// Float64Ptr returns a pointer to a float64 value
func Float64Ptr(v float64) *float64 {
	return &v
}

// IntPtr returns a pointer to an int value
func IntPtr(v int) *int {
	return &v
}

// StringPtr returns a pointer to a string value
func StringPtr(v string) *string {
	return &v
}

// AssertNoError is a helper to assert no error occurred
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// AssertError is a helper to assert an error occurred
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error but got nil", msg)
	}
}
