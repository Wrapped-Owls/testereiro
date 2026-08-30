package puppetest

import (
	"database/sql"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

type (
	// DBLifecycle is implemented to support an engine puppetest ships no lifecycle for.
	DBLifecycle = dbastidor.DBLifecycle
	// DBLifecycleBuilder receives a root connection, which must target a database other than the
	// ones the lifecycle then creates and drops.
	DBLifecycleBuilder = dbastidor.LifecycleBuilder
)

var ErrInvalidIdentifier = dbastidor.ErrInvalidIdentifier

var (
	_ DBLifecycleBuilder = MySQLLifecycle
	_ DBLifecycleBuilder = PostgresLifecycle
)

// QuoteIdentifier doubles the quote rune so a crafted database name cannot close the identifier
// and append its own statement. Assumes a symmetric delimiter, as MySQL, Postgres and SQLite use.
func QuoteIdentifier(name string, quote rune) (string, error) {
	return dbastidor.QuoteIdentifier(name, quote)
}

//nolint:ireturn // the DBLifecycle return is what makes this a DBLifecycleBuilder value.
func MySQLLifecycle(rootDB *sql.DB) DBLifecycle {
	return dbastidor.NewMySQLLifecycle(rootDB)
}

// PostgresLifecycle needs a root connection on another database, conventionally postgres or
// template1; concurrent creates of one name stay race-safe only on pgx.
//
//nolint:ireturn // the DBLifecycle return is what makes this a DBLifecycleBuilder value.
func PostgresLifecycle(rootDB *sql.DB) DBLifecycle {
	return dbastidor.NewPostgresLifecycle(rootDB)
}
