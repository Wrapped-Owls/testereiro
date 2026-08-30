package dbastidor

import (
	"context"
	"database/sql"
)

type (
	DBLifecycle interface {
		Create(ctx context.Context, dbName string) error
		Drop(ctx context.Context, dbName string) error
	}
	LifecycleBuilder func(rootDB *sql.DB) DBLifecycle
)

type NoOpLifecycle struct{}

var _ DBLifecycle = NoOpLifecycle{}

func (NoOpLifecycle) Create(context.Context, string) error { return nil }
func (NoOpLifecycle) Drop(context.Context, string) error   { return nil }
