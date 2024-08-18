package mocks

import (
	"github.com/justinbachtell/quote-table-go/internal/models"
)

type UserModel struct {}

// Insert a user
func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "duplicate@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

// Authenticate a user
func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return 1, nil
	}
	return 0, models.ErrInvalidCredentials
}

// Check if a user exists
func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

// Get user by id
func (m *UserModel) Get(id int) (models.User, error) {
	return models.User{}, nil
}

// Get user by email
func (m *UserModel) GetByEmail(email string) (models.User, error) {
	return models.User{}, nil
}

// Update user's name and email
func (m *UserModel) Update(id int, name, email string) error {
	return nil
}

// Change user's password
func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	return nil
}