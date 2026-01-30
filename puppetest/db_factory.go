package puppetest

import (
	"context"
	"database/sql"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
)

type (
	DBFactory          = dbastidor.ConnectionFactory
	DBConnectionConfig = dbastidor.ConnectionConfig
)

func ConnectDB(
	connector func(conf DBConnectionConfig) (*sql.DB, error),
) dbastidor.ConnectionPerformer {
	return func(_ context.Context, conf dbastidor.ConnectionConfig) (*sql.DB, error) {
		return connector(conf)
	}
}

func ConnectDBFromDSN(
	driver string, dsnGen func(conf DBConnectionConfig) string,
) dbastidor.ConnectionPerformer {
	return func(_ context.Context, conf dbastidor.ConnectionConfig) (*sql.DB, error) {
		dsn := dsnGen(conf)
		return sql.Open(driver, dsn)
	}
}
