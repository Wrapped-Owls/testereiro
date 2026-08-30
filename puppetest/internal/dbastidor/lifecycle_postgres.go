package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const (
	postgresIdentifierQuote = '"'
	// duplicateDatabaseSQLState is SQLSTATE 42P04, raised when CREATE DATABASE loses a race.
	duplicateDatabaseSQLState = "42P04"

	postgresExistsStmt    = "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)"
	postgresTerminateStmt = "SELECT pg_terminate_backend(pid) FROM pg_stat_activity " +
		"WHERE datname = $1 AND pid <> pg_backend_pid()"
)

// pgx's *pgconn.PgError satisfies this, so no driver dependency is needed. lib/pq's *pq.Error does
// not, leaving racing Create calls unsafe on that driver.
type sqlStateCoder interface {
	SQLState() string
}

type PostgresLifecycle struct {
	rootDB *sql.DB
}

var _ DBLifecycle = (*PostgresLifecycle)(nil)

func NewPostgresLifecycle(rootDB *sql.DB) *PostgresLifecycle {
	return &PostgresLifecycle{rootDB: rootDB}
}

func (l *PostgresLifecycle) Create(ctx context.Context, dbName string) error {
	quoted, err := QuoteIdentifier(dbName, postgresIdentifierQuote)
	if err != nil {
		return fmt.Errorf("quote postgres database name: %w", err)
	}

	// Postgres has no CREATE DATABASE ... IF NOT EXISTS, and rejects it inside a transaction.
	var alreadyExists bool
	existsRow := l.rootDB.QueryRowContext(ctx, postgresExistsStmt, dbName)
	if err = existsRow.Scan(&alreadyExists); err != nil {
		return fmt.Errorf("check postgres database existence: %w", err)
	}
	if alreadyExists {
		return nil
	}

	// #nosec G201: a statement cannot bind an identifier as a parameter; QuoteIdentifier escapes it.
	if _, err = l.rootDB.ExecContext(ctx, "CREATE DATABASE "+quoted); err != nil {
		if isDuplicateDatabase(err) {
			return nil
		}
		return fmt.Errorf("create postgres database: %w", err)
	}

	return nil
}

func (l *PostgresLifecycle) Drop(ctx context.Context, dbName string) error {
	quoted, err := QuoteIdentifier(dbName, postgresIdentifierQuote)
	if err != nil {
		return fmt.Errorf("quote postgres database name: %w", err)
	}

	// Postgres refuses to drop a database holding live connections; WITH (FORCE) would do the same
	// but restrict this to server 13 and up.
	if _, err = l.rootDB.ExecContext(ctx, postgresTerminateStmt, dbName); err != nil {
		return fmt.Errorf("terminate postgres backends: %w", err)
	}

	// #nosec G201: a statement cannot bind an identifier as a parameter; QuoteIdentifier escapes it.
	if _, err = l.rootDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoted); err != nil {
		return fmt.Errorf("drop postgres database: %w", err)
	}

	return nil
}

func isDuplicateDatabase(err error) bool {
	var coder sqlStateCoder
	return errors.As(err, &coder) && coder.SQLState() == duplicateDatabaseSQLState
}
