package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestUserModelExists(t *testing.T) {
    // Skip the test if the "-short" flag is passed
	if testing.Short() {
        t.Skip("models: skipping integration tests in short mode")
    }

    // Create a new test database
	db := newTestDatabase(t)
    
	// Create a new UserModel instance
	m := UserModel{Client: db}

    // Clean up before and after tests
    cleanupTestUsers(t, db)
    t.Cleanup(func() {
        cleanupTestUsers(t, db)
    })

	// Insert a test user
    testUser, err := InsertTestUser(db, "Jane Doe", "jane.doe@example.com", "password123")
    if err != nil {
        t.Fatalf("Failed to insert test user: %v", err)
    }

	// Parse the returned string ID directly
	testUserID, err := uuid.Parse(testUser)
    if err != nil {
        t.Fatalf("Failed to parse test user ID: %v", err)
    }

    // Create an invalid UUID
	invalidUUID := uuid.New()

    // Define test cases
	tests := []struct {
        name   string
        userID uuid.UUID
        want   bool
    }{
        {"Valid ID", testUserID, true},
        {"Invalid ID", invalidUUID, false},
    }

    // Loop through each test case
	for _, tt := range tests {
        // Run each test case
		t.Run(tt.name, func(t *testing.T) {
            // Check if the user exists
			exists, err := m.Exists(tt.userID)
            if err != nil {
                t.Fatalf("Error in Exists method: %v", err)
            }
            if exists != tt.want {
                t.Errorf("Test case %s failed: got %v, want %v", tt.name, exists, tt.want)
            }
        })
    }
}