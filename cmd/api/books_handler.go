package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/justinbachtell/quote-table-go/internal/models"
	"github.com/justinbachtell/quote-table-go/internal/validator"
)

// Struct to represent the book form data
type bookCreateForm struct {
	Title        string `form:"title"`
	PublishYear  int    `form:"publish_year"`
	CalendarTime string `form:"calendar_time"`
	ISBN         string `form:"isbn"`
	Source       string `form:"source"`
	validator.Validator `form:"-"`
}

// Handler for the view book page
func (app *application) bookView(w http.ResponseWriter, r *http.Request) {
    id, err := app.readIDParam(r)
    if err != nil || id < 1 {
        app.notFoundResponse(w, r)
        return
    }

    book, err := app.books.Get(int(id))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            app.notFoundResponse(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

    // Fetch quotes for this book
    quotes, err := app.quotes.GetByBookID(id)
    if err != nil {
        // Log the error but don't fail the request
        log.Printf("Error fetching quotes for book %d: %v", id, err)
        quotes = []models.Quote{} // Set an empty slice of quotes
    }

    book.Quotes = quotes

    data := app.newTemplateData(r)
    data.Book = book
	data.Quotes = quotes

    app.render(w, r, http.StatusOK, "view-book.go.tmpl", data)
}

// Handler for the books page
func (app *application) bookList(w http.ResponseWriter, r *http.Request) {
	books, err := app.books.GetAllWithAuthors()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Books = books

	app.render(w, r, http.StatusOK, "books.go.tmpl", data)
}

// Handler for the create book page
func (app *application) bookCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = bookCreateForm{}

	app.render(w, r, http.StatusOK, "create-book.go.tmpl", data)
}

// Handler to process and post the book data
func (app *application) bookCreatePost(w http.ResponseWriter, r *http.Request) {
	var form bookCreateForm

	err := app.formDecoder.Decode(&form, r.PostForm)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 200), "title", "This field cannot be more than 200 characters long")
	form.CheckField(validator.PermittedInt(form.PublishYear, 1, 9999), "publish_year", "This field must be between 1 and 9999")
	form.CheckField(validator.PermittedValues(form.CalendarTime, "A.D.", "B.C."), "calendar_time", "This field must be either A.D. or B.C.")
	form.CheckField(validator.NotBlank(form.ISBN), "isbn", "This field cannot be blank")
	form.CheckField(validator.Matches(form.ISBN, validator.ISBNRegex), "isbn", "This field must be a valid ISBN")
	form.CheckField(validator.MaxChars(form.Source, 500), "source", "This field cannot be more than 500 characters long")

	if !form.ValidField() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create-book.go.tmpl", data)
		return
	}

	id, err := app.books.Insert(form.Title, form.PublishYear, form.CalendarTime, form.ISBN, form.Source)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Book successfully created")

	http.Redirect(w, r, fmt.Sprintf("/book/view/%d", id), http.StatusSeeOther)
}

// Handler for the edit book page
func (app *application) bookEdit(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	book, err := app.books.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFoundResponse(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Book = book

	data.Form = bookCreateForm{
		Title:        book.Title,
		PublishYear:  book.PublishYear,
		CalendarTime: book.CalendarTime,
		ISBN:         book.ISBN,
		Source:       book.Source,
	}

	app.render(w, r, http.StatusOK, "edit-book.go.tmpl", data)
}

// Handler to process and post the book data
func (app *application) bookEditPost(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	_, err = app.books.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFoundResponse(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	var form bookCreateForm
	err = app.formDecoder.Decode(&form, r.PostForm)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 200), "title", "This field cannot be more than 200 characters long")
	form.CheckField(validator.PermittedInt(form.PublishYear, 1, 9999), "publish_year", "This field must be between 1 and 9999")
	form.CheckField(validator.PermittedValues(form.CalendarTime, "A.D.", "B.C."), "calendar_time", "This field must be either A.D. or B.C.")
	form.CheckField(validator.NotBlank(form.ISBN), "isbn", "This field cannot be blank")
	form.CheckField(validator.Matches(form.ISBN, validator.ISBNRegex), "isbn", "This field must be a valid ISBN")
	form.CheckField(validator.MaxChars(form.Source, 500), "source", "This field cannot be more than 500 characters long")

	if !form.ValidField() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "edit-book.go.tmpl", data)
		return
	}

	err = app.books.Update(id, form.Title, form.PublishYear, form.CalendarTime, form.ISBN, form.Source)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Book successfully updated")

	http.Redirect(w, r, fmt.Sprintf("/book/view/%d", id), http.StatusSeeOther)
}

// Handler to delete a book
func (app *application) bookDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.books.Delete(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Book successfully deleted")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}