package dbastidor

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// In WAL mode these outlive the database file and leak between tests if left behind.
var sqliteSidecarSuffixes = [...]string{"", "-wal", "-shm"}

type (
	// An empty path marks the database as memory-backed, with nothing on disk to create or remove.
	DBFilePathResolver func(dbName string) string

	SQLiteLifecycle struct {
		pathFor DBFilePathResolver
	}
)

var _ DBLifecycle = (*SQLiteLifecycle)(nil)

func NewSQLiteLifecycle(pathFor DBFilePathResolver) *SQLiteLifecycle {
	return &SQLiteLifecycle{pathFor: pathFor}
}

func (l *SQLiteLifecycle) Create(_ context.Context, dbName string) error {
	path := l.resolve(dbName)
	if path == "" {
		return nil
	}

	// SQLite has no CREATE DATABASE; the driver writes the file itself when it first opens one.
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create sqlite database directory: %w", err)
	}

	return nil
}

func (l *SQLiteLifecycle) Drop(_ context.Context, dbName string) error {
	path := l.resolve(dbName)
	if path == "" {
		return nil
	}

	var removeErrs []error
	for _, suffix := range sqliteSidecarSuffixes {
		if err := os.Remove(path + suffix); err != nil && !errors.Is(err, fs.ErrNotExist) {
			removeErrs = append(removeErrs, err)
		}
	}
	if err := errors.Join(removeErrs...); err != nil {
		return fmt.Errorf("remove sqlite database files: %w", err)
	}

	return nil
}

func (l *SQLiteLifecycle) resolve(dbName string) string {
	if l.pathFor == nil {
		return ""
	}
	return l.pathFor(dbName)
}
