package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/justinbachtell/quote-table-go/internal/models"
	"github.com/justinbachtell/quote-table-go/internal/validator"
)

// Struct to represent the quote form data
type quoteCreateForm struct {
	Quote string `form:"quote"`
	Author string `form:"author"`
	CreatedAt time.Time `form:"created_at"`
	validator.Validator `form:"-"`
}

// Handler for the view quote page
func (app *application) quoteView(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
    if err != nil || id < 1 {
        app.notFoundResponse(w, r)
        return
    }

    quote, err := app.quotes.Get(int(id))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

    data := app.newTemplateData(r)
    data.Quote = quote

    app.render(w, r, http.StatusOK, "view-quote.go.tmpl", data)
}

// Handler for the create quote page
func (app *application) quoteCreate(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    data.Form = quoteCreateForm{}
    app.render(w, r, http.StatusOK, "create-quote.go.tmpl", data)
}

// Handler to process and post the quote data
func (app *application) quoteCreatePost(w http.ResponseWriter, r *http.Request) {
	// Initialize the template data
	data := app.newTemplateData(r)

	// Fetch the authenticated user
    user, err := app.users.Get(data.AuthenticatedUserID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

	// Initialize the form
    var form quoteCreateForm

    err = app.formDecoder.Decode(&form, r.PostForm)
    if err != nil {
        app.logError(r, err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	validator.ValidateQuote(&form.Validator, form.Quote)
	validator.ValidateAuthor(&form.Validator, form.Author)
	validator.ValidateCharacters(form.Quote)

    if !form.ValidField() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "create-quote.go.tmpl", data)
        return
    }

    id, err := app.quotes.Insert(form.Quote, form.Author, user.ID)
    if err != nil {
		app.logError(r, err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Quote created successfully")

    http.Redirect(w, r, fmt.Sprintf("/quote/view/%d", id), http.StatusSeeOther)
}

// Handler for the edit quote page
func (app *application) quoteEdit(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/quote/edit/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil || id < 1 {
        app.notFoundResponse(w, r)
        return
    }

    quote, err := app.quotes.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            app.notFoundResponse(w, r)
        } else {
            app.logError(r, err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    form := quoteCreateForm{
        Quote:     quote.Quote,
        Author:    quote.Author,
        Validator: validator.Validator{},
    }

    data := app.newTemplateData(r)
    data.Form = form
    data.Quote = quote

    app.render(w, r, http.StatusOK, "edit-quote.go.tmpl", data)
}

// Handler to process and post the quote data
func (app *application) quoteEditPost(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/quote/edit/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

    _, err = app.quotes.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.logError(r, err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    var form quoteCreateForm

    err = app.formDecoder.Decode(&form, r.PostForm)
    if err != nil {
        app.logError(r, err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	validator.ValidateQuote(&form.Validator, form.Quote)
	validator.ValidateAuthor(&form.Validator, form.Author)
	validator.ValidateCharacters(form.Quote)

    if !form.ValidField() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "edit-quote.go.tmpl", data)
        return
    }

    _, err = app.quotes.Update(id, form.Quote, form.Author)
    if err != nil {
        app.logError(r, err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Quote updated successfully")

    http.Redirect(w, r, fmt.Sprintf("/quote/view/%d", id), http.StatusSeeOther)
}

// Handler to delete a quote
func (app *application) quoteDeletePost(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/quote/delete/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

    err = app.quotes.Delete(id)
    if err != nil {
        app.logError(r, err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Quote deleted successfully")

    http.Redirect(w, r, "/", http.StatusSeeOther)
}