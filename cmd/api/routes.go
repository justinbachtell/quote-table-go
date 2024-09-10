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

	// Custom 404/405 handlers
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

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
	router.Handler("GET", "/authors", dynamicRouter.ThenFunc(app.authorList))
	router.Handler("GET", "/author/view/:id", dynamicRouter.ThenFunc(app.authorView))
	router.Handler("GET", "/books", dynamicRouter.ThenFunc(app.bookList))
	router.Handler("GET", "/book/view/:id", dynamicRouter.ThenFunc(app.bookView))
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
	//router.Handler("GET", "/author/create", protected.ThenFunc(app.authorCreate))
	//router.Handler("POST", "/author/create", protected.ThenFunc(app.authorCreatePost))
	router.Handler("GET", "/book/create", protected.ThenFunc(app.bookCreate))
	router.Handler("POST", "/book/create", protected.ThenFunc(app.bookCreatePost))
	router.Handler("GET", "/book/edit/:id", protected.ThenFunc(app.bookEdit))
	router.Handler("POST", "/book/edit/:id", protected.ThenFunc(app.bookEditPost))
	router.Handler("POST", "/book/delete/:id", protected.ThenFunc(app.bookDeletePost))
	router.Handler("POST", "/user/logout", protected.ThenFunc(app.userLogout))
	router.Handler("GET", "/user/profile/edit", protected.ThenFunc(app.userEditProfile))
	router.Handler("POST", "/user/profile/edit", protected.ThenFunc(app.userEditProfilePost))
	router.Handler("GET", "/user/profile/change-password", protected.ThenFunc(app.userChangePassword))
	router.Handler("POST", "/user/profile/change-password", protected.ThenFunc(app.userChangePasswordPost))
	router.Handler("GET", "/user/profile/view/:urlName", protected.ThenFunc(app.userProfileView))

	// Create middleware chain with standard middleware for every request
	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(router)
}