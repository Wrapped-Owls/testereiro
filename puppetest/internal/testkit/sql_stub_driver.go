package testkit

import "database/sql/driver"

type stubSQLDriver struct{}

func (stubSQLDriver) Open(name string) (driver.Conn, error) {
	stubSQLRegistry.mu.Lock()
	defer stubSQLRegistry.mu.Unlock()
	stubSQLRegistry.opened = append(stubSQLRegistry.opened, name)

	state := stubSQLRegistry.states[name]
	if state == nil {
		state = &SQLState{}
		stubSQLRegistry.states[name] = state
	}

	return &stubSQLConn{state: state}, nil
}
