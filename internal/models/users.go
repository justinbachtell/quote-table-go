package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"golang.org/x/crypto/bcrypt"
)

// Define an interface for the UserModel4
type UserModelInterface interface {
	Insert(name, email, password string) (uuid.UUID, error)
	Authenticate(email string, password string) (uuid.UUID, error)
	Exists(id uuid.UUID) (bool, error)
	Update(id uuid.UUID, name, email string) error
	ChangePassword(id uuid.UUID, currentPassword, newPassword string) error
	// Delete(id uuid.UUID) error
	Get(id uuid.UUID) (User, error)
	GetByEmail(email string) (User, error)
	GetByURLName(urlName string) (User, error)
	SetAuthUserID(id uuid.UUID)
}

// User represents a user in the database
type User struct {
	ID        uuid.UUID    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  []byte    `json:"-"`
	ProfileSlug string `json:"profile_slug"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

// The model used in the connection pool
type UserModel struct {
	Client *supabase.Client
	AuthUserID uuid.UUID
}

// Set the authenticated user ID
func (m *UserModel) SetAuthUserID(id uuid.UUID) {
    m.AuthUserID = id
}

// Insert adds a new user to the database
func (m *UserModel) Insert(name, email, password string) (uuid.UUID, error) {
	// Hash the password with the number of specified salt rounds
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return uuid.Nil, err
    }

	// Create a map to hold the user data
    data := map[string]interface{}{
        "name":            name,
        "email":           email,
        "hashed_password": string(hashedPassword),
		"profile_slug":    strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-"),
        "created":         time.Now(),
    }

	// Insert the user into the database
    response, _, err := m.Client.From("users").Insert(data, false, "", "", "").ExecuteString()
    if err != nil {
        // Check if the error is due to a duplicate email
        if strings.Contains(err.Error(), "users_uc_email") {
            return uuid.Nil, ErrDuplicateEmail
        }
        return uuid.Nil, err
    }

	// Check if the user was inserted
	if len(response) == 0 {
		return uuid.Nil, errors.New("no user was inserted")
	}

	// Decode the response
	var insertedUser []struct {
		ID uuid.UUID `json:"id"`
	}

	// Check if the user was inserted
	err = json.NewDecoder(strings.NewReader(response)).Decode(&insertedUser)
	if err != nil {
		fmt.Printf("Failed to decode insert response: %v", err)
		return uuid.Nil, err
	}

	if len(insertedUser) == 0 {
		return uuid.Nil, errors.New("no user ID returned")
	}

	fmt.Printf("Successfully inserted user with ID: %s", insertedUser[0].ID)
	return insertedUser[0].ID, nil
}

// Authenticate verifies the user's email and password.
func (m *UserModel) Authenticate(email string, password string) (uuid.UUID, error) {
	// Query the database for the user with the given email
	response, count, err := m.Client.From("users").Select("*", "exact", false).Eq("email", email).Single().ExecuteString()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "PGRST116")  {
			return uuid.Nil, ErrInvalidCredentials
		} else {
			return uuid.Nil, err
		}
	}

	// Ensure a user was found
	if count == 0 {
		return uuid.Nil, ErrInvalidCredentials
	}

	// Use a temporary struct to unmarshal the response
	var tempUser struct {
		ID              uuid.UUID       `json:"id"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		HashedPassword  string    `json:"hashed_password"`
		ProfileSlug     string    `json:"profile_slug"`
		Created         time.Time `json:"created_at"`
		Updated         time.Time `json:"updated"`
	}

	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&tempUser)
	if err != nil {
		return uuid.Nil, err
	}

	// Check if the password is a valid bcrypt hash
	if len(tempUser.HashedPassword) < bcrypt.MinCost {
		return uuid.Nil, errors.New("invalid bcrypt hash")
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(tempUser.HashedPassword), []byte(password))
	if err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	return tempUser.ID, nil
}

// Check if the user exists
func (m *UserModel) Exists(id uuid.UUID) (bool, error) {
    // Query the database for the user with the given id
    response, count, err := m.Client.From("users").Select("id", "exact", false).Eq("id", id.String()).ExecuteString()
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
        ID uuid.UUID `json:"id"`
    }
    err = json.NewDecoder(strings.NewReader(string(response))).Decode(&users)
    if err != nil {
        return false, err
    }

    // Check if any user was found
    return len(users) > 0, nil
}

// Get user by id
func (m *UserModel) Get(id uuid.UUID) (User, error) {
	// Query the database for the user with the given id
	response, count, err := m.Client.From("users").Select("*", "exact", false).Eq("id", id.String()).Single().ExecuteString()
	if err != nil {
		return User{}, err
	}

	if count == 0 {
		return User{}, ErrNoRecord
	}

	var user User
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&user)
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
func (m *UserModel) Update(id uuid.UUID, name, email string) error {
	// Create a map to hold the user data
	data := map[string]interface{}{
		"name":  name,
		"email": email,
		"profile_slug":    strings.ReplaceAll(strings.ToLower(strings.TrimSpace(name)), " ", "-"),
	}

	// Update the user in the database
	_, _, err := m.Client.From("users").Update(data, "", "").Eq("id", id.String()).ExecuteString()
	if err != nil {
		return err
	}
	
	// Return the user
	return nil
}

// ChangePassword updates the user's password in the database
func (m *UserModel) ChangePassword(id uuid.UUID, currentPassword, newPassword string) error {
	// Get the current user data
	response, count, err := m.Client.From("users").Select("*", "exact", false).Eq("id", id.String()).Single().ExecuteString()
	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNoRecord
	}

	var user User
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&user)
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

	_, _, err = m.Client.From("users").Update(data, "", "").Eq("id", id.String()).ExecuteString()
	if err != nil {
		return err
	}

	return nil
}

// Get user by URL name
func (m *UserModel) GetByURLName(urlName string) (User, error) {
	// Query the database for the user with the given URL name
	response, _, err := m.Client.From("users").Select("*", "exact", false).Eq("profile_slug", urlName).Single().ExecuteString()
	if err != nil {
		return User{}, err
	}

	// Decode the response
	var user User
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&user)
	if err != nil {
		return User{}, err
	}

	// Return the user
	return user, nil
}

// Get user's id from URL name
func (m *UserModel) GetIDFromURLName(urlName string) (uuid.UUID, error) {
	// Query the database for the user with the given URL name
	response, _, err := m.Client.From("users").Select("id", "exact", false).Eq("profile_slug", urlName).Single().ExecuteString()
	if err != nil {
		return uuid.Nil, err
	}

	// Decode the response
	var user User
	err = json.NewDecoder(strings.NewReader(string(response))).Decode(&user)
	if err != nil {
		return uuid.Nil, err
	}

	// Return the user
	return user.ID, nil
}