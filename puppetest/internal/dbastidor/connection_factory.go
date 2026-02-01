package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type ConnectionFactory struct {
	rootDB        *sql.DB
	connPerformer ConnectionPerformer
	connTimeout   time.Duration
	lifecycle     DBLifecycle
}

func NewConnectionFactory(
	ctx context.Context, executeCreateDbStmt bool, performer ConnectionPerformer,
) (factory *ConnectionFactory, err error) {
	if performer == nil {
		return &ConnectionFactory{}, errors.New("nil connection performer")
	}
	factory = &ConnectionFactory{
		connPerformer: performer,
		connTimeout:   time.Second,
		lifecycle:     NoOpLifecycle{},
	}

	factory.rootDB, err = factory.connPerformer.Execute(
		ctx, ConnectionConfig{}, factory.connTimeout,
	)
	if err != nil {
		return factory, err
	}

	if executeCreateDbStmt {
		factory.lifecycle = NewSQLLifecycle(factory.rootDB)
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
		Teardown   func() error
	},
	err error,
) {
	newDb.Name = dbNameNormalizer(dbName)
	// The database creation must be done initially to allow connecting directly with it
	if err = fac.lifecycle.Create(ctx, newDb.Name); err != nil {
		return newDb, err
	}

	dbConf := ConnectionConfig{
		DBName:               newDb.Name,
		AllowMultiStatements: true,
	}
	if newDb.Connection, err = fac.connPerformer.Execute(ctx, dbConf, fac.connTimeout); err != nil {
		return newDb, err
	}

	newDb.Teardown = func() error {
		return errors.Join(newDb.Connection.Close(), fac.lifecycle.Drop(ctx, newDb.Name))
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
