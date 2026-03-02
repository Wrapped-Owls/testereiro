package testkit

import (
	"database/sql/driver"
	"errors"
)

type stubSQLStmt struct{}

func (stubSQLStmt) Close() error {
	return nil
}

func (stubSQLStmt) NumInput() int {
	return -1
}

func (stubSQLStmt) Exec([]driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (stubSQLStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, errors.New("queries not supported")
}
