package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	// Retrieve the template from the cache
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	// Initialize a new buffer
	buf := new(bytes.Buffer)

	// Write the template to the buffer
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Write the status code to the header
	w.WriteHeader(status)

	// Write the buffer to the response writer
	buf.WriteTo(w)
}

// Returns a pointer to a templateData struct
func (app *application) newTemplateData(r *http.Request) templateData {
    data := templateData{
        CurrentYear:     time.Now().Year(),
        Flash:           app.sessionManager.PopString(r.Context(), "flash"),
        IsAuthenticated: app.isAuthenticated(r),
        CSRFToken:       nosurf.Token(r),
    }

    if data.IsAuthenticated {
		userIdStr := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			app.logger.Error("Failed to parse the user ID", "error", err)
		} else {
			data.AuthenticatedUserID = userId
			user, err := app.users.Get(userId)
			if err == nil {
				data.User = &user
			} else {
				app.logger.Error("Failed to get the user", "error", err)
			}
		}
    }

    return data
}

// Decodes the form data into the provided target destination
func (app *application) decodePostForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	// Call the decoder on the decoder instance and pass in the target destination
	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		// Check if the target destination is invalid
		var invalidDecodeError *form.InvalidDecoderError
		if errors.As(err, &invalidDecodeError) {
			panic(err)
		}

		// For any other error, return normally
		return err
	}

	return nil
}

// Returns true if the current request is from an authenticated user
func (app *application) isAuthenticated(r *http.Request) bool {
	return app.sessionManager.Exists(r.Context(), "authenticatedUserID")
}

// Helper function to urlize a string
func (app *application) urlize(s string) string {
    // Convert to lowercase and replace spaces with hyphens
    return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), " ", "-")
}

// Define a JSON envelope type
type envelope map[string]any

// Defines a helper to send a JSON response
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append a newline to the JSON
	js = append(js, '\n')

	// Add the headers to the response
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the content type header
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response
	w.WriteHeader(status)

	// Write the JSON data to the response
	w.Write(js)

	return nil
}

// Defines a helper to render a JSON response and triage errors
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Limit the request body size to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize a new decoder instance
	dec := json.NewDecoder(r.Body)

	// Disable the strict mode of the decoder
	dec.DisallowUnknownFields()

	// Decode the JSON request body into the destination
	err := dec.Decode(dst)
	if err != nil {
		// If there is an error, triage the error
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		// Check if the error is of the expected type
		switch {
			case errors.As(err, &syntaxError):
				return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
			case errors.Is(err, io.ErrUnexpectedEOF):
				return fmt.Errorf("body contains badly-formed JSON")
			case errors.As(err, &unmarshalTypeError):
				if unmarshalTypeError.Field != "" {
					return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
				}
				return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
			case errors.Is(err, io.EOF):
				return errors.New("body must not be empty")
			case errors.As(err, &maxBytesError):
				return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
			case errors.As(err, &invalidUnmarshalError):
				panic(err)
			default:
				return err
		}
	}

	return nil
}

// Converts a string to a UUID
func (app *application) convertStringToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}