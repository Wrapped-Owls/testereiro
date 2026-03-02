package testkit

import (
	"database/sql"
	"testing"
)

func OpenStubDB(t testing.TB, dsn string, state *SQLState) *sql.DB {
	t.Helper()
	EnsureSQLDriver(t)
	RegisterSQLState(dsn, state)

	db, err := sql.Open(StubSQLDriverName, dsn)
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	return db
}
