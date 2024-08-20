package main

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
)

// Define a list of common errors
var (
    ErrNoRecord           = errors.New("models: no matching record found")
    ErrInvalidCredentials = errors.New("models: invalid credentials")
    ErrDuplicateEmail     = errors.New("models: duplicate email")
)

// Define a generic helper for logging errors
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
}

// Define a server error helper to log the error message and return a generic error message
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	message := "The server encountered a problem and could not process your request"

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// Define a client error helper to return a specific status code and message
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// Define an error helper for sending json-formatted error messages
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) error {
	env := envelope{"error": message}

	// Write the response using the writeJSON helper
	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
	}
	return nil
}

// Define a not found helper to return a 404 status code and message
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

// Define a method not allowed helper to return a 405 status code and message
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

// Define a failed validation response
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusBadRequest, errors)
}