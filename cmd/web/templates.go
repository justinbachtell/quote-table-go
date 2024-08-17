package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"quotetable.com/internal/models"
	"quotetable.com/ui"
)

// Define a struct to pass data to html templates
type templateData struct {
    CurrentYear int
    Quote       models.Quote
    Quotes      []models.Quote
    User        *models.User
    Form        any
    Flash       string
    IsAuthenticated bool
    CSRFToken   string
    AuthenticatedUserID int
}

// Format the dates to be human readable
func humanDate(t time.Time) string {
	// Return the empty string if the time is zero
	if t.IsZero() {
		return ""
	}

	// Format the time to be human readable
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Initialize a function map object to store a key/value of functions
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// Parses all the templates and caches them
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a map to be the cache
	cache := map[string]*template.Template{}

	// Get a slice of all the embedded page templates
	pages, err := fs.Glob(ui.Files, "html/pages/*.go.tmpl")
	if err != nil {
		return nil, err
	}

	// Iterate through the page templates
	for _, page := range pages {
		// Get the base template
		name := filepath.Base(page)

		// Define the patterns for the template
		patterns := []string{
			"html/base.go.tmpl",
			"html/partials/*.go.tmpl",
			page,
		}

		// Parse the template using the embedded filesystem
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Add the template to the cache
		cache[name] = ts
	}

	// Return the cache
	return cache, nil
}