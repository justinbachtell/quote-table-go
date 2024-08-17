package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"quotetable.com/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	// Initialize a new response recorder and dummy request
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock handler that returns a 200 OK status code
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Call the common headers middleware
	commonHeaders(next).ServeHTTP(rr, r)

	// Call the result method on the response recorder
	rs := rr.Result()

	// Check that the middleware adds the correct headers
	expectedValue := "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"
	assert.Equal(t, rs.Header.Get("Content-Security-Policy"), expectedValue)

	// Check that the middleware sets the correct referrer policy
	expectedValue = "origin-when-cross-origin"
	assert.Equal(t, rs.Header.Get("Referrer-Policy"), expectedValue)

	// Check that the middleware sets the correct content type
	expectedValue = "nosniff"
	assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), expectedValue)

	// Check that the middleware sets the correct frame options
	expectedValue = "deny"
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), expectedValue)

	// Check that the middleware sets the correct XSS protection
	expectedValue = "0"
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), expectedValue)

	// Check that the middleware sets the correct Server header
	expectedValue = "Go"
	assert.Equal(t, rs.Header.Get("Server"), expectedValue)

	// Check that the middleware calls the next handler with the correct status code
	assert.Equal(t, rs.StatusCode, http.StatusOK)

	// Close the response body when the function returns
	defer rs.Body.Close()

	// Read the response body
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	// Check that the body returns a 200 OK status code
	assert.Equal(t, string(body), "OK")
}