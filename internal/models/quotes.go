package models

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
)

// Define an interface for the QuoteModel
type QuoteModelInterface interface {
	Insert(quote string, authorID int, bookID int, pageNumber string, isPrivate bool, userID uuid.UUID) (int, error)
	Get(id int) (Quote, error)
	GetByAuthorID(authorID int) ([]Quote, error)
	GetByUserID(userID uuid.UUID) ([]Quote, error)
	GetWithAuthorAndBook(id int) (Quote, error)
	Update(id int, quote string, authorID int, bookID int, pageNumber string, isPrivate bool) (int, error)
	Latest() ([]Quote, error)
	Exists(id int) (bool, error)
	Delete(id int) error
	SetAuthUserID(id uuid.UUID)
	GetByBookID(bookID int) ([]Quote, error)
}

// Define a Quote struct to hold the quote data
type Quote struct {
	ID      int       `json:"id"`
	Quote   string    `json:"quote"`
	AuthorID  int    `json:"author_id"`
	Author    Author `json:"author"`
	BookID    int    `json:"book_id"`
	Book      Book   `json:"book"`
	UserID  uuid.UUID `json:"user_id"`
	PageNumber string `json:"page_number"`
	IsPrivate bool `json:"is_private"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Define a QuoteModel struct to hold the database connection pool
type QuoteModel struct {
	Client *supabase.Client
	AuthClient *supabase.Client
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

// Return a list of quotes by author ID
func (m *QuoteModel) GetByAuthorID(authorID int) ([]Quote, error) {
	var quotes []Quote
	
	_, err := m.Client.From("quotes").Select("*", "exact", false).Eq("author_id", strconv.Itoa(authorID)).ExecuteTo(&quotes)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	return quotes, nil
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

// Return a quote with the author and book
func (m *QuoteModel) GetWithAuthorAndBook(id int) (Quote, error) {
	var q Quote
	var a Author
	var b Book

	// Query the database for the quote and join with the author
	_, err := m.Client.From("quotes").Select("*", "exact", false).Eq("id", strconv.Itoa(id)).Single().ExecuteTo(&q)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Quote{}, err
	}

	// Query the database for the author
	_, err = m.Client.From("authors").Select("*", "exact", false).Eq("id", strconv.Itoa(q.AuthorID)).Single().ExecuteTo(&a)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Quote{}, err
	}

	// Query the database for the book
	_, err = m.Client.From("books").Select("*", "exact", false).Eq("id", strconv.Itoa(q.BookID)).Single().ExecuteTo(&b)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Quote{}, err
	}

	// Return the quote with the author and book	
	q.Author = a
	q.Book = b

	return q, nil
}

// Return a list of the 10 most recent quotes
func (m *QuoteModel) Latest() ([]Quote, error) {
	var quotes []Quote
	
	// Query the database for the quotes with authors
	_, err := m.Client.From("quotes").Select("*", "exact", false).Order("created_at", &postgrest.OrderOpts{Ascending: true}).Limit(10, "").ExecuteTo(&quotes)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	// Fetch authors for each quote
	for i := range quotes {
		var author Author
		_, err := m.Client.From("authors").Select("*", "exact", false).Eq("id", strconv.Itoa(quotes[i].AuthorID)).Single().ExecuteTo(&author)
		if err != nil {
			log.Printf("Error fetching author for quote %d: %v", quotes[i].ID, err)
			continue
		}
		quotes[i].Author = author
	}

	// Fetch books for each quote
	for i := range quotes {
		var book Book
		_, err := m.Client.From("books").Select("*", "exact", false).Eq("id", strconv.Itoa(quotes[i].BookID)).Single().ExecuteTo(&book)
		if err != nil {
			log.Printf("Error fetching book for quote %d: %v", quotes[i].ID, err)
			continue
		}
		quotes[i].Book = book
	}

	// Return the slice of quotes
	return quotes, nil
}

// Insert a new quote into the database
func (m *QuoteModel) Insert(quote string, authorID int, bookID int, pageNumber string, isPrivate bool, userID uuid.UUID) (int, error) {
	// Verify the user exists
	_, _, err := m.AuthClient.From("users").Select("id", "exact", false).Eq("id", userID.String()).ExecuteString()
	if err != nil {
		log.Printf("Error verifying user exists to insert quote: %v", err)
		return 0, err
	}

	// Create a map to hold the quote data
	data := map[string]interface{}{
		"quote":   quote,
		"author_id":  authorID,
		"book_id": bookID,
		"page_number": pageNumber,
		"is_private": isPrivate,
		"created_at": time.Now(),
		"updated_at": time.Now(),
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

	// Update the user's last quote added at timestamp
	_, _, err = m.AuthClient.From("users").Update(map[string]interface{}{"last_quote_added_at": time.Now()}, "", "").Eq("id", userID.String()).Execute()
	if err != nil {
		log.Printf("Error updating user's last quote added at timestamp: %v", err)
		return 0, err
	}

	// Return the ID of the inserted quote
	return int(insertedQuote[0].ID), nil
}

// Update a quote in the database
func (m *QuoteModel) Update(id int, quote string, authorID int, bookID int, pageNumber string, isPrivate bool) (int, error) {
	// Create a map to hold the quote data
	data := map[string]interface{}{
		"quote":   quote,
		"author_id":  authorID,
		"book_id": bookID,
		"page_number": pageNumber,
		"is_private": isPrivate,
		"user_id": m.AuthUserID,
		"updated_at": time.Now(),
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

	// Decode the JSON response
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

// GetByBookID returns all quotes for a given book ID
func (m *QuoteModel) GetByBookID(bookID int) ([]Quote, error) {
    var quotes []Quote
    
    response, count, err := m.Client.From("quotes").Select("*", "exact", false).Eq("book_id", strconv.Itoa(bookID)).ExecuteString()
    if err != nil {
        log.Printf("Error fetching quotes for book %d: %v", bookID, err)
        return nil, err
    }

    if count == 0 {
        return []Quote{}, nil // Return an empty slice if no quotes found
    }

    err = json.NewDecoder(strings.NewReader(response)).Decode(&quotes)
    if err != nil {
        log.Printf("Error decoding quotes JSON: %v", err)
        return nil, err
    }

    return quotes, nil
}