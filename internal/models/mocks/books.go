package mocks

import (
	"github.com/google/uuid"
	"github.com/justinbachtell/quote-table-go/internal/models"
)

// A mock book for testing
var mockBook = models.Book{
	ID: 1,
	Title: "Test Book",
	UserID: uuid.New(),
}

type BookModel struct {}

// Insert a book
func (m *BookModel) Insert(title string, publishYear int, calendarTime string, isbn string, source string) (int, error) {
	return 2, nil
}

// Get a book by ID
func (m *BookModel) Get(id int) (models.Book, error) {
	switch id {
	case 1:
		return mockBook, nil
	default:
		return models.Book{}, models.ErrNoRecord
	}
}

// Get a list of books by author ID
func (m *BookModel) GetByAuthorID(authorID int) ([]models.Book, error) {
	return []models.Book{mockBook}, nil
}

// Get all books
func (m *BookModel) GetAll() ([]models.Book, error) {
	return []models.Book{mockBook}, nil
}

// Get all books with authors
func (m *BookModel) GetAllWithAuthors() ([]models.Book, error) {
	return []models.Book{mockBook}, nil
}

// Update a book
func (m *BookModel) Update(id int, title string, publishYear int, calendarTime string, isbn string, source string) error {
	return nil
}

// Delete a book
func (m *BookModel) Delete(id int) error {
	return nil
}

// Check if the book exists
func (m *BookModel) Exists(id int) (bool, error) {
	return true, nil
}

// Set the authenticated user ID
func (m *BookModel) SetAuthUserID(id uuid.UUID) {}