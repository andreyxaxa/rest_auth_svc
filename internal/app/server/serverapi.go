package server

import (
	"database/sql"
	"net/http"

	sqlstorage "github.com/andreyxaxa/rest_auth_svc/internal/app/storage/postgre"
)

func Start(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}

	storage := sqlstorage.New(db)
	srv := newServer(storage, config.JwtSecretKey)

	return http.ListenAndServe(config.Addr, srv.router)
}

func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
