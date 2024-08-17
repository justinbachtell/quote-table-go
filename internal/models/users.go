package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/supabase-community/supabase-go"
	"golang.org/x/crypto/bcrypt"
)

// Define an interface for the UserModel4
type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email string, password string) (int, error)
	Exists(id int) (bool, error)
	Update(id int, name, email string) error
	ChangePassword(id int, currentPassword, newPassword string) error
	// Delete(id int) error
	Get(id int) (User, error)
	GetByEmail(email string) (User, error)
}

// User represents a user in the database
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  []byte    `json:"hashed_password"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// The model used in the connection pool
type UserModel struct {
	Client *supabase.Client
}

// Insert adds a new user to the database
func (m *UserModel) Insert(name, email, password string) error {
	// Hash the password with the number of specified salt rounds
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12) // Default cost is 12
	if err != nil {
		return err
	}

	// Create a map to hold the user data
	data := map[string]interface{}{
		"name":            name,
		"email":           email,
		"hashed_password": hashedPassword,
		"created":         time.Now(),
	}

	// Insert the user into the database and get the response
	_, _, err = m.Client.From("users").Insert(data, false, "", "", "").ExecuteString()
    if err != nil {
        if strings.Contains(err.Error(), "23505") {
            log.Printf("User already exists: %v", err)
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

// Authenticate verifies the user's email and password.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	// Query the database for the user with the given email
	response, count, err := m.Client.From("users").Select("*", "exact", false).Eq("email", email).Single().ExecuteString()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "PGRST116")  {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	// Ensure a user was found
	if count == 0 {
		return 0, ErrInvalidCredentials
	}

	// Extract the user data from the response
	var user struct {
		ID        int       `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		Password  []byte    `json:"hashed_password"`
		Created time.Time `json:"created_at"`
		Updated time.Time `json:"updated_at"`
	}
	err = json.Unmarshal([]byte(response), &user)
	if err != nil {
		return 0, err
	}

	// Check if the password is a valid bcrypt hash
	if len(user.Password) < bcrypt.MinCost {
		return 0, errors.New("invalid bcrypt hash")
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return 0, ErrInvalidCredentials
	}

	return user.ID, nil
}

// Check if the user exists
func (m *UserModel) Exists(id int) (bool, error) {
    // Convert id to string
    idStr := strconv.Itoa(id)

    // Query the database for the user with the given id
    response, count, err := m.Client.From("users").Select("id", "exact", false).Eq("id", idStr).ExecuteString()
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
    var users []struct {
        ID int `json:"id"`
    }
    err = json.Unmarshal([]byte(response), &users)
    if err != nil {
        return false, err
    }

    // Check if any user was found
    return len(users) > 0, nil
}

// Get user by id
func (m *UserModel) Get(id int) (User, error) {
	// Convert id to string
	idStr := strconv.Itoa(id)

	// Query the database for the user with the given id
	response, count, err := m.Client.From("users").Select("*", "exact", false).Eq("id", idStr).Single().ExecuteString()
	if err != nil {
		return User{}, err
	}

	if count == 0 {
		return User{}, ErrNoRecord
	}

	var user User
	err = json.Unmarshal([]byte(response), &user)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// Get user by email
func (m *UserModel) GetByEmail(email string) (User, error) {
	// Query the database for the user with the given email
	_, _, err := m.Client.From("users").Select("name, email", "exact", false).Eq("email", email).Single().ExecuteString()
	if err != nil {
		return User{}, err
	}

	// Return the user
	return User{}, nil
}

// Update user's name and email
func (m *UserModel) Update(id int, name, email string) error {
	// Convert id to string
	idStr := strconv.Itoa(id)

	// Create a map to hold the user data
	data := map[string]interface{}{
		"name":  name,
		"email": email,
	}

	// Update the user in the database
	_, _, err := m.Client.From("users").Update(data, "", "").Eq("id", idStr).ExecuteString()
	if err != nil {
		return err
	}
	
	// Return the user
	return nil
}

// ChangePassword updates the user's password in the database
func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	// Convert id to string
	idStr := strconv.Itoa(id)

	// Get the current user data
	response, count, err := m.Client.From("users").Select("*", "exact", false).Eq("id", idStr).Single().ExecuteString()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRecord
	}

	var user User
	err = json.Unmarshal([]byte(response), &user)
	if err != nil {
		return err
	}

	// Verify the current password
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	// Update the password in the database
	data := map[string]interface{}{
		"hashed_password": hashedPassword,
	}

	_, _, err = m.Client.From("users").Update(data, "", "").Eq("id", idStr).ExecuteString()
	if err != nil {
		return err
	}

	return nil
}