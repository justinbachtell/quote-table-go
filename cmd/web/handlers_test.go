package main

import (
	"net/http"
	"net/url"
	"testing"

	"quotetable.com/internal/assert"
)

// Tests the ping route
func TestPing(t *testing.T) {
	// Create a new test application
	app := newTestApplication(t)

	// Create a new test server
	ts := newTestServer(t, app.routes())

	// Shutdown the test server when the test is complete
	defer ts.Close()

	// Make a GET request to the /ping route
	code, _, body := ts.get(t, "/ping")

	// Check the status code
	assert.Equal(t, code, http.StatusOK)

	// Check the response body
	assert.Equal(t, body, "OK")
}

// Tests the snippet view route
func TestQuoteView(t *testing.T) {
	// Create a new test application
	app := newTestApplication(t)

	// Establish a new test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Set up test data to check responses
	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{"Valid ID", "/quote/view/1", http.StatusOK, ""},
		{"Non-existent ID", "/quote/view/10000000", http.StatusNotFound, ""},
		{"Negative ID", "/quote/view/-1", http.StatusNotFound, ""},
		{"Decimal ID", "/quote/view/1.23", http.StatusNotFound, ""},
		{"String ID", "/quote/view/foo", http.StatusNotFound, ""},
		{"Empty ID", "/quote/view/", http.StatusNotFound, ""},
	}

	// Loop through each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a GET request to the /quote/view route
			code, _, body := ts.get(t, tt.urlPath)
			
			// Check the status code
			assert.Equal(t, code, tt.wantCode)

			// Check the response body if it is expected
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

// Tests the user signup route
func TestUserSignup(t *testing.T) {
	// Create a new test application
	app := newTestApplication(t)

	// Establish a new test server
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Make a GET request to the /user/signup route and extract the CSRF token
	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)

	// Define some constants to be used for the test cases
	const (
		validName = "John Doe"
		validEmail = "john.doe@example.com"
		validPassword = "validpa$$word"
	)

	// Define the test cases
	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		},
		{
			name:         "Invalid CSRF Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantBody:     "The name field cannot be blank",
		},
		{
			name:         "Empty email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantBody:     "The email field cannot be blank",
		},
		{
			name:         "Empty password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantBody:     "The password field cannot be blank",
		},
		{
			name:         "Invalid email",
			userName:     validName,
			userEmail:    "invalidEmail",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantBody:     "The email field is not a valid email address",
		},
		{
			name:         "Short password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantBody:     "The password field is too short",
		},
		{
			name:         "Duplicate email",
			userName:     validName,
			userEmail:    "duplicate@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantBody:     "This email address is already in use",
		},
	}

	// Loop through each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			// Make a POST request to the /user/signup route
			code, _, body := ts.postForm(t, "/user/signup", form)

			// Check the status code
			assert.Equal(t, code, tt.wantCode)

			// Check the response body
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}