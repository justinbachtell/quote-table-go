package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/justinbachtell/quote-table-go/internal/models"
)

// A mock quote for testing
var mockQuote = models.Quote{
	ID: 1,
	Author: "John Doe",
	Quote: "To be or not to be, that is the question.",
	UserID: uuid.New(),
	CreatedAt: time.Now(),
}

type QuoteModel struct {}

// Set the authenticated user ID
func (m *QuoteModel) SetAuthUserID(id uuid.UUID) {}

// Insert a quote
func (m *QuoteModel) Insert(quote string, author string, userID uuid.UUID) (int, error) {
	return 2, nil
}

// Get a quote by ID
func (m *QuoteModel) Get(id int) (models.Quote, error) {
	switch id {
	case 1:
		return mockQuote, nil
	default:
		return models.Quote{}, models.ErrNoRecord
	}
}

// Get a quote by UserID
func (m *QuoteModel) GetByUserID(userID uuid.UUID) ([]models.Quote, error) {
	return []models.Quote{mockQuote}, nil
}

// Update a quote
func (m *QuoteModel) Update(id int, quote string, author string) (int, error) {
	return 2, nil
}

// Get the latest quote
func (m *QuoteModel) Latest() ([]models.Quote, error) {
	return []models.Quote{mockQuote}, nil
}

// Check if the quote exists
func (m *QuoteModel) Exists(id int) (bool, error) {
	return true, nil
}

// Delete a quote
func (m *QuoteModel) Delete(id int) error {
	return nil
}