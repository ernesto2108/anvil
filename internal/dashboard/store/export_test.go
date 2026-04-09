package store

import "database/sql"

// TestableStore wraps SQLiteStore to expose internals for black-box tests.
type TestableStore struct {
	*SQLiteStore
}

// NewTestableStore creates a TestableStore from an SQLiteStore.
func NewTestableStore(s *SQLiteStore) *TestableStore {
	return &TestableStore{SQLiteStore: s}
}

// DB returns the underlying *sql.DB for direct queries in tests.
func (ts *TestableStore) DB() *sql.DB {
	return ts.db
}

// MaxRuns returns the configured maxRuns value.
func (ts *TestableStore) MaxRuns() int {
	return ts.maxRuns
}
