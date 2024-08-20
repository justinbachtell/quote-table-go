package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/justinas/nosurf"
)

// Middleware to add common headers to the response
func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

// Logs HTTP requests to the console
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip = r.RemoteAddr
			proto = r.Proto
			method = r.Method
			uri = r.URL.RequestURI()
		)

		app.logger.Info("Received request", "ip", ip, "proto", proto, "method", method, "uri", uri)
		next.ServeHTTP(w, r)
	})
}

// Recover from panic and return a 500 Internal Server Error
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Set the connection to close
				w.Header().Set("Connection", "close")
				// Return a 500 Internal Server Error
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Requires authentication for all requests
func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		// Prevent caching in the user's browser
		w.Header().Add("Cache-Control", "no-store")

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// NoSurf middleware to protect against CSRF attacks
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: true,
	})
	return csrfHandler
}

// Authenticates the user and adds the user to the request context
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
		if idStr == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Parse the string as a UUID
		id, err := uuid.Parse(idStr)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// Use the UUID to check if the user exists
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// If the user exists, add the user to the request context
		if exists {
			// Set the actual user ID in the context
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, id)
			r = r.WithContext(ctx)

			// Update the UserModel and QuoteModel with the authenticated user ID
			app.users.SetAuthUserID(id)
			app.quotes.SetAuthUserID(id)
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}