package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// DefaultConnectionTimeout bounds the whole connect path when no timeout is configured.
const DefaultConnectionTimeout = time.Second

type ConnectionFactory struct {
	rootDB        *sql.DB
	connPerformer ConnectionPerformer
	connTimeout   time.Duration
	lifecycle     DBLifecycle
}

func NewConnectionFactory(
	ctx context.Context,
	performer ConnectionPerformer,
	buildLifecycle LifecycleBuilder,
	connTimeout time.Duration,
) (factory *ConnectionFactory, err error) {
	if performer == nil {
		return &ConnectionFactory{}, errors.New("nil connection performer")
	}
	if connTimeout <= 0 {
		connTimeout = DefaultConnectionTimeout
	}
	factory = &ConnectionFactory{
		connPerformer: performer,
		connTimeout:   connTimeout,
		lifecycle:     NoOpLifecycle{},
	}

	factory.rootDB, err = factory.connPerformer.Execute(
		ctx, ConnectionConfig{}, factory.connTimeout,
	)
	if err != nil {
		return factory, err
	}

	if buildLifecycle != nil {
		if factory.lifecycle = buildLifecycle(factory.rootDB); factory.lifecycle == nil {
			return factory, errors.Join(
				errors.New("database lifecycle builder returned nil"), factory.Close(),
			)
		}
	}

	return factory, nil
}

func (fac *ConnectionFactory) IsSetup() bool {
	return fac.rootDB != nil && fac.connPerformer != nil
}

func (fac *ConnectionFactory) NewDatabase(ctx context.Context, dbName string) (
	newDb struct {
		Connection *sql.DB
		Name       string
		Teardown   func(ctx context.Context) error
	},
	err error,
) {
	newDb.Name = NormalizeDBName(dbName)
	// The database creation must be done initially to allow connecting directly with it
	if err = fac.lifecycle.Create(ctx, newDb.Name); err != nil {
		return newDb, err
	}

	dbConf := ConnectionConfig{
		DBName:               newDb.Name,
		AllowMultiStatements: true,
	}
	if newDb.Connection, err = fac.connPerformer.Execute(ctx, dbConf, fac.connTimeout); err != nil {
		// The database exists by now, so a failure to connect would strand it on the server.
		return newDb, errors.Join(err, fac.lifecycle.Drop(ctx, newDb.Name))
	}

	newDb.Teardown = func(subCtx context.Context) error {
		return errors.Join(newDb.Connection.Close(), fac.lifecycle.Drop(subCtx, newDb.Name))
	}
	return newDb, nil
}

func (fac *ConnectionFactory) Close() error {
	if fac.rootDB != nil {
		err := fac.rootDB.Close()
		fac.rootDB = nil
		return err
	}

	return nil
}
