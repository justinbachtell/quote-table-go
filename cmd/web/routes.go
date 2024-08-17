package main

import (
	"net/http"

	"quotetable.com/ui"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// Initialize a new http.ServeMux
	mux := http.NewServeMux()

	// Register the static file server to serve embedded static files
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// Register the ping handler for testing
	mux.HandleFunc("GET /ping", ping)

	/*
	* Register unprotected app routes 
	*/
	// Create a middleware chain for dynamic routes
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /quote/view/{id}", dynamic.ThenFunc(app.quoteView))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("GET /terms", dynamic.ThenFunc(app.terms))
	mux.Handle("GET /privacy", dynamic.ThenFunc(app.privacy))
	mux.Handle("GET /pricing", dynamic.ThenFunc(app.pricing))
	
	/*
	* Register protected app routes 
	*/
	// Create a middleware chain for dynamic, authenticated routes
	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /quote/create", protected.ThenFunc(app.quoteCreate))
	mux.Handle("POST /quote/create", protected.ThenFunc(app.quoteCreatePost))
	mux.Handle("GET /quote/edit/{id}", protected.ThenFunc(app.quoteEdit))
	mux.Handle("POST /quote/edit/{id}", protected.ThenFunc(app.quoteEditPost))
	mux.Handle("POST /quote/delete/{id}", protected.ThenFunc(app.quoteDeletePost))
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogout))
	mux.Handle("GET /user/profile/{id}", protected.ThenFunc(app.userProfileView))
	mux.Handle("GET /user/profile/edit", protected.ThenFunc(app.userEditProfile))
	mux.Handle("POST /user/profile/edit", protected.ThenFunc(app.userEditProfilePost))
	mux.Handle("GET /user/profile/change-password", protected.ThenFunc(app.userChangePassword))
	mux.Handle("POST /user/profile/change-password", protected.ThenFunc(app.userChangePasswordPost))	

	// Create middleware chain with standard middleware for every request
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}