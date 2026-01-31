package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ConnectionFactory struct {
	rootDB          *sql.DB
	connPerformer   ConnectionPerformer
	connTimeout     time.Duration
	executeCreateDb bool
}

func NewConnectionFactory(
	ctx context.Context, executeCreateDbStmt bool, performer ConnectionPerformer,
) (factory ConnectionFactory, err error) {
	if performer == nil {
		return ConnectionFactory{}, errors.New("nil connection performer")
	}
	factory.connPerformer = performer
	factory.connTimeout = time.Second
	factory.executeCreateDb = executeCreateDbStmt

	factory.rootDB, err = factory.connPerformer.Execute(
		ctx, ConnectionConfig{}, factory.connTimeout,
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
	dbConf := ConnectionConfig{
		DBName:               dbNameNormalizer(dbName),
		AllowMultiStatements: true,
	}
	if fac.executeCreateDb {
		// Some drivers, such as SQLite, don't support the CREATE DATABASE command, so we need to skip this line
		if _, err = fac.rootDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", newDb.Name)); err != nil {
			return newDb, err
		}
		newDb.Name = dbConf.DBName
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
