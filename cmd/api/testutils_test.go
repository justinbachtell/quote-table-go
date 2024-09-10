package main

import (
	"bytes"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/justinbachtell/quote-table-go/internal/models/mocks"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

// Create a new application struct with mocked dependencies
func newTestApplication(t *testing.T) *application {
	// Create an instance of the template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	// Create an instance of the form decoder
	formDecoder := form.NewDecoder()

	// Create an instance of the session manager
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		quotes: &mocks.QuoteModel{},
		authors: &mocks.AuthorModel{},
		books: &mocks.BookModel{},
		users: &mocks.UserModel{},
		templateCache: templateCache,
		formDecoder: formDecoder,
		sessionManager: sessionManager,
	}
}

// Define a custom test server struct that embeds a httptest.Server instance
type testServer struct {
	*httptest.Server
}

// Create a new test server instance
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	// Initialize a cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the test server client
	ts.Client().Jar = jar

	// Disable redirect-following for the test server client for 3xx responses
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// Implement a GET method on the test server struct
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	// Close the response body when the function returns
	defer rs.Body.Close()

	// Read the response body
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Return the status code, headers, and body
	return rs.StatusCode, rs.Header, string(body)
}

// Defines a regexp to capture the CSRF token from the signup page
var csrfTokenRX = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)">`)

// Extracts the CSRF token from the body of a response
func extractCSRFToken(t *testing.T, body string) string {
	// Extract the CSRF token from the body as an array
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no CSRF token found in body")
	}
	return html.UnescapeString(matches[1])
}

// Posts a form to the test server and returns the response
func (ts *testServer) postForm(t *testing.T, urlPath string, formValues url.Values) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, formValues)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Trim the whitespace from the body
	body = bytes.TrimSpace(body)


	// Return the status code, headers, and body
	return rs.StatusCode, rs.Header, string(body)
}