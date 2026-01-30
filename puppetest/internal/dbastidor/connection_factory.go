package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ConnectionFactory struct {
	rootDB        *sql.DB
	connPerformer ConnectionPerformer
	connTimeout   time.Duration
}

func NewConnectionFactory(
	ctx context.Context, performer ConnectionPerformer,
) (factory ConnectionFactory, err error) {
	if performer == nil {
		return ConnectionFactory{}, errors.New("nil connection performer")
	}
	factory.connPerformer = performer
	factory.connTimeout = time.Second

	factory.rootDB, err = factory.connPerformer.Execute(
		ctx,
		ConnectionConfig{},
		factory.connTimeout,
	)
	return factory, err
}

func (fac *ConnectionFactory) NewDatabase(ctx context.Context, dbName string) (
	newDb struct {
		Connection *sql.DB
		Name       string
	},
	err error,
) {
	newDb.Name = dbNameNormalizer(dbName)
	if _, err = fac.rootDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", newDb.Name)); err != nil {
		return newDb, err
	}

	dbConf := ConnectionConfig{
		DBName:               newDb.Name,
		AllowMultiStatements: true,
	}
	newDb.Connection, err = fac.connPerformer.Execute(ctx, dbConf, fac.connTimeout)
	return newDb, err
}

func (fac *ConnectionFactory) Close() error {
	if fac.rootDB != nil {
		err := fac.rootDB.Close()
		fac.rootDB = nil
		return err
	}

	return nil
}
