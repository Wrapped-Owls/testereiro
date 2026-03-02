package testkit

import (
	"context"
	"testing"
)

func TestSQLRegistry_BasicOperations(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{name: "tracks opened dsn after ping", dsn: "registry-opened"},
		{name: "works with different dsn", dsn: "registry-opened-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetSQLRegistry()
			EnsureSQLDriver(t)
			state := &SQLState{}
			RegisterSQLState(tt.dsn, state)

			db := OpenStubDB(t, tt.dsn, state)
			if err := db.PingContext(context.Background()); err != nil {
				t.Fatalf("ping error: %v", err)
			}
			if err := db.Close(); err != nil {
				t.Fatalf("close error: %v", err)
			}

			opened := OpenedDSNs()
			if len(opened) != 1 || opened[0] != tt.dsn {
				t.Fatalf("expected opened dsns [%q], got %v", tt.dsn, opened)
			}
		})
	}
}

func TestSQLRegistry_OpenedDSNsReturnsCopy(t *testing.T) {
	tests := []struct{ name string }{{name: "mutating returned slice does not mutate registry"}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetSQLRegistry()
			stubSQLRegistry.mu.Lock()
			stubSQLRegistry.opened = []string{"one"}
			stubSQLRegistry.mu.Unlock()

			opened := OpenedDSNs()
			opened[0] = "mutated"

			reloaded := OpenedDSNs()
			if reloaded[0] != "one" {
				t.Fatalf("expected registry opened value to remain %q, got %q", "one", reloaded[0])
			}
		})
	}
}

func TestEnsureSQLDriver_Idempotent(t *testing.T) {
	tests := []struct {
		name  string
		calls int
	}{
		{name: "single call", calls: 1},
		{name: "multiple calls", calls: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.calls; i++ {
				EnsureSQLDriver(t)
			}
		})
	}
}

func TestResetSQLRegistry_ClearsStateAndOpened(t *testing.T) {
	tests := []struct{ name string }{{name: "clears maps and slices"}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetSQLRegistry()
			RegisterSQLState("x", &SQLState{})
			stubSQLRegistry.mu.Lock()
			stubSQLRegistry.opened = append(stubSQLRegistry.opened, "x")
			stubSQLRegistry.mu.Unlock()

			ResetSQLRegistry()

			stubSQLRegistry.mu.Lock()
			stateCount := len(stubSQLRegistry.states)
			openedCount := len(stubSQLRegistry.opened)
			stubSQLRegistry.mu.Unlock()

			if stateCount != 0 || openedCount != 0 {
				t.Fatalf(
					"expected registry to be cleared, states=%d opened=%d",
					stateCount,
					openedCount,
				)
			}
		})
	}
}
