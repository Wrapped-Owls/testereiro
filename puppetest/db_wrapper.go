package puppetest

import (
	"database/sql"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

// DBWrapper stores a normalized database name and its connection.
type DBWrapper struct {
	name string
	conn *sql.DB
}

// NewDBWrapper creates a DBWrapper with a normalized database name.
func NewDBWrapper(name string, conn *sql.DB) *DBWrapper {
	return &DBWrapper{name: dbastidor.NormalizeDBName(name), conn: conn}
}

// Teardown closes the wrapped database connection.
func (dw *DBWrapper) Teardown() error {
	if dw.conn == nil {
		return nil
	}
	closeErr := dw.conn.Close()
	return closeErr
}

// Connection returns the wrapped database handle.
func (dw *DBWrapper) Connection() *sql.DB {
	return dw.conn
}

// IsZero reports whether the wrapper has no database connection.
func (dw *DBWrapper) IsZero() bool {
	return dw.conn == nil
}
