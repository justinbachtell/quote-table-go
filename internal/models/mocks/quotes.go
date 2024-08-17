package mocks

import (
	"time"

	"quotetable.com/internal/models"
)

// A mock quote for testing
var mockQuote = models.Quote{
	ID: 1,
	Author: "John Doe",
	Quote: "To be or not to be, that is the question",
	Created: time.Now(),
}

type QuoteModel struct {}

// Insert a quote
func (m *QuoteModel) Insert(quote string, author string) (int, error) {
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