package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
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

// Define a server error helper to log the error message and return a generic error message
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Define a client error helper to return a specific status code and message
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
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
        userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
        user, err := app.users.Get(userID)
        if err == nil {
            data.User = &user
        } else {
            app.logger.Error(err.Error())
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
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}