package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestBookModelExists(t *testing.T) {
	// Skip the test if the "-short" flag is passed
	if testing.Short() {
		t.Skip("models: skipping integration tests in short mode")
	}

	// Set up a tests struct
	tests := []struct {
		name string
		bookID int
		want bool
	}{
		{"Valid ID", -1, true}, // We'll replace -1 with the actual ID
		{"Non-existent ID", 9999, false},
		{"Zero ID", 0, false},
		{"Negative ID", -1, false},
	}

	// Create a new test database once for all tests
	db := newTestDatabase(t)

	// Create a new BookModel instance
	m := BookModel{Client: db}

	// Insert a test book and get the ID
	testBookID, err := InsertTestBook(db, "Test Book")
	if err != nil {
		t.Fatalf("Failed to insert test book: %v", err)
	}

	// Replace the -1 in the "Valid ID" test with the actual ID
	tests[0].bookID = testBookID

	// Loop through each test
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the Exists method to check if the book exists
			exists, err := m.Exists(tt.bookID)

			// Check for error first
			if err != nil {
				t.Fatalf("Error in Exists method: %v", err)
			}

			// Check if the result matches the expected value
			if exists != tt.want {
				t.Errorf("Test case %s failed: got %v, want %v", tt.name, exists, tt.want)
			}
		})
	}
}

func TestBookModelSetAuthUserID(t *testing.T) {
	// Skip the test if the "-short" flag is passed
	if testing.Short() {
		t.Skip("models: skipping integration tests in short mode")
	}

	// Create a new test database
	db := newTestDatabase(t)

	// Create a new BookModel instance
	m := BookModel{Client: db}

	// Create a new user
	userID := uuid.New()

	// Test cases
	testCases := []struct {
		name     string
		userID   uuid.UUID
		expected uuid.UUID
	}{
		{"Set valid user ID", userID, userID},
		{"Set empty user ID", uuid.Nil, uuid.Nil},
	}

	// Loop through each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the AuthUserID
			m.SetAuthUserID(tc.userID)

			// Check if the AuthUserID was set correctly
			if m.AuthUserID != tc.expected {
				t.Errorf("Expected AuthUserID to be %s, but got %s", tc.expected, m.AuthUserID)
			}
		})
	}
}