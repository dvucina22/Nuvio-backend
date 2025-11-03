package repository

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectPostgres(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}
	return db
}
