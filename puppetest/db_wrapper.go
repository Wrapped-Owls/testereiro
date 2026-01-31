package puppetest

import (
	"database/sql"
	"errors"
	"fmt"
)

type DBWrapper struct {
	name string
	conn *sql.DB
}

func NewRootDBWrapper(name string, conn *sql.DB) *DBWrapper {
	return &DBWrapper{name: name, conn: conn}
}

func (dw *DBWrapper) Teardown() error {
	var execErr error
	if dw.name != "" {
		_, execErr = dw.conn.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dw.name))
	}
	closeErr := dw.conn.Close()
	if err := errors.Join(execErr, closeErr); err != nil {
		return err
	}

	return nil
}

func (dw *DBWrapper) Connection() *sql.DB {
	return dw.conn
}

func (dw *DBWrapper) Name() string {
	return dw.name
}

func (dw *DBWrapper) IsZero() bool {
	return dw.conn == nil
}
