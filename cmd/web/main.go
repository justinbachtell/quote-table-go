package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"quotetable.com/internal/models"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/supabase-community/supabase-go"
)

// Define struct to hold application-wide dependencies
type application struct {
	logger        *slog.Logger
	quotes        models.QuoteModelInterface
	users         models.UserModelInterface
	templateCache map[string]*template.Template
	client        *supabase.Client
	formDecoder   *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// Define command-line flag for the network address
	addr := flag.String("addr", "127.0.0.1:4000", "HTTP network address")

	// Parse the command-line flag
	flag.Parse()

	// Create a new logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		logger.Error("error loading .env file")
	}

	// Log the environment variables to check if they are loaded correctly
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_PUBLIC_KEY")
	supabaseURIString := os.Getenv("SUPABASE_URI_STRING")

	// Check if the environment variables are set
	if supabaseURL == "" || supabaseKey == "" {
		logger.Error("database environment variables are not set")
		os.Exit(1)
	}

	// Initialize database connection
	db, err := sql.Open("postgres", supabaseURIString)
	if err != nil {
		logger.Error("error connecting to the database for initialization")
		os.Exit(1)
	} else {
		logger.Info("connected to database for initialization")
	}
	// Close the database connection when the main function returns and print a message
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("error closing the database initialization connection")
		} else {
			logger.Info("database initialization connection closed")
		}
	}()

	// Initialize supabase client
	client, err := connectSupabase(logger, supabaseURL, supabaseKey)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	} else {
		logger.Info("connected to database for rest api")
	}

	// TODO; close the supabase client when the main function returns and print a message

	// Initialize template cache
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Initialize form decoder instance
	formDecoder := form.NewDecoder()

	// Initialize session manager and configure it to use postgres store
	sessionManager := scs.New()
	sessionManager.Store = postgresstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// Initialize a new instance of application struct dependencies
	app := &application{
		logger:        logger,
		quotes:        &models.QuoteModel{Client: client},
		users:         &models.UserModel{Client: client},
		templateCache: templateCache,
		client:        client,
		sessionManager: sessionManager,
		formDecoder:   formDecoder,
	}

	// Initialize a tls config struct to configure the tls settings
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:       tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// Initialize a new http server struct
	srv := &http.Server{
		Addr:    *addr,
		Handler: app.routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig: tlsConfig,
		IdleTimeout: time.Minute,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Log the address the server is starting on
	logger.Info("starting server", slog.String("addr", ":4000"))

	// Call the listen and serve method on the http server struct
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")

	// Log any errors that occur
	logger.Error(err.Error())
	os.Exit(1)
}

// Connect to the supabase database
func connectSupabase(logger *slog.Logger, supabaseURL string, supabaseKey string) (*supabase.Client, error) {
	// Initialize supabase client
	client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		logger.Error("error connecting to the database for rest api")
		return nil, err
	}

    // Return the connection is successful
    return client, nil
}
