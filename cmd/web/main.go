package main

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/dangoodie/snippetbox/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type application struct {
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	_ = godotenv.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "", "MySQL data source name")
	flag.Parse()

	if *dsn == "" {
		err := setDefaultDSN(dsn)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	db, err := openDb(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManger := scs.New()
	sessionManger.Store = mysqlstore.New(db)
	sessionManger.Lifetime = 12 * time.Hour
	sessionManger.Cookie.Secure = true

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManger,
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("Starting server", "addr", *addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

func openDb(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setDefaultDSN(dsn *string) error {
	dbPass, ok := os.LookupEnv("MYSQL_PASSWORD")
	if !ok {
		return errors.New("MYSQL_PASSWORD environment variable is not set")

	}
	*dsn = fmt.Sprintf("web:%s@/snippetbox?parseTime=true", dbPass)
	return nil
}
