package dbastidor

import (
	"context"
	"database/sql"
	"fmt"
)

type DBLifecycle interface {
	Create(ctx context.Context, dbName string) error
	Drop(ctx context.Context, dbName string) error
}

type SQLLifecycle struct {
	rootDB *sql.DB
}

func NewSQLLifecycle(rootDB *sql.DB) *SQLLifecycle {
	return &SQLLifecycle{
		rootDB: rootDB,
	}
}

func (l *SQLLifecycle) Create(ctx context.Context, dbName string) error {
	_, err := l.rootDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	return err
}

func (l *SQLLifecycle) Drop(ctx context.Context, dbName string) error {
	_, err := l.rootDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	return err
}

type NoOpLifecycle struct{}

func (n NoOpLifecycle) Create(context.Context, string) error { return nil }
func (n NoOpLifecycle) Drop(context.Context, string) error   { return nil }
