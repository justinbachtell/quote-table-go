package models

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

// Define an interface for the AuthorModel
type AuthorModelInterface interface {
	Insert(name string) (int, error)
	Get(id int) (Author, error)
	GetBooksByAuthor(authorID int) ([]Book, error)
	GetQuotesByAuthor(authorID int) ([]Quote, error)
	GetWithCounts(id int) (Author, error)
	GetByName(name string) (Author, error)
	Update(id int, name string) (int, error)
	Delete(id int) error
	Exists(id int) (bool, error)
	GetAll() ([]Author, error)
	GetAllWithCounts() ([]AuthorWithCounts, error)
	SetAuthUserID(id uuid.UUID)
}

// Author represents an author in the database
type Author struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	UserID uuid.UUID `json:"user_id"`
	QuoteCount int `json:"quote_count"`
	BookCount int `json:"book_count"`
	Books []Book `json:"books"`
	Quotes []Quote `json:"quotes"`
}

// The model used in the connection pool
type AuthorModel struct {
	Client *supabase.Client
	AuthUserID uuid.UUID
}

// Insert adds a new author to the database
func (m *AuthorModel) Insert(name string) (int, error) {
	// Create a map to hold the author data
	data := map[string]interface{}{
		"name": name,
	}

	// Insert the author into the database
	response, _, err := m.Client.From("authors").Insert(data, false, "", "", "").ExecuteString()
	if err != nil {
		return 0, err
	}

	// Parse the JSON response to extract the ID
	var insertedAuthor []Author
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&insertedAuthor)
	if err != nil {
		return 0, err
	}

	// Check if the author was successfully inserted
	if len(insertedAuthor) == 0 {
		return 0, errors.New("no authors returned in response")
	}

	// Return the ID of the inserted author
	return insertedAuthor[0].ID, nil
}

// Get a single author by ID
func (m *AuthorModel) Get(id int) (Author, error) {
	// Initialize a new Author struct to hold the data
	var a Author

	// Convert id to string
	idStr := strconv.Itoa(id)

	// Query the database for the author
	count, err := m.Client.From("authors").Select("*", "exact", false).Eq("id", idStr).Single().ExecuteTo(&a)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Author{}, ErrNoRecord
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for ID: %d", id)
		return Author{}, err
	} else if count == 0 {
		log.Printf("No record found for ID: %d", id)
		return Author{}, ErrNoRecord
	}

	// Return the Author struct
	return a, nil
}

// Get a single author by name
func (m *AuthorModel) GetByName(name string) (Author, error) {
	// Initialize a new Author struct to hold the data
	var a Author

	// Query the database for the author
	count, err := m.Client.From("authors").Select("*", "exact", false).Eq("name", name).Single().ExecuteTo(&a)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Author{}, ErrNoRecord
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for name: %s", name)
		return Author{}, err
	} else if count == 0 {
		log.Printf("No record found for name: %s", name)
		return Author{}, ErrNoRecord
	}

	// Return the Author struct
	return a, nil
}

// Get books by author
func (m *AuthorModel) GetBooksByAuthor(authorID int) ([]Book, error) {
	// Initialize a new Book slice to hold the data
	var books []Book

	// Convert authorID to string
	authorIDStr := strconv.Itoa(authorID)

	// Query the database for the books by author
	response, count, err := m.Client.From("books").Select("*", "exact", false).ExecuteString()
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return []Book{}, err
	}

	// If no books were found, return an empty slice
	if count == 0 {
		log.Printf("No books found for author ID: %d", authorID)
		return []Book{}, nil
	}

	// Parse the JSON response
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&books)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return []Book{}, err
	}

	// Query the quotes to get the authorID for each book
	quotesResponse, _, err := m.Client.From("quotes").Select("*", "exact", false).Eq("author_id", authorIDStr).ExecuteString()
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return []Book{}, err
	}

	// Parse the JSON response
	var quotes []Quote
	err = json.NewDecoder(strings.NewReader(string(quotesResponse))).Decode(&quotes)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return []Book{}, err
	}

	// Filter the books by the authorID
	for _, quote := range quotes {
		for _, book := range books {
			if quote.BookID == book.ID {
				// Filter the books by the authorID
				books = append(books, book)
				break
			}
		}
	}

	// Return the books
	return books, nil
}

// Get quotes by author
func (m *AuthorModel) GetQuotesByAuthor(authorID int) ([]Quote, error) {
	// Initialize a new Quote slice to hold the data
	var quotes []Quote

	// Convert authorID to string
	authorIDStr := strconv.Itoa(authorID)

	// Query the database for the quotes by author
	response, count, err := m.Client.From("quotes").Select("*", "exact", false).Eq("author_id", authorIDStr).ExecuteString()
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return []Quote{}, err
	}

	// If no quotes were found, return an empty slice
	if count == 0 {
		log.Printf("No quotes found for author ID: %d", authorID)
		return []Quote{}, nil
	}

	// Parse the JSON response
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&quotes)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return []Quote{}, err
	}

	// Return the quotes
	return quotes, nil
}

// Update an author by ID
func (m *AuthorModel) Update(id int, name string) (int, error) {
	// Create a map to hold the author data
	data := map[string]interface{}{
		"name": name,
	}

	// Convert id to string
	idStr := strconv.Itoa(id)

	// Update the author in the database
	response, _, err := m.Client.From("authors").Update(data, "", "exact").Eq("id", idStr).Execute()
	if err != nil {
		log.Printf("Error updating author: %v", err)
		return 0, err
	}

	// Parse the JSON response
	var updatedAuthor []Author
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&updatedAuthor)
	if err != nil {
		log.Printf("Error parsing JSON response: %v", err)
		return 0, err
	}

	// Check if the author was successfully updated
	if len(updatedAuthor) == 0 {
		log.Printf("No authors returned in response")
		return 0, errors.New("no authors returned in response")
	}

	// Return the ID of the updated author
	return int(updatedAuthor[0].ID), nil
}

// Delete an author by ID
func (m *AuthorModel) Delete(id int) error {
	// Convert id to string
	idStr := strconv.Itoa(id)

	// Delete the author from the database
	_, _, err := m.Client.From("authors").Delete("", "exact").Eq("id", idStr).Execute()
	if err != nil {
		log.Printf("Error deleting author: %v", err)
		return err
	}

	// Return nil
	return nil
}

// Check if an author exists by ID
func (m *AuthorModel) Exists(id int) (bool, error) {
	// Convert id to string
	idStr := strconv.Itoa(id)

	// Query the database for the author by id
	response, count, err := m.Client.From("authors").Select("id", "exact", false).Eq("id", idStr).ExecuteString()
	if err != nil {
		if strings.Contains(err.Error(), "PGRST116") {
			// No rows returned
			return false, nil
		}
		return false, err
	}

	// If count is 0, author doesn't exist
	if count == 0 {
		return false, nil
	}

	// Parse the JSON response
	var authors []struct {
		ID int `json:"id"`
	}
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&authors)
	if err != nil {
		return false, err
	}

	// Check if any author was found
	return len(authors) > 0, nil
}

// Get all authors
func (m *AuthorModel) GetAll() ([]Author, error) {
	// Query the database for all authors
	response, count, err := m.Client.From("authors").Select("*", "exact", false).ExecuteString()
	if err != nil {
		return nil, err
	}

	// If no authors were found, return an empty slice
	if count == 0 {
		return []Author{}, nil
	}

	// Parse the JSON response
	var authors []Author
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&authors)
	if err != nil {
		return nil, err
	}

	// Return the authors
	return authors, nil
}

// Set the AuthUserID for the author
func (m *AuthorModel) SetAuthUserID(id uuid.UUID) {
	m.AuthUserID = id
}

// AuthorWithCounts represents an author with the count of their books
type AuthorWithCounts struct {
	Author
	QuoteCount int `json:"quote_count"`
	BookCount  int `json:"book_count"`
}

// GetWithCounts returns an author with their book count
func (m *AuthorModel) GetWithCounts(id int) (Author, error) {
    var author Author

    // Convert id to string
    idStr := strconv.Itoa(id)

    // Query the database for the author
    count, err := m.Client.From("authors").Select("*", "exact", false).Eq("id", idStr).Single().ExecuteTo(&author)
    if err != nil {
        log.Printf("Error executing query: %v", err)
        return Author{}, ErrNoRecord
    } else if count == 0 {
        log.Printf("No record found for ID: %d", id)
        return Author{}, ErrNoRecord
    }

    // Get quote count
    _, quoteCount, err := m.Client.From("quotes").Select("id", "exact", false).Eq("author_id", idStr).Execute()
    if err != nil {
        log.Printf("Error getting quote count: %v", err)
        return Author{}, err
    }
    author.QuoteCount = int(quoteCount)

    // Get unique books for this author
    var books []Book
    _, err = m.Client.From("quotes").Select("book_id", "exact", false).Eq("author_id", idStr).ExecuteTo(&books)
    if err != nil {
        log.Printf("Error getting books for author: %v", err)
        return Author{}, err
    }

    author.BookCount = len(books)

    return author, nil
}

// GetAllWithCounts returns all authors with their book count
func (m *AuthorModel) GetAllWithCounts() ([]AuthorWithCounts, error) {
	var authorsWithCount []AuthorWithCounts

	// Query the database for authors and their book count
	authorsResponse, authorCount, err := m.Client.From("authors").Select("*", "exact", false).ExecuteString()
	if err != nil {
		log.Printf("Failed to get authors: %v", err)
		return nil, err
	}

	if authorCount == 0 {
		log.Printf("No authors found")
		return []AuthorWithCounts{}, nil
	}

	// Parse the JSON response
	var authorData []map[string]interface{}
	err = json.NewDecoder(strings.NewReader(string(authorsResponse))).Decode(&authorData)
	if err != nil {
		log.Printf("Failed to decode authors: %v", err)
		return nil, err
	}

	// Get the books 
	booksResponse, bookCount, err := m.Client.From("books").Select("*", "exact", false).ExecuteString()
	if err != nil {
		log.Printf("Failed to get books: %v", err)
		return nil, err
	}

	if bookCount == 0 {
		log.Printf("No books found")
		return []AuthorWithCounts{}, nil
	}

	// Parse the JSON response
	var books []Book
	err = json.NewDecoder(strings.NewReader(string(booksResponse))).Decode(&books)
	if err != nil {
		log.Printf("Failed to decode books: %v", err)
		return nil, err
	}

	// Get the quotes
	quotesResponse, quoteCount, err := m.Client.From("quotes").Select("*", "exact", false).ExecuteString()
	if err != nil {
		log.Printf("Failed to get quotes: %v", err)
		return nil, err
	}

	if quoteCount == 0 {
		log.Printf("No quotes found")
		return []AuthorWithCounts{}, nil
	}

	// Parse the JSON response
	var quotes []Quote
	err = json.NewDecoder(strings.NewReader(string(quotesResponse))).Decode(&quotes)
	if err != nil {
		log.Printf("Failed to decode quotes: %v", err)
		return nil, err
	}

	// Create a map to hold the counts for each author
	quoteCountMap := make(map[int]int)
	bookCountMap := make(map[int]int)

	// Iterate through quotes and count the unique books for each author
	for _, quote := range quotes {
		quoteCountMap[quote.AuthorID]++
		bookCountMap[quote.AuthorID]++
	}

	// Process the raw data to create AuthorWithCounts structs
	for _, data := range authorData {
		author := Author{
			ID:   int(data["id"].(float64)),
			Name: data["name"].(string),
		}

		// Check if the user_id is in the data
		if userID, ok := data["user_id"].(string); ok {
			author.UserID, err = uuid.Parse(userID)
			if err != nil {
				log.Printf("Failed to parse user_id: %v", err)
				continue
			}
		}

		// Create an AuthorWithCounts struct
		authorWithCount := AuthorWithCounts{
			Author: author,
			QuoteCount: quoteCountMap[author.ID],
			BookCount: bookCountMap[author.ID],
		}

		// Add the struct to the slice
		authorsWithCount = append(authorsWithCount, authorWithCount)

		// Add the counts to the author
		author.QuoteCount = quoteCountMap[author.ID]
		author.BookCount = bookCountMap[author.ID]
	}

	// Return the authors with their book count
	return authorsWithCount, nil
}