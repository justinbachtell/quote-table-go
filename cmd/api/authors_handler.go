package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/justinbachtell/quote-table-go/internal/models"
	"github.com/justinbachtell/quote-table-go/internal/validator"
)

// Struct to represent the author form data
type authorCreateForm struct {
	Name string `form:"name"`
	validator.Validator `form:"-"`
}

// Handler for the authors page
func (app *application) authorList(w http.ResponseWriter, r *http.Request) {
	authors, err := app.authors.GetAllWithCounts()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Authors = authors

	app.render(w, r, http.StatusOK, "authors.go.tmpl", data)
}

// Handler for the view author page
func (app *application) authorView(w http.ResponseWriter, r *http.Request) {
	// Extract the author ID from the URL
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.serverError(w, r, err)
		return
	}

	// Fetch the author
	author, err := app.authors.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.serverError(w, r, ErrNoRecord)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Fetch books by this author
	books, err := app.books.GetByAuthorID(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Author = author
	data.Books = books

	app.render(w, r, http.StatusOK, "view-author.go.tmpl", data)
}