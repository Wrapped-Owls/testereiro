package testkit

import "testing"

func TestStubSQLDriver_Open(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		preRegister bool
	}{
		{name: "creates state when dsn is unknown", dsn: "unknown", preRegister: false},
		{name: "reuses pre-registered state", dsn: "known", preRegister: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetSQLRegistry()
			var registered *SQLState
			if tt.preRegister {
				registered = &SQLState{}
				RegisterSQLState(tt.dsn, registered)
			}

			conn, err := (stubSQLDriver{}).Open(tt.dsn)
			if err != nil {
				t.Fatalf("open returned error: %v", err)
			}

			asConn, ok := conn.(*stubSQLConn)
			if !ok {
				t.Fatalf("expected *stubSQLConn, got %T", conn)
			}

			stubSQLRegistry.mu.Lock()
			storedState := stubSQLRegistry.states[tt.dsn]
			opened := append([]string(nil), stubSQLRegistry.opened...)
			stubSQLRegistry.mu.Unlock()

			if storedState == nil {
				t.Fatalf("expected state to exist for dsn %q", tt.dsn)
			}
			if tt.preRegister && storedState != registered {
				t.Fatalf("expected pre-registered state to be reused")
			}
			if asConn.state != storedState {
				t.Fatalf("expected connection to hold registry state")
			}
			if len(opened) != 1 || opened[0] != tt.dsn {
				t.Fatalf("expected opened dsns [%q], got %v", tt.dsn, opened)
			}

			_ = conn.Close()
		})
	}
}
