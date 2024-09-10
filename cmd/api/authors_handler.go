package main

import (
	"errors"
	"net/http"

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
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	author, err := app.authors.Get(int(id))
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFoundResponse(w, r)
		} else {
			app.serverError(w, r, err)
		}
	}

	data := app.newTemplateData(r)
	data.Author = author

	app.render(w, r, http.StatusOK, "view-author.go.tmpl", data)
}