package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func OpenPostgres(dsn string) (*sql.DB, func() error, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, nil, err
	}
	return db, db.Close, nil
}
