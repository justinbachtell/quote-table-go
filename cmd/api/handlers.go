package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/justinbachtell/quote-table-go/internal/models"
	"github.com/justinbachtell/quote-table-go/internal/validator"
)

// Handler for the home page
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	// Get the latest quotes from the database
	quotes, err := app.quotes.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Get the template data and add the quotes slice
	data := app.newTemplateData(r)
	data.Quotes = quotes

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

// Struct to represent the quote form data
type quoteCreateForm struct {
	Quote string `form:"quote"`
	Author string `form:"author"`
	Created time.Time `form:"created"`
	validator.Validator `form:"-"`
}

// Handler for the view quote page
func (app *application) quoteView(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Path[len("/quote/view/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

    quote, err := app.quotes.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.logger.Error(err.Error())
            http.Error(w, err.Error(), http.StatusInternalServerError)
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
    var form quoteCreateForm

    err := app.formDecoder.Decode(&form, r.PostForm)
    if err != nil {
        app.logger.Error(err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    form.CheckField(validator.NotBlank(form.Quote), "quote", "The quote field cannot be blank.")
    form.CheckField(validator.MaxChars(form.Quote, 19000), "quote", "The quote field is too long (max. 19,000 characters).")
    form.CheckField(validator.NoInvalidCharacters(form.Quote), "quote", "The quote field contains invalid characters.")

    form.CheckField(validator.NotBlank(form.Author), "author", "The author field cannot be blank.")
    form.CheckField(validator.MaxChars(form.Author, 100), "author", "The author field is too long (max. 100 characters).")
    form.CheckField(validator.NoInvalidCharacters(form.Author), "author", "The author field contains invalid characters.")

    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "create-quote.go.tmpl", data)
        return
    }

    id, err := app.quotes.Insert(form.Quote, form.Author)
    if err != nil {
		app.logger.Error(err.Error())
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
        http.NotFound(w, r)
        return
    }

    quote, err := app.quotes.Get(id)
    if err != nil {
        if errors.Is(err, models.ErrNoRecord) {
            http.NotFound(w, r)
        } else {
            app.logger.Error(err.Error())
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
            app.logger.Error(err.Error())
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    var form quoteCreateForm

    err = app.formDecoder.Decode(&form, r.PostForm)
    if err != nil {
        app.logger.Error(err.Error())
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    form.CheckField(validator.NotBlank(form.Quote), "quote", "The quote field cannot be blank.")
    form.CheckField(validator.MaxChars(form.Quote, 19000), "quote", "The quote field is too long (max. 19,000 characters).")
    form.CheckField(validator.NoInvalidCharacters(form.Quote), "quote", "The quote field contains invalid characters.")

    form.CheckField(validator.NotBlank(form.Author), "author", "The author field cannot be blank.")
    form.CheckField(validator.MaxChars(form.Author, 100), "author", "The author field is too long (max. 100 characters).")
    form.CheckField(validator.NoInvalidCharacters(form.Author), "author", "The author field contains invalid characters.")

    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "edit-quote.go.tmpl", data)
        return
    }

    _, err = app.quotes.Update(id, form.Quote, form.Author)
    if err != nil {
        app.logger.Error(err.Error())
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
        app.logger.Error(err.Error())
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Quote deleted successfully")

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Struct to represent signup form data
type userSignupForm struct {
	Name string `form:"name"`
	Email string `form:"email"`
	Password string `form:"password"`
	validator.Validator `form:"-"`
}

// Handler for the user signup page
func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.go.tmpl", data)
}

// Handler to process and and post the user data
func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	// Declare an empty instance
	var form userSignupForm

	// Parse the form data into the struct
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form data
	form.CheckField(validator.NotBlank(form.Name), "name", "The name field cannot be blank.")
	form.CheckField(validator.MaxChars(form.Name, 100), "name", "The name field is too long (max. 100 characters).")
	form.CheckField(validator.MinChars(form.Name, 2), "name", "The name field is too short (min. 2 characters).")
	form.CheckField(validator.NoInvalidCharacters(form.Name), "name", "The name field contains invalid characters.")
	form.CheckField(validator.NotBlank(form.Email), "email", "The email field cannot be blank.")
	form.CheckField(validator.MaxChars(form.Email, 255), "email", "The email field is too long (max. 255 characters).")
	form.CheckField(validator.MinChars(form.Email, 5), "email", "The email field is too short (min. 5 characters).")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "The email field is not a valid email address.")
	form.CheckField(validator.NotBlank(form.Password), "password", "The password field cannot be blank.")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "The password field is too short (min. 8 characters).")
	form.CheckField(validator.MaxChars(form.Password, 70), "password", "The password field is too long (max. 70 characters).")
	form.CheckField(validator.NoInvalidCharacters(form.Password), "password", "The password field contains invalid characters.")

	// If there are errors, render the form again with the errors
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.go.tmpl", data)
		return
	}

	// Pass the data to insert method and get the ID of the new user
	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "This email address is already in use.")
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.go.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Flash success message
	app.sessionManager.Put(r.Context(), "flash", "User created successfully. Please log in to continue.")
	
	// Redirect to the login page
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

// Struct to represent login form data
type userLoginForm struct {
	Email string `form:"email"`
	Password string `form:"password"`
	validator.Validator `form:"-"`
}

// Handler for the user login page
func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.go.tmpl", data)
}

// Handler to process and and post the user data
func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	// Decode the form data
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// Validate the form data
	form.CheckField(validator.NotBlank(form.Email), "email", "The email field cannot be blank.")
	form.CheckField(validator.MaxChars(form.Email, 255), "email", "The email field is too long (max. 255 characters).")
	form.CheckField(validator.MinChars(form.Email, 5), "email", "The email field is too short (min. 5 characters).")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "The email field is not a valid email address.")
	form.CheckField(validator.NotBlank(form.Password), "password", "The password field cannot be blank.")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "The password field is too short (min. 8 characters).")
	form.CheckField(validator.MaxChars(form.Password, 70), "password", "The password field is too long (max. 70 characters).")

	// If there are errors, render the form again with the errors
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.go.tmpl", data)
		return
	}

	// Check if the credentials are valid
	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("Email or password is incorrect")

			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.go.tmpl", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// Use renew token to change the session token
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	log.Printf("Renewed session token: %v", id)

	// Add the user ID to the session
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)
	
	// Redirect to the home page
	http.Redirect(w, r, "/quote/create", http.StatusSeeOther)
}

// Handler for the user logout page
func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
	// Renew the session token
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Remove the authenticated session
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	// Flash success message to confirm logout
	app.sessionManager.Put(r.Context(), "flash", "You have been logged out successfully")

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handler for the user profile page
func (app *application) userProfileView(w http.ResponseWriter, r *http.Request) {
    // Get the id from the URL
    idStr := r.URL.Path[len("/user/profile/"):]
    id, err := strconv.Atoi(idStr)
    if err != nil || id < 1 {
        http.NotFound(w, r)
        return
    }

	// Fetch the user profile
	user, err := app.users.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.User = &user
	data.AuthenticatedUserID = app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	app.render(w, r, http.StatusOK, "profile.go.tmpl", data)
}

// Handler for the user profile edit page
func (app *application) userEditProfile(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)

    // Fetch the authenticated user
    userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
    user, err := app.users.Get(userID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    // If the user is not authenticated, redirect to the login page
    if userID == 0 {
        http.Redirect(w, r, "/user/login", http.StatusSeeOther)
        return
    }

    // Initialize the form with the current user data
    data.Form = userProfileForm{
        Name:  user.Name,
        Email: user.Email,
    }

    // Render the form
    data.User = &user
    app.render(w, r, http.StatusOK, "edit-profile.go.tmpl", data)
}

type userProfileForm struct {
    Name     string `form:"name"`
    Email    string `form:"email"`
    validator.Validator
}

// Handler to update the user profile
func (app *application) userEditProfilePost(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)

    if data.User == nil {
        http.Redirect(w, r, "/user/login", http.StatusSeeOther)
        return
    }

    var form userProfileForm

    err := app.decodePostForm(r, &form)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
        return
    }

    // Validate the form data
    form.CheckField(validator.NotBlank(form.Name), "name", "The name field cannot be blank.")
    form.CheckField(validator.MaxChars(form.Name, 100), "name", "The name field is too long (max. 100 characters).")
    form.CheckField(validator.MinChars(form.Name, 2), "name", "The name field is too short (min. 2 characters).")
    form.CheckField(validator.NoInvalidCharacters(form.Name), "name", "The name field contains invalid characters.")
    form.CheckField(validator.NotBlank(form.Email), "email", "The email field cannot be blank.")
    form.CheckField(validator.MaxChars(form.Email, 255), "email", "The email field is too long (max. 255 characters).")
    form.CheckField(validator.MinChars(form.Email, 5), "email", "The email field is too short (min. 5 characters).")
    form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "The email field is not a valid email address.")

    if !form.Valid() {
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "edit-profile.go.tmpl", data)
        return
    }

    err = app.users.Update(data.User.ID, form.Name, form.Email)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Profile updated successfully")

    http.Redirect(w, r, fmt.Sprintf("/user/profile/%d", data.User.ID), http.StatusSeeOther)
}

// Add this struct at the top of the file with other form structs
type userChangePasswordForm struct {
    CurrentPassword    string `form:"currentPassword"`
    NewPassword        string `form:"newPassword"`
    ConfirmPassword    string `form:"confirmPassword"`
    validator.Validator `form:"-"`
}

// Handler for the user change password page
func (app *application) userChangePassword(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    
	// Fetch the authenticated user
	userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// If the user is not authenticated, redirect to the login page
    if userID == 0 {
        http.Redirect(w, r, "/user/login", http.StatusSeeOther)
        return
    }

    // Initialize the form
	data.Form = userChangePasswordForm{
		CurrentPassword: "",
		NewPassword: "",
		ConfirmPassword: "",
	}

	// Render the form
	data.User = &user
	app.render(w, r, http.StatusOK, "change-password.go.tmpl", data)
}

// Add this function to the handlers.go file
func (app *application) userChangePasswordPost(w http.ResponseWriter, r *http.Request) {
    var form userChangePasswordForm

    err := app.decodePostForm(r, &form)
    if err != nil {
        app.clientError(w, http.StatusBadRequest)
        return
    }

    form.CheckField(validator.NotBlank(form.CurrentPassword), "currentPassword", "This field cannot be blank")
    form.CheckField(validator.NotBlank(form.NewPassword), "newPassword", "This field cannot be blank")
    form.CheckField(validator.MinChars(form.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
    form.CheckField(validator.NotBlank(form.ConfirmPassword), "confirmPassword", "This field cannot be blank")
    form.CheckField(form.NewPassword == form.ConfirmPassword, "confirmPassword", "Passwords do not match")

    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "change-password.go.tmpl", data)
        return
    }

    userID := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

    err = app.users.ChangePassword(userID, form.CurrentPassword, form.NewPassword)
    if err != nil {
        if errors.Is(err, models.ErrInvalidCredentials) {
            form.AddFieldError("currentPassword", "Current password is incorrect")
            data := app.newTemplateData(r)
            data.Form = form
            app.render(w, r, http.StatusUnprocessableEntity, "change-password.go.tmpl", data)
        } else {
            app.serverError(w, r, err)
        }
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Your password has been changed successfully")

    http.Redirect(w, r, fmt.Sprintf("/user/profile/%d", userID), http.StatusSeeOther)
}

// Handler to return a 200 OK status code
func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
	
	r.Body = http.MaxBytesReader(w, r.Body, 100)
}

// Handler to return application health status
func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}