package testkit

import (
	"database/sql"
	"sync"
	"testing"
)

const StubSQLDriverName = "puppetest_unit_stub"

var (
	stubSQLRegisterOnce sync.Once
	stubSQLRegistry     = struct {
		mu     sync.Mutex
		states map[string]*SQLState
		opened []string
	}{states: map[string]*SQLState{}}
)

func EnsureSQLDriver(t testing.TB) {
	t.Helper()
	stubSQLRegisterOnce.Do(func() {
		sql.Register(StubSQLDriverName, stubSQLDriver{})
	})
}

func ResetSQLRegistry() {
	stubSQLRegistry.mu.Lock()
	defer stubSQLRegistry.mu.Unlock()
	stubSQLRegistry.states = map[string]*SQLState{}
	stubSQLRegistry.opened = nil
}

func RegisterSQLState(dsn string, state *SQLState) {
	stubSQLRegistry.mu.Lock()
	defer stubSQLRegistry.mu.Unlock()
	stubSQLRegistry.states[dsn] = state
}

func OpenedDSNs() []string {
	stubSQLRegistry.mu.Lock()
	defer stubSQLRegistry.mu.Unlock()
	copyOfOpened := make([]string, len(stubSQLRegistry.opened))
	copy(copyOfOpened, stubSQLRegistry.opened)
	return copyOfOpened
}
