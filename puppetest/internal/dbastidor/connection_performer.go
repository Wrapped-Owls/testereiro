package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type (
	ConnectionConfig struct {
		DBName               string
		AllowMultiStatements bool
	}
	ConnectionPerformer func(ctx context.Context, conf ConnectionConfig) (*sql.DB, error)
)

func (connPerf ConnectionPerformer) Execute(
	ctx context.Context, conf ConnectionConfig, timeout time.Duration,
) (*sql.DB, error) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := connPerf(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("error while performing sql.DB connection: %w", err)
	}

	if err = conn.PingContext(ctx); err != nil {
		return nil, errors.Join(
			fmt.Errorf("failed to ping the connection: %w", err), conn.Close(),
		)
	}

	return conn, nil
}
