package puppetest

import (
	"context"
	"database/sql"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

type (
	// DBFactory is an alias for the engine database connection factory.
	DBFactory = dbastidor.ConnectionFactory
	// DBConnectionConfig is an alias for database connection creation inputs.
	DBConnectionConfig = dbastidor.ConnectionConfig
	// ConnectionPerformer is an alias for the low-level connection creator function.
	ConnectionPerformer = dbastidor.ConnectionPerformer
)

// ConnectDB adapts a config-only connector into a ConnectionPerformer.
func ConnectDB(
	connector func(conf DBConnectionConfig) (*sql.DB, error),
) dbastidor.ConnectionPerformer {
	return func(_ context.Context, conf dbastidor.ConnectionConfig) (*sql.DB, error) {
		return connector(conf)
	}
}

// ConnectDBFromDSN builds a ConnectionPerformer that opens a DB using a generated DSN.
func ConnectDBFromDSN(
	driver string, dsnGen func(conf DBConnectionConfig) string,
) dbastidor.ConnectionPerformer {
	return func(_ context.Context, conf dbastidor.ConnectionConfig) (*sql.DB, error) {
		dsn := dsnGen(conf)
		return sql.Open(driver, dsn)
	}
}
