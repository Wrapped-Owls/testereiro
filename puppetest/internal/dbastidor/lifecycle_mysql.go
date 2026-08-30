package dbastidor

import (
	"context"
	"database/sql"
	"fmt"
)

const mysqlIdentifierQuote = '`'

type MySQLLifecycle struct {
	rootDB *sql.DB
}

var _ DBLifecycle = (*MySQLLifecycle)(nil)

func NewMySQLLifecycle(rootDB *sql.DB) *MySQLLifecycle {
	return &MySQLLifecycle{rootDB: rootDB}
}

func (l *MySQLLifecycle) Create(ctx context.Context, dbName string) error {
	quoted, err := QuoteIdentifier(dbName, mysqlIdentifierQuote)
	if err != nil {
		return fmt.Errorf("quote mysql database name: %w", err)
	}

	// #nosec G201: a statement cannot bind an identifier as a parameter; QuoteIdentifier escapes it.
	if _, err = l.rootDB.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+quoted); err != nil {
		return fmt.Errorf("create mysql database: %w", err)
	}

	return nil
}

func (l *MySQLLifecycle) Drop(ctx context.Context, dbName string) error {
	quoted, err := QuoteIdentifier(dbName, mysqlIdentifierQuote)
	if err != nil {
		return fmt.Errorf("quote mysql database name: %w", err)
	}

	// #nosec G201: a statement cannot bind an identifier as a parameter; QuoteIdentifier escapes it.
	if _, err = l.rootDB.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoted); err != nil {
		return fmt.Errorf("drop mysql database: %w", err)
	}

	return nil
}
