# Quote Table

Quote Table is a Go-based web application for managing and displaying quotes. It uses Supabase for data storage and includes features like user authentication, quote creation, and viewing.

## 🚀 Features

- User authentication (signup, login, logout, edit profile)
- Create and view quotes
- Responsive web design
- Integration with Supabase for data storage
- Secure session management

## 🛠️ Technology Stack

- **Backend:** Go 1.23.0
- **Database:** Supabase (PostgreSQL)
- **Frontend:** HTML templates with Go's `html/template` package, and HTMX
- **CSS:** TailwindCSS with some custom CSS styles

## 📁 Project Structure

- `cmd/api/`: Main application code
- `internal/models/`: Database models and interfaces
- `internal/validator/`: Input validation logic
- `ui/`: HTML templates and static assets

## 🏗️ Setup and Installation

1. Clone the repository
2. Set up a Supabase project and obtain the URL and API key
3. Create a `.env` file in the project root with the following content:
   ```
   SUPABASE_URL=your_supabase_url
   SUPABASE_KEY=your_supabase_api_key
   ```
4. Run `go mod tidy` to install dependencies
5. Start the application with `go run ./cmd/api` or run `air` to run a server with live reloads.
6. Start the TailwindCSS watcher with `./tailwindcss -i ui/static/css/twinput.css -o ui/static/css/twoutput.css --watch`
7. Start the Sass watcher with `sass --watch ui/static/sass/globals.scss ui/static/css/globals.css --style compressed`
8. Compile and minify the CSS for production with `./tailwindcss -i ui/static/css/twinput.css -o ui/static/css/twoutput.css --minify`

## 🧪 Running Tests

## Test Types

- To run tests, it is necessary to first be running Docker to set up a local Supabase instance for testing

| Test Type | Command | Description |
|-----------|---------|-------------|
| All Tests | `go test -v ./...` | Runs all tests in the project |
| Short Tests | `go test -v -short ./...` | Skips long-running tests |
| Unit & E2E Tests | `go test -v ./cmd/api` | Runs tests in the `cmd/api` package |
| Integration Tests | `go test -v ./internal/models` | Runs tests in the `internal/models` package |

- 💡 **Tip 1:** Use the `-v` flag for verbose output in all test commands.
- 💡 **Tip 2:** Use the `-cover` flag to generate metrics for code test coverage.
- 💡 **Tip 3:** Use the `-coverprofile=/tmp/profile.out` or `-covermode=count -coverprofile=/tmp/profile.out` flag to generate a detailed breakdown of code test coverage by method and function, and then view the coverage profile with either of the two commands:
- - `go tool cover -html=/tmp/profile.out` (visual)
- - `go tool cover -func=/tmp/profile.out` (terminal)

## 📚 API Documentation

### Healthcheck
- `GET /healthcheck`: Display the application status, environment, and version

### Quotes

- `GET /`: Home page, displays latest quotes
- `GET /quote/view/:id`: View a specific quote
- `GET /quote/create`: Display the quote creation form
- `POST /quote/create`: Submit a new quote
- `GET /quote/edit/:id`: Display the edit quote form
- `POST /quote/edit:id`: Edit a quote

### Users

- `GET /user/signup`: Display user signup form
- `POST /user/signup`: Create a new user account
- `GET /user/login`: Display user login form
- `POST /user/login`: Authenticate user
- `POST /user/logout`: Log out the current user
- `GET /user/profile/:id`: View a user's profile
- `GET /user/profile/edit`: View the edit profile form
- `POST /user/profile/edit`: Edit the user's profile
- `GET /user/profile/change-password`: View the change password form
- `POST /user/profile/change-password`: Edit the user's password