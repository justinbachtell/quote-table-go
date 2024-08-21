package models

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"golang.org/x/crypto/bcrypt"
)

// Opens a new local database connection for the test suite
func newTestDatabase(t *testing.T) *supabase.Client {
	// Create a new logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Log the environment variables to check if they are loaded correctly
	supabaseURL := "http://127.0.0.1:54321" // local supabase url
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0" // local supabase key

	// Check if the environment variables are set
	if supabaseURL == "" || supabaseKey == "" {
		logger.Error("SUPABASE_URL or SUPABASE_PUBLIC_KEY is not set")
		t.Fatal("SUPABASE_URL or SUPABASE_PUBLIC_KEY is not set")
	}

	// Initialize supabase client
	db, err := connectSupabase(logger, supabaseURL, supabaseKey)
	if err != nil {
		logger.Error(err.Error())
		t.Fatal(err)
	} else {
		logger.Info("connected to test database")
	}

	// Insert a test user
	_, err = InsertTestUser(db, "John Doe", "john.doe@example.com", "$2a$12$NuTjWXm3KKntReFwyBVH")
	if err != nil {
		logger.Error("Failed to insert or retrieve test user: " + err.Error())
		t.Fatal(err)
	}

	// Insert a test quote
	_, err = InsertTestQuote(db, "The quick brown fox jumps over the lazy dog", "Thomas A. Edison")
	if err != nil {
		logger.Error("Failed to insert or retrieve test quote: " + err.Error())
		t.Fatal(err)
	}

	return db
}

// Deletes all users from the database
func cleanupTestUsers(t *testing.T, db *supabase.Client) {
    t.Helper()
    _, _, err := db.From("users").Delete("", "exact").Eq("email", "john.doe@example.com").Execute()
    if err != nil {
        t.Fatalf("Failed to clean up test user: %v", err)
    }

	_, _, err = db.From("users").Delete("", "exact").Eq("email", "jane.doe@example.com").Execute()
	if err != nil {
		t.Fatalf("Failed to clean up test user: %v", err)
	}
}

// Connect to the supabase database
func connectSupabase(logger *slog.Logger, supabaseURL string, supabaseKey string) (*supabase.Client, error) {
	// Initialize supabase client
	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		logger.Error("error connecting to the database")
		return nil, err
	}

    // Return the connection is successful
    return client, nil
}

// Insert adds a new user to the database
func InsertTestUser(db *supabase.Client, name, email, password string) (string, error) {
	// Check if user already exists
    existingUser, count, err := db.From("users").Select("id", "exact", false).Eq("email", email).Single().ExecuteString()
    if err == nil && count > 0 {
        var user struct {
            ID uuid.UUID `json:"id"`
        }
        err = json.NewDecoder(strings.NewReader(existingUser)).Decode(&user)
        if err != nil {
            return "", err
        }
        return user.ID.String(), nil
    }

	// Hash the password with the number of specified salt rounds
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12) // Default cost is 12
	if err != nil {
		return "", err
	}

    // Create a new UUID
	newUUID := uuid.New()
    
	// Create a map to hold the user data
	data := map[string]interface{}{
        "id":              newUUID,
        "name":            name,
        "email":           email,
        "hashed_password": hashedPassword,
		"profile_slug":    strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-"),
        "created_at":         time.Now(),
		"updated_at":         time.Now(),
    }

    // Insert the user into the database
	response, count, err := db.From("users").Insert(data, false, "", "", "").ExecuteString()
    if err != nil {
        return "", err
    }

    if len(data) == 0 {
		return "", errors.New("no user was inserted")
	}

    // Parse the JSON response to extract the ID
    var insertedUser []struct {
        ID uuid.UUID `json:"id"`
    }

	// Decode the JSON response
    err = json.NewDecoder(strings.NewReader(response)).Decode(&insertedUser)
    if err != nil {
        return "", err
    }

    // Check if the user was successfully inserted
	if len(insertedUser) == 0 {
        return "", errors.New("no user ID returned")
    }

    return insertedUser[0].ID.String(), nil
}

// Insert a new quote into the database
func InsertTestQuote(db *supabase.Client, quote string, author string) (int, error) {
	// Create a map to hold the quote data
	data := map[string]interface{}{
		"quote":   quote,
		"author":  author,
		"created_at": time.Now(),
	}

	// Insert the quote into the database
	response, count, err := db.From("quotes").Insert(data, false, "", "", "").ExecuteString()
	if err != nil {
		log.Printf("Error inserting quote: %v", err)
		return 0, err
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for insert")
		return 0, errors.New("unexpected count > 1 for insert")
	}

	// Parse the JSON response to extract the ID
	var insertedQuote []Quote
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&insertedQuote)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return 0, err
	}

	// Check if the quote was successfully inserted
	if len(insertedQuote) == 0 {
		log.Printf("No quotes returned in response")
		return 0, errors.New("no quotes returned in response")
	}

	return int(insertedQuote[0].ID), nil
}