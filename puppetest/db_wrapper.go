package puppetest

import (
	"database/sql"
)

type DBWrapper struct {
	name string
	conn *sql.DB
}

func NewDBWrapper(name string, conn *sql.DB) *DBWrapper {
	return &DBWrapper{name: name, conn: conn}
}

func (dw *DBWrapper) Teardown() error {
	closeErr := dw.conn.Close()
	return closeErr
}

func (dw *DBWrapper) Connection() *sql.DB {
	return dw.conn
}

func (dw *DBWrapper) IsZero() bool {
	return dw.conn == nil
}
