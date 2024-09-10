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
	AuthorID int `form:"author-selector"`
	NewAuthorName string `form:"new_author_name"`
	BookID int `form:"book-selector"`
	NewBookTitle string `form:"new_book_title"`
	NewBookPublishYear int `form:"new_book_publish_year"`
	NewBookCalendarTime string `form:"new_book_calendar_time"`
	NewBookISBN string `form:"new_book_isbn"`
	NewBookSource string `form:"new_book_source"`
	PageNumber string `form:"page_number"`
	IsPrivate bool `form:"is_private"`
	CreatedAt time.Time `form:"created_at"`
	UpdatedAt time.Time `form:"updated_at"`
	validator.Validator `form:"-"`
}

// Handler for the view quote page
func (app *application) quoteView(w http.ResponseWriter, r *http.Request) {
	// Get the quote ID from the URL
	id, err := app.readIDParam(r)
    if err != nil || id < 1 {
        app.notFoundResponse(w, r)
        return
    }

	// Get the quote
    quote, err := app.quotes.GetWithAuthorAndBook(int(id))
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            app.notFoundResponse(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

	// Initialize the template data
    data := app.newTemplateData(r)
    data.Quote = quote
	data.Author = quote.Author

	// Render the view quote page
    app.render(w, r, http.StatusOK, "view-quote.go.tmpl", data)
}

// Handler for the create quote page
func (app *application) quoteCreate(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    data.Form = quoteCreateForm{}

    authors, err := app.authors.GetAllWithCounts()
    if err != nil {
        app.serverError(w, r, err)
        return
    }
    data.Authors = authors

    // Fetch all books
    books, err := app.books.GetAll()
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    // Add the authors and books to the template data
    data.Books = books

    // Render the create quote page
    app.render(w, r, http.StatusOK, "create-quote.go.tmpl", data)
}

// Handler to process and post the quote data
func (app *application) quoteCreatePost(w http.ResponseWriter, r *http.Request) {
    // Initialize the template data
    data := app.newTemplateData(r)

    // Initialize the form
    var form quoteCreateForm
    var authorID int
    var bookID int
    var err error

    // Fetch the authenticated user
    user, err := app.users.Get(data.AuthenticatedUserID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    // Decode the form
    err = app.formDecoder.Decode(&form, r.PostForm)
    if err != nil {
        app.logError(r, err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Validate the form
    validator.ValidateQuote(&form.Validator, form.Quote)
    validator.ValidateCharacters(form.Quote)

    // Handle author
    authorSelector := r.PostForm.Get("author-selector")
    if authorSelector != "" {
        authorID, err = strconv.Atoi(authorSelector)
        if err != nil {
            form.AddFieldError("author", "Invalid author selection")
        }
    } else if form.NewAuthorName != "" {
        validator.ValidateAuthor(&form.Validator, form.NewAuthorName)
        if form.ValidField() {
            authorID, err = app.authors.Insert(form.NewAuthorName)
            if err != nil {
                app.serverError(w, r, err)
                return
            }
        }
    } else {
        form.AddFieldError("author", "Please select an author or enter a new one.")
    }

    // Handle book
    bookSelector := r.PostForm.Get("book-selector")
    if bookSelector != "" {
        bookID, err = strconv.Atoi(bookSelector)
        if err != nil {
            form.AddFieldError("book", "Invalid book selection")
        }
    } else if form.NewBookTitle != "" {
        validator.ValidateBook(&form.Validator, form.NewBookTitle, form.NewBookPublishYear, form.NewBookCalendarTime, form.NewBookISBN, form.NewBookSource)
        if form.ValidField() {
            bookID, err = app.books.Insert(form.NewBookTitle, form.NewBookPublishYear, form.NewBookCalendarTime, form.NewBookISBN, form.NewBookSource)
            if err != nil {
                app.serverError(w, r, err)
                return
            }
        }
    } else {
        form.AddFieldError("book", "Please select a book or enter a new one.")
    }

    // If the form is not valid, re-render the form
    if !form.ValidField() {
        data.Form = form
        // Fetch all authors and books for the dropdowns
        authors, err := app.authors.GetAllWithCounts()
        if err != nil {
            app.serverError(w, r, err)
            return
        }
        books, err := app.books.GetAll()
        if err != nil {
            app.serverError(w, r, err)
            return
        }
        data.Authors = authors
        data.Books = books
        app.render(w, r, http.StatusUnprocessableEntity, "create-quote.go.tmpl", data)
        return
    }

    // Insert the quote
    id, err := app.quotes.Insert(form.Quote, authorID, bookID, form.PageNumber, form.IsPrivate, user.ID)
    if err != nil {
        app.logError(r, err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Add a flash message
    app.sessionManager.Put(r.Context(), "flash", "Quote created successfully")

    // Redirect to the quote view page
    http.Redirect(w, r, fmt.Sprintf("/quote/view/%d", id), http.StatusSeeOther)
}

// Handler for the edit quote page
func (app *application) quoteEdit(w http.ResponseWriter, r *http.Request) {
    id, err := app.readIDParam(r)
    if err != nil || id < 1 {
        app.notFoundResponse(w, r)
        return
    }

    // Get the quote
	quote, err := app.quotes.GetWithAuthorAndBook(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            app.notFoundResponse(w, r)
        } else {
			app.logError(r, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    // Fetch all authors
    authors, err := app.authors.GetAllWithCounts()
    if err != nil {
        app.serverError(w, r, err)
        return
    }

	// Fetch all books
	books, err := app.books.GetAll()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Initialize the template data
    data := app.newTemplateData(r)
    data.Quote = quote
    data.Authors = authors
	data.Books = books

	// Initialize the form
    data.Form = quoteCreateForm{
        Quote:    quote.Quote,
        AuthorID: quote.AuthorID,
		BookID: quote.BookID,
		PageNumber: quote.PageNumber,
		IsPrivate: quote.IsPrivate,
    }

	// Render the edit quote page
    app.render(w, r, http.StatusOK, "edit-quote.go.tmpl", data)
}

// Handler to process and post the quote data
func (app *application) quoteEditPost(w http.ResponseWriter, r *http.Request) {
    id, err := app.readIDParam(r)
    if err != nil || id < 1 {
        app.notFoundResponse(w, r)
        return
    }

    // Fetch the original quote
    originalQuote, err := app.quotes.GetWithAuthorAndBook(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            app.notFoundResponse(w, r)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

	// Initialize the form
    var form quoteCreateForm

	// Decode the form
    err = app.formDecoder.Decode(&form, r.PostForm)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
        return
    }

	// Validate the form
    validator.ValidateQuote(&form.Validator, form.Quote)
    validator.ValidateCharacters(form.Quote)

	// Initialize authorID and bookID
	var authorID int
	var bookID int
	var pageNumber string
	var isPrivate bool

	// Handle author
    authorSelector := r.PostForm.Get("author-selector")
    if authorSelector != "" {
        authorID, err = strconv.Atoi(authorSelector)
        if err != nil {
            form.AddFieldError("author", "Invalid author selection")
        }
    } else if form.NewAuthorName != "" {
        validator.ValidateAuthor(&form.Validator, form.NewAuthorName)
        if form.ValidField() {
            authorID, err = app.authors.Insert(form.NewAuthorName)
            if err != nil {
                app.serverError(w, r, err)
                return
            }
        }
    } else {
        form.AddFieldError("author", "Please select an author or enter a new one.")
    }

	// Handle book
    bookSelector := r.PostForm.Get("book-selector")
    if bookSelector != "" {
        bookID, err = strconv.Atoi(bookSelector)
        if err != nil {
            form.AddFieldError("book", "Invalid book selection")
        }
    } else if form.NewBookTitle != "" {
        validator.ValidateBook(&form.Validator, form.NewBookTitle, form.NewBookPublishYear, form.NewBookCalendarTime, form.NewBookISBN, form.NewBookSource)
		if form.ValidField() {
			bookID, err = app.books.Insert(form.NewBookTitle, form.NewBookPublishYear, form.NewBookCalendarTime, form.NewBookISBN, form.NewBookSource)
			if err != nil {
				app.serverError(w, r, err)
				return
			}
		}
    } else {
        form.AddFieldError("book", "Please select a book or enter a new one.")
	}

	// Handle page number
	pageNumber = r.PostForm.Get("page_number")
	if pageNumber != "" {
		form.PageNumber = pageNumber
	}

	// Handle isPrivate
	isPrivate = r.PostForm.Get("is_private") == "on"
	form.IsPrivate = isPrivate
	

	// If the form is not valid, re-render the form
    if !form.ValidField() {
        data := app.newTemplateData(r)
        data.Form = form
        data.Quote = originalQuote
        
        authors, err := app.authors.GetAllWithCounts()
        if err != nil {
            app.serverError(w, r, err)
            return
        }

		books, err := app.books.GetAll()
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		data.Authors = authors
		data.Books = books
		data.Quote = originalQuote
		
		app.render(w, r, http.StatusUnprocessableEntity, "edit-quote.go.tmpl", data)
        return
    }

	// Update the quote
    _, err = app.quotes.Update(id, form.Quote, authorID, bookID, form.PageNumber, form.IsPrivate)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

	// Add a flash message
    app.sessionManager.Put(r.Context(), "flash", "Quote updated successfully!")

	// Redirect to the quote view page
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