package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/justinbachtell/quote-table-go/internal/models"
	"github.com/justinbachtell/quote-table-go/internal/validator"
)

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

    // Decode the form data
	err := app.decodePostForm(r, &form)
    if err != nil {
        app.logger.Error("Failed to decode post form", "error", err)
        app.clientError(w, http.StatusBadRequest)
        return
    }

    validator.ValidateName(&form.Validator, form.Name)
    validator.ValidateEmail(&form.Validator, form.Email)
    validator.ValidatePassword(&form.Validator, form.Password)

    // If there are errors, render the form again with the errors
	if !form.ValidField() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "signup.go.tmpl", data)
        return
    }

	// Pass the form data to the users model to insert the user
    _, err = app.users.Insert(form.Name, form.Email, form.Password)
    if err != nil {
        app.logger.Error("Failed to insert user", "error", err)
        switch {
        case errors.Is(err, models.ErrDuplicateEmail):
            form.AddFieldError("email", "Email address is already in use")
            data := app.newTemplateData(r)
            data.Form = form
            app.render(w, r, http.StatusUnprocessableEntity, "signup.go.tmpl", data)
        default:
            app.serverError(w, r, err)
        }
        return
    }

	// Flash success message to confirm user creation
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
    validator.ValidateEmail(&form.Validator, form.Email)
    validator.ValidatePassword(&form.Validator, form.Password)

    // If there are errors, render the form again with the errors
    if !form.ValidField() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "login.go.tmpl", data)
        return
    }

    // Check if the credentials are valid
    id, err := app.users.Authenticate(form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrInvalidCredentials) {
            form.AddNonFieldError("Authentication failed. Please check your credentials and try again.")

            data := app.newTemplateData(r)
            data.Form = form
            app.render(w, r, http.StatusUnprocessableEntity, "login.go.tmpl", data)
        } else {
            app.logger.Error("Failed to authenticate user", "error", err)
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

    // Add the user ID to the session
    app.sessionManager.Put(r.Context(), "authenticatedUserID", id.String())
    app.logger.Info("User logged in", "userID", id)

    // Redirect to the home page
    http.Redirect(w, r, "/", http.StatusSeeOther)
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
    // Get the name from the URL
    urlName := r.URL.Path[len("/user/profile/view/"):]
    if urlName == "" {
        http.NotFound(w, r)
        return
    }

    // Fetch the user profile
    user, err := app.users.GetByURLName(urlName)
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

    app.render(w, r, http.StatusOK, "profile.go.tmpl", data)
}

// Handler for the user profile edit page
func (app *application) userEditProfile(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)

    // Fetch the authenticated user
    user, err := app.users.Get(data.AuthenticatedUserID)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

	// Convert the authenticated user ID to a UUID
	user.ID, err = app.convertStringToUUID(app.sessionManager.GetString(r.Context(), "authenticatedUserID"))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

    // If the user is not authenticated, redirect to the login page
    if user.ID == uuid.Nil {
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
	Phone    string `form:"phone"`
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
	validator.ValidateName(&form.Validator, form.Name)
	validator.ValidateEmail(&form.Validator, form.Email)
	validator.ValidatePhone(&form.Validator, form.Phone)
    if !form.ValidField() {
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "edit-profile.go.tmpl", data)
        return
    }

    err = app.users.Update(data.User.ID, form.Name, form.Email, form.Phone)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    app.sessionManager.Put(r.Context(), "flash", "Profile updated successfully")

    http.Redirect(w, r, fmt.Sprintf("/user/profile/view/%s", data.User.ProfileSlug), http.StatusSeeOther)
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
	userID, err := app.convertStringToUUID(app.sessionManager.GetString(r.Context(), "authenticatedUserID"))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	
	// If the user is not authenticated, redirect to the login page
    if userID == uuid.Nil {
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

	// Validate the form
	validator.ValidatePassword(&form.Validator, form.CurrentPassword)
	validator.ValidatePassword(&form.Validator, form.NewPassword)
	validator.ValidatePassword(&form.Validator, form.ConfirmPassword)

    if !form.ValidField() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "change-password.go.tmpl", data)
        return
    }

    // Convert the authenticated user ID to a UUID
	userID, err := app.convertStringToUUID(app.sessionManager.GetString(r.Context(), "authenticatedUserID"))
	if err != nil {
		app.serverError(w, r, err)
		return
	}

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

    http.Redirect(w, r, fmt.Sprintf("/user/profile/%s", userID), http.StatusSeeOther)
}