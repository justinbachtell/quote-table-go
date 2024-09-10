package mocks

import (
	"github.com/google/uuid"
	"github.com/justinbachtell/quote-table-go/internal/models"
)

// A mock author for testing
var mockAuthor = models.Author{
	ID: 1,
	Name: "John Doe",
}

type AuthorModel struct {}

// Insert an author
func (m *AuthorModel) Insert(name string) (int, error) {
	return 2, nil
}


// Get an author by ID
func (m *AuthorModel) Get(id int) (models.Author, error) {
	switch id {
	case 1:
		return mockAuthor, nil
	default:
		return models.Author{}, models.ErrNoRecord
	}
}

// Get an author by name
func (m *AuthorModel) GetByName(name string) (models.Author, error) {
	switch name {
	case "John Doe":
		return mockAuthor, nil
	default:
		return models.Author{}, models.ErrNoRecord
	}
}

// Get quotes by author
func (m *AuthorModel) GetQuotesByAuthor(authorID int) ([]models.Quote, error) {
	return []models.Quote{
		{
			ID: 1,
			Quote: "Hello, world!",
		},
	}, nil
}

// Get books by author
func (m *AuthorModel) GetBooksByAuthor(authorID int) ([]models.Book, error) {
	return []models.Book{
		{
			ID: 1,
			Title: "Hello, world!",
		},
	}, nil
}

// Get an author with counts
func (m *AuthorModel) GetWithCounts(id int) (models.AuthorWithCounts, error) {
	return models.AuthorWithCounts{
		Author: mockAuthor,
		BookCount: 1,
		QuoteCount: 1,
	}, nil
}

// Update an author
func (m *AuthorModel) Update(id int, name string) (int, error) {
	return 2, nil
}

// Get all authors
func (m *AuthorModel) GetAll() ([]models.Author, error) {
	return []models.Author{mockAuthor}, nil
}

// Get all authors with book count
func (m *AuthorModel) GetAllWithCounts() ([]models.AuthorWithCounts, error) {
	return []models.AuthorWithCounts{
		{
			Author:    mockAuthor,
			BookCount: 1,
		},
	}, nil
}

// Check if the author exists
func (m *AuthorModel) Exists(id int) (bool, error) {
	return true, nil
}

// Delete an author
func (m *AuthorModel) Delete(id int) error {
	return nil
}

// Set the authenticated user ID
func (m *AuthorModel) SetAuthUserID(id uuid.UUID) {}