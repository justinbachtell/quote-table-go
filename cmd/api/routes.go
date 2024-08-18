package main

import (
	"net/http"

	"github.com/justinbachtell/quote-table-go/ui"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

// HTTP router
func (app *application) routes() http.Handler {
	// Initialize a new httprouter instance
	router := httprouter.New()

	// Register the static file server to serve embedded static files
	router.Handler("GET", "/static/*filepath", http.FileServerFS(ui.Files))

	// Register the ping handler for testing
	router.HandlerFunc(http.MethodGet, "/ping", ping)

	// Register the healthcheck handler
	router.HandlerFunc(http.MethodGet, "/healthcheck", app.healthCheck)

	// Create a middleware chain for dynamic routes
	dynamicRouter := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	// Register the unprotected app routes
	router.Handler("GET", "/", dynamicRouter.ThenFunc(app.home))
	router.Handler("GET", "/quote/view/:id", dynamicRouter.ThenFunc(app.quoteView))
	router.Handler("GET", "/user/signup", dynamicRouter.ThenFunc(app.userSignup))
	router.Handler("POST", "/user/signup", dynamicRouter.ThenFunc(app.userSignupPost))
	router.Handler("GET", "/user/login", dynamicRouter.ThenFunc(app.userLogin))
	router.Handler("POST", "/user/login", dynamicRouter.ThenFunc(app.userLoginPost))
	router.Handler("GET", "/terms", dynamicRouter.ThenFunc(app.terms))
	router.Handler("GET", "/privacy", dynamicRouter.ThenFunc(app.privacy))
	router.Handler("GET", "/pricing", dynamicRouter.ThenFunc(app.pricing))

	// Create a middleware chain for dynamic, authenticated routes
	protected := dynamicRouter.Append(app.requireAuthentication)

	// Register the protected app routes
	router.Handler("GET", "/quote/create", protected.ThenFunc(app.quoteCreate))
	router.Handler("POST", "/quote/create", protected.ThenFunc(app.quoteCreatePost))
	router.Handler("GET", "/quote/edit/:id", protected.ThenFunc(app.quoteEdit))
	router.Handler("POST", "/quote/edit/:id", protected.ThenFunc(app.quoteEditPost))
	router.Handler("POST", "/quote/delete/:id", protected.ThenFunc(app.quoteDeletePost))
	router.Handler("POST", "/user/logout", protected.ThenFunc(app.userLogout))
	router.Handler("GET", "/user/profile/edit", protected.ThenFunc(app.userEditProfile))
	router.Handler("POST", "/user/profile/edit", protected.ThenFunc(app.userEditProfilePost))
	router.Handler("GET", "/user/profile/change-password", protected.ThenFunc(app.userChangePassword))
	router.Handler("POST", "/user/profile/change-password", protected.ThenFunc(app.userChangePasswordPost))
	router.Handler("GET", "/user/profile/view/:id", protected.ThenFunc(app.userProfileView))

	// Create middleware chain with standard middleware for every request
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(router)
}

/* // Net/Http default router fallback
func (app *application) routes() http.Handler {
	// Initialize a new http.ServeMux
	mux := http.NewServeMux()

	// Register the static file server to serve embedded static files
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// Register the ping handler for testing
	mux.HandleFunc("GET /ping", ping)

	// Register unprotected app routes

	// Create a middleware chain for dynamic routes
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	mux.Handle("GET /healthcheck", dynamic.ThenFunc(app.healthCheck))

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home))
	mux.Handle("GET /quote/view/{id}", dynamic.ThenFunc(app.quoteView))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("GET /terms", dynamic.ThenFunc(app.terms))
	mux.Handle("GET /privacy", dynamic.ThenFunc(app.privacy))
	mux.Handle("GET /pricing", dynamic.ThenFunc(app.pricing))
	
	// Register protected app routes

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
} */