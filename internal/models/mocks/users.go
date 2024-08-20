package mocks

import (
	"github.com/google/uuid"
	"github.com/justinbachtell/quote-table-go/internal/models"
)

type UserModel struct {}

// Set the authenticated user ID
func (m *UserModel) SetAuthUserID(id uuid.UUID) {}

// Insert a user
func (m *UserModel) Insert(name, email, password string) (uuid.UUID, error) {
	switch email {
	case "duplicate@example.com":
		return uuid.Nil, models.ErrDuplicateEmail
	default:
		return uuid.New(), nil
	}
}

// Authenticate a user
func (m *UserModel) Authenticate(email, password string) (uuid.UUID, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return uuid.New(), nil
	}
	return uuid.Nil, models.ErrInvalidCredentials
}

// Check if a user exists
func (m *UserModel) Exists(id uuid.UUID) (bool, error) {
	switch id {
	case uuid.New():
		return true, nil
	default:
		return false, nil
	}
}

// Get user by id
func (m *UserModel) Get(id uuid.UUID) (models.User, error) {
	return models.User{}, nil
}

// Get user by email
func (m *UserModel) GetByEmail(email string) (models.User, error) {
	return models.User{}, nil
}

// Update user's name and email
func (m *UserModel) Update(id uuid.UUID, name, email string) error {
	return nil
}

// Change user's password
func (m *UserModel) ChangePassword(id uuid.UUID, currentPassword, newPassword string) error {
	return nil
}

// Get user by URL name
func (m *UserModel) GetByURLName(urlName string) (models.User, error) {
	return models.User{}, nil
}