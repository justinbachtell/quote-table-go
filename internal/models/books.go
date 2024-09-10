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
	GetByAuthorID(authorID int) ([]Book, error)
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
	Quotes       []Quote   `json:"quotes"`
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
    var books []Book

    idStr := strconv.Itoa(id)

    response, count, err := m.Client.From("books").Select("*", "exact", false).Eq("id", idStr).ExecuteString()
    if err != nil {
        log.Printf("Error executing query: %v", err)
        return Book{}, err
    }

    if count == 0 {
        log.Printf("No record found for ID: %d", id)
        return Book{}, ErrNoRecord
    }

    err = json.NewDecoder(strings.NewReader(response)).Decode(&books)
    if err != nil {
        log.Printf("Error decoding JSON: %v", err)
        return Book{}, err
    }

    if len(books) == 0 {
        log.Printf("No book found for ID: %d", id)
        return Book{}, ErrNoRecord
    }

    book := books[0]

	// Get the quotes for this book
    quotesResponse, count, err := m.Client.From("quotes").Select("*", "exact", false).Eq("book_id", strconv.Itoa(book.ID)).ExecuteString()
    if err != nil {
        log.Printf("Error fetching quotes for book %d: %v", book.ID, err)
        return Book{}, err
    }

	// Decode the quotes
	var quotes []Quote
	err = json.NewDecoder(strings.NewReader(quotesResponse)).Decode(&quotes)
	if err != nil {
		log.Printf("Error decoding quotes JSON: %v", err)
		return Book{}, err
	}

	// Set the quotes for this book
	book.Quotes = quotes

    if count > 0 {
		// Fetch the author for this book based on the author id
		authorResponse, _, err := m.Client.From("authors").Select("*", "exact", false).Eq("id", strconv.Itoa(quotes[0].AuthorID)).Single().ExecuteString()
		if err != nil {
			log.Printf("Error fetching author for book %d: %v", book.ID, err)
			return Book{}, err
		}

		var author Author
		err = json.NewDecoder(strings.NewReader(authorResponse)).Decode(&author)
		if err != nil {
			log.Printf("Error decoding author JSON for book %d: %v", book.ID, err)
			return Book{}, err
		}

		book.Author = author
    } else {
		log.Printf("No quotes found for book %d", book.ID)
		book.Quotes = []Quote{}
		book.Author = Author{ID: 0, Name: "Unknown"}
	}

    return book, nil
}

// Get a list of books by author ID
func (m *BookModel) GetByAuthorID(authorID int) ([]Book, error) {
    var books []Book

    // First, get all quotes for this author
    quotesResponse, count, err := m.Client.From("quotes").Select("*", "exact", false).Eq("author_id", strconv.Itoa(authorID)).ExecuteString()
    if err != nil {
        log.Printf("Error fetching quotes for author %d: %v", authorID, err)
        return nil, err
    }

    if count == 0 {
        log.Println("No quotes found for this author")
        return []Book{}, nil
    }

    var quotes []Quote
    err = json.NewDecoder(strings.NewReader(quotesResponse)).Decode(&quotes)
    if err != nil {
        log.Printf("Error decoding quotes JSON: %v", err)
        return nil, err
    }

    // Create a map to store unique books
    bookMap := make(map[int]*Book)

    // Fetch book details for each quote and add to the map
    for _, quote := range quotes {
        if _, exists := bookMap[quote.BookID]; !exists {
            bookResponse, _, err := m.Client.From("books").Select("*", "exact", false).Eq("id", strconv.Itoa(quote.BookID)).Single().ExecuteString()
            if err != nil {
                log.Printf("Error fetching book for quote %d: %v", quote.ID, err)
                continue
            }

            var book Book
            err = json.NewDecoder(strings.NewReader(bookResponse)).Decode(&book)
            if err != nil {
                log.Printf("Error decoding book JSON for quote %d: %v", quote.ID, err)
                continue
            }

            bookMap[quote.BookID] = &book
        }

        // Add the quote to the book's quotes slice
        bookMap[quote.BookID].Quotes = append(bookMap[quote.BookID].Quotes, quote)
    }

    // Convert the map to a slice
    for _, book := range bookMap {
        books = append(books, *book)
    }

    // Fetch author details
    authorResponse, _, err := m.Client.From("authors").Select("*", "exact", false).Eq("id", strconv.Itoa(authorID)).Single().ExecuteString()
    if err != nil {
        log.Printf("Error fetching author %d: %v", authorID, err)
        return books, nil // Return books without author details
    }

    var author Author
    err = json.NewDecoder(strings.NewReader(authorResponse)).Decode(&author)
    if err != nil {
        log.Printf("Error decoding author JSON: %v", err)
        return books, nil // Return books without author details
    }

    // Set the author for all books
    for i := range books {
        books[i].Author = author
    }

    return books, nil
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