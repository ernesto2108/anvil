package query

import "database/sql"

const defaultListLimit = 100

// Reader provides read-only access to the dashboard database.
type Reader struct {
	db *sql.DB
}

// NewReader creates a Reader wrapping the given *sql.DB.
func NewReader(db *sql.DB) *Reader {
	return &Reader{db: db}
}

// DB returns the underlying *sql.DB for use in tests.
func (r *Reader) DB() *sql.DB {
	return r.db
}
