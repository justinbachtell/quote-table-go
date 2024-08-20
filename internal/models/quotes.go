package models

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

// Define an interface for the QuoteModel
type QuoteModelInterface interface {
	Insert(quote string, author string, userID uuid.UUID) (int, error)
	Get(id int) (Quote, error)
	GetByUserID(userID uuid.UUID) ([]Quote, error)
	Update(id int, quote string, author string) (int, error)
	Latest() ([]Quote, error)
	Exists(id int) (bool, error)
	Delete(id int) error
	SetAuthUserID(id uuid.UUID)
}

// Define a Quote struct to hold the quote data
type Quote struct {
	ID      int       `json:"id"`
	Quote   string    `json:"quote"`
	Author  string    `json:"author"`
	UserID  uuid.UUID `json:"user_id"`
	Created time.Time `json:"created"`
}

// Define a QuoteModel struct to hold the database connection pool
type QuoteModel struct {
	Client *supabase.Client
	AuthUserID uuid.UUID
}

// Set the authenticated user ID
func (m *QuoteModel) SetAuthUserID(id uuid.UUID) {
	m.AuthUserID = id
}

// Return a specific quote based on the ID
func (m *QuoteModel) Get(id int) (Quote, error) {
	// Initialize a new Quote struct to hold the data
	var q Quote

	// Convert id to string
	idStr := strconv.Itoa(id)

	// Query the database for the quote
	count, err := m.Client.From("quotes").Select("*", "exact", false).Eq("id", idStr).Single().ExecuteTo(&q)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Quote{}, ErrNoRecord
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for ID: %d", id)
		return Quote{}, err
	} else if count == 0 {
		log.Printf("No record found for ID: %d", id)
		return Quote{}, ErrNoRecord
	}

	// Return the Quote struct
	log.Printf("Retrieved quote: %+v", q)
	return q, nil
}

// Return a list of quotes by user ID
func (m *QuoteModel) GetByUserID(userID uuid.UUID) ([]Quote, error) {
	var quotes []Quote
	
	_, err := m.Client.From("quotes").Select("*", "exact", false).Eq("user_id", userID.String()).Limit(10, "").ExecuteTo(&quotes)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	return quotes, nil
}

// Return a list of the 10 most recent quotes
func (m *QuoteModel) Latest() ([]Quote, error) {
	var quotes []Quote
	
	_, err := m.Client.From("quotes").Select("*", "exact", false).Limit(10, "").ExecuteTo(&quotes)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	// Return the slice of quotes
	// log.Printf("Retrieved quotes: %+v", quotes)
	return quotes, nil
}

// Insert a new quote into the database
func (m *QuoteModel) Insert(quote string, author string, userID uuid.UUID) (int, error) {
	// Create a map to hold the quote data
	data := map[string]interface{}{
		"quote":   quote,
		"author":  author,
		"created": time.Now(),
		"user_id": userID,
	}

	// Insert the quote into the database
	response, count, err := m.Client.From("quotes").Insert(data, false, "", "", "").ExecuteString()
	if err != nil {
		log.Printf("Error inserting quote: %v", err)
		return 0, err
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for insert")
		return 0, err
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

// Update a quote in the database
func (m *QuoteModel) Update(id int, quote string, author string) (int, error) {
	// Create a map to hold the quote data
	data := map[string]interface{}{
		"quote":   quote,
		"author":  author,
		"user_id": m.AuthUserID,
	}

	// Convert id to string
	idStr := strconv.Itoa(id)

	// Update the quote in the database
	response, _, err := m.Client.From("quotes").Update(data, "", "exact").Eq("id", idStr).Execute()
	if err != nil {
		log.Printf("Error updating quote: %v", err)
		return 0, err
	}
	
	// Parse the JSON response
	var updatedQuote []Quote
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&updatedQuote)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return 0, err
	}

	// Check if the quote was successfully updated
	if len(updatedQuote) == 0 {
		log.Printf("No quotes returned in response")
		return 0, errors.New("no quotes returned in response")
	}

	return int(updatedQuote[0].ID), nil
}

// Check if the quote exists
func (m *QuoteModel) Exists(id int) (bool, error) {
    // Convert id to string
    idStr := strconv.Itoa(id)

    // Query the database for the user with the given id
    response, count, err := m.Client.From("quotes").Select("id", "exact", false).Eq("id", idStr).ExecuteString()
    if err != nil {
        if strings.Contains(err.Error(), "PGRST116") {
            // No rows returned
            return false, nil
        }
        return false, err
    }

    // If count is 0, user doesn't exist
    if count == 0 {
        return false, nil
    }

    // Parse the JSON response
    var quotes []struct {
        ID int `json:"id"`
    }
    err = json.NewDecoder(strings.NewReader(string(response))).Decode(&quotes)
    if err != nil {
        return false, err
    }

    // Check if any user was found
    return len(quotes) > 0, nil
}

// Delete a quote
func (m *QuoteModel) Delete(id int) error {
	// Convert id to string
	idStr := strconv.Itoa(id)

	// Delete the quote from the database
	_, _, err := m.Client.From("quotes").Delete("", "exact").Eq("id", idStr).Execute()
	return err
}