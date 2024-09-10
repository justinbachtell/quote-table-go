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

// Define an interface for the BookModel
type BookModelInterface interface {
	Insert(title string, publishYear int, calendarTime string, isbn string, source string) (int, error)
	Get(id int) (Book, error)
	GetAllWithAuthors() ([]Book, error)
	Update(id int, title string, publishYear int, calendarTime string, isbn string, source string) error
	Delete(id int) error
	GetAll() ([]Book, error)
	Exists(id int) (bool, error)
	SetAuthUserID(id uuid.UUID)
}

// Book represents a book in the database
type Book struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	PublishYear  int       `json:"publish_year"`
	CalendarTime string    `json:"calendar_time"`
	ISBN         string    `json:"isbn"`
	Source       string    `json:"source"`
	UserID       uuid.UUID `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Author       Author    `json:"author"`
}

// The model used in the connection pool
type BookModel struct {
	Client *supabase.Client
	AuthUserID uuid.UUID
}

// Insert adds a new book to the database
func (m *BookModel) Insert(title string, publishYear int, calendarTime string, isbn string, source string) (int, error) {
	data := map[string]interface{}{
		"title":         title,
		"publish_year":  publishYear,
		"calendar_time": calendarTime,
		"isbn":          isbn,
		"source":        source,
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	response, _, err := m.Client.From("books").Insert(data, false, "", "", "").ExecuteString()
	if err != nil {
		return 0, err
	}

	var insertedBook []Book
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&insertedBook)
	if err != nil {
		return 0, err
	}

	if len(insertedBook) == 0 {
		return 0, errors.New("no books returned in response")
	}

	return insertedBook[0].ID, nil
}

// Get a single book by ID
func (m *BookModel) Get(id int) (Book, error) {
	var b Book
	var a Author

	idStr := strconv.Itoa(id)

	count, err := m.Client.From("books").Select("*", "exact", false).Eq("id", idStr).Single().ExecuteTo(&b)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Book{}, ErrNoRecord
	} else if count > 1 {
		log.Printf("Unexpected count > 1 for ID: %d", id)
		return Book{}, err
	} else if count == 0 {
		log.Printf("No record found for ID: %d", id)
		return Book{}, ErrNoRecord
	}

	log.Printf("Retrieved book: %+v", b)

	_, err = m.Client.From("authors").Select("*", "exact", false).Eq("id", strconv.Itoa(b.Author.ID)).Single().ExecuteTo(&a)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return Book{}, err
	}

	// Set the author to the book
	b.Author = a

	return b, nil
}

// Get all books with authors
func (m *BookModel) GetAllWithAuthors() ([]Book, error) {
    var books []Book

    response, count, err := m.Client.From("books").Select("*", "exact", false).ExecuteString()
    if err != nil {
        log.Printf("Error fetching books: %v", err)
        return nil, err
    }

    if count == 0 {
        log.Println("No books found")
        return []Book{}, nil
    }

    err = json.NewDecoder(strings.NewReader(string(response))).Decode(&books)
    if err != nil {
        log.Printf("Error decoding books JSON: %v", err)
        return nil, err
    }

    for i, book := range books {

        // Fetch quotes for this book
        var quotes []Quote
        quotesResponse, _, err := m.Client.From("quotes").Select("*", "exact", false).Eq("book_id", strconv.Itoa(book.ID)).ExecuteString()
        if err != nil {
            log.Printf("Error fetching quotes for book %d: %v", book.ID, err)
            continue
        }

        err = json.NewDecoder(strings.NewReader(quotesResponse)).Decode(&quotes)
        if err != nil {
            log.Printf("Error decoding quotes JSON for book %d: %v", book.ID, err)
            continue
        }

        if len(quotes) > 0 {
            // Fetch the author of the first quote
            var author Author
            authorResponse, _, err := m.Client.From("authors").Select("*", "exact", false).Eq("id", strconv.Itoa(quotes[0].AuthorID)).Single().ExecuteString()
            if err != nil {
                log.Printf("Error fetching author for book %d: %v", book.ID, err)
                books[i].Author = Author{ID: 0, Name: "Unknown"}
                continue
            }

            err = json.NewDecoder(strings.NewReader(authorResponse)).Decode(&author)
            if err != nil {
                log.Printf("Error decoding author JSON for book %d: %v", book.ID, err)
                books[i].Author = Author{ID: 0, Name: "Unknown"}
                continue
            }

            books[i].Author = author
        } else {
            books[i].Author = Author{ID: 0, Name: "Unknown"}
        }
    }

    return books, nil
}

// Update a book by ID
func (m *BookModel) Update(id int, title string, publishYear int, calendarTime string, isbn string, source string) error {
	data := map[string]interface{}{
		"title":         title,
		"publish_year":  publishYear,
		"calendar_time": calendarTime,
		"isbn":          isbn,
		"source":        source,
		"updated_at":    time.Now(),
	}

	idStr := strconv.Itoa(id)

	_, _, err := m.Client.From("books").Update(data, "", "exact").Eq("id", idStr).Execute()
	if err != nil {
		log.Printf("Error updating book: %v", err)
		return err
	}

	return nil
}

// Delete a book by ID
func (m *BookModel) Delete(id int) error {
	idStr := strconv.Itoa(id)

	_, _, err := m.Client.From("books").Delete("", "exact").Eq("id", idStr).Execute()
	if err != nil {
		log.Printf("Error deleting book: %v", err)
		return err
	}

	return nil
}

// Get all books
func (m *BookModel) GetAll() ([]Book, error) {
	response, count, err := m.Client.From("books").Select("*", "exact", false).ExecuteString()
	if err != nil {
		return nil, err
	}


	// If no books were found, return an empty slice
	if count == 0 {
		return []Book{}, nil
	}

	var books []Book
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&books)
	if err != nil {
		return nil,  err
	}


	return books, nil
}

// Check if the book exists
func (m *BookModel) Exists(id int) (bool, error) {
	idStr := strconv.Itoa(id)

	response, count, err := m.Client.From("books").Select("id", "exact", false).Eq("id", idStr).ExecuteString()
	if err != nil {
		if strings.Contains(err.Error(), "PGRST116") {
			// No rows returned
			return false, nil
		}
		return false, err
	}

	// If count is 0, book doesn't exist
	if count == 0 {
		return false, nil
	}

	// Parse the JSON response
	var books []struct {
		ID int `json:"id"`
	}

	// Decode the JSON response
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&books)
	if err != nil {
		return false, err
	}

	// Check if any book was found
	return len(books) > 0, nil
}

// Set the AuthUserID for the book
func (m *BookModel) SetAuthUserID(id uuid.UUID) {
	m.AuthUserID = id
}