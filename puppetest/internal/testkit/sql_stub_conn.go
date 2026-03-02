package testkit

import (
	"context"
	"database/sql/driver"
	"errors"
)

type stubSQLConn struct {
	state *SQLState
}

func (c *stubSQLConn) Prepare(string) (driver.Stmt, error) {
	return stubSQLStmt{}, nil
}

func (c *stubSQLConn) Close() error {
	if c.state != nil {
		c.state.recordClose()
	}
	return nil
}

func (c *stubSQLConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

func (c *stubSQLConn) Ping(context.Context) error {
	if c.state == nil {
		return nil
	}
	return c.state.PingErr
}

func (c *stubSQLConn) ExecContext(
	_ context.Context,
	query string,
	_ []driver.NamedValue,
) (driver.Result, error) {
	if c.state != nil {
		c.state.recordExec(query)
		if c.state.ExecErr != nil {
			return nil, c.state.ExecErr
		}
	}
	return driver.RowsAffected(1), nil
}
