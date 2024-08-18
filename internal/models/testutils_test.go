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
	supabaseURL := "http://127.0.0.1:54321"
	supabaseKey := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0"

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
func InsertTestUser(db *supabase.Client, name, email, password string) (int, error) {
	// Hash the password with the number of specified salt rounds
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12) // Default cost is 12
	if err != nil {
		return 0, err
	}

	// Create a map to hold the user data
	data := map[string]interface{}{
		"name":            name,
		"email":           email,
		"hashed_password": hashedPassword,
		"created":         time.Now(),
	}

	// Insert the user into the database and get the response
	response, count, err := db.From("users").Insert(data, false, "", "", "").ExecuteString()
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			// User already exists, fetch the existing user's ID
			existingUser, _, err := db.From("users").Select("id", "exact", false).Eq("email", email).Single().ExecuteString()
			if err != nil {
				return 0, err
			}
			var user struct {
				ID int `json:"id"`
			}
			err = json.Unmarshal([]byte(existingUser), &user)
			if err != nil {
				return 0, err
			}
			return user.ID, nil
		}
		return 0, err
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for insert")
		return 0, errors.New("unexpected count > 1 for insert")
	}
	
	// Parse the JSON response to extract the ID
	var insertedUser []User
	err = json.Unmarshal([]byte(response), &insertedUser)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return 0, err
	}

	// Check if the user was successfully inserted
	if len(insertedUser) == 0 {
		log.Printf("No users returned in response")
		return 0, errors.New("no users returned in response")
	}

	return insertedUser[0].ID, nil
}

// Insert a new quote into the database
func InsertTestQuote(db *supabase.Client, quote string, author string) (int, error) {
	// Create a map to hold the quote data
	data := map[string]interface{}{
		"quote":   quote,
		"author":  author,
		"created": time.Now(),
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
	err = json.Unmarshal([]byte(response), &insertedQuote)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return 0, err
	}

	// Check if the quote was successfully inserted
	if len(insertedQuote) == 0 {
		log.Printf("No quotes returned in response")
		return 0, errors.New("no quotes returned in response")
	}

	return insertedQuote[0].ID, nil
}