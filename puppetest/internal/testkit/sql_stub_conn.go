package testkit

import (
	"context"
	"database/sql/driver"
	"errors"
)

type stubSQLConn struct {
	state *SQLState
}

var (
	_ driver.ExecerContext  = (*stubSQLConn)(nil)
	_ driver.QueryerContext = (*stubSQLConn)(nil)
)

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
	args []driver.NamedValue,
) (driver.Result, error) {
	if c.state != nil {
		c.state.recordExec(query, plainValues(args))
		if c.state.ExecErr != nil {
			return nil, c.state.ExecErr
		}
	}
	return driver.RowsAffected(1), nil
}

func (c *stubSQLConn) QueryContext(
	_ context.Context,
	query string,
	args []driver.NamedValue,
) (driver.Rows, error) {
	if c.state == nil {
		return &stubSQLRows{}, nil
	}

	c.state.recordQuery(query, plainValues(args))
	if c.state.QueryErr != nil {
		return nil, c.state.QueryErr
	}

	return &stubSQLRows{columns: c.state.QueryCols, rows: c.state.QueryRows}, nil
}

func plainValues(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for index, arg := range args {
		values[index] = arg.Value
	}
	return values
}
