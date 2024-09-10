package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 8)
	if err != nil {
		return 0, errors.New("invalid id parameter")
	}

	return int(id), nil
}

// Handler for the home page
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the latest quotes from the database
	quotes, err := app.quotes.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Get all authors
	authors, err := app.authors.GetAllWithCounts()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Get all books
	books, err := app.books.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Get the template data and add the quotes, authors, and books slices
	data := app.newTemplateData(r)
	data.Quotes = quotes
	data.Authors = authors
	data.Books = books

	// render the home page
	app.render(w, r, http.StatusOK, "home.go.tmpl", data)
}

// Handler for the terms page
func (app *application) terms(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "terms.go.tmpl", data)
}

// Handler for the privacy page
func (app *application) privacy(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "privacy.go.tmpl", data)
}

// Handler for the pricing page
func (app *application) pricing(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "pricing.go.tmpl", data)
}

// Handler to return a 200 OK status code
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	
	r.Body = http.MaxBytesReader(w, r.Body, 100)
}