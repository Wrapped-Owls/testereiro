package dbastidor

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestConnectionPerformer_Execute(t *testing.T) {
	t.Parallel()

	openErr := errors.New("open failed")
	pingErr := errors.New("ping failed")
	testCases := []struct {
		name       string
		state      *testkit.SQLState
		openErr    error
		wantErr    error
		wantClosed int
		wantNonNil bool
	}{
		{
			name:       "returns the connection when the ping succeeds",
			state:      &testkit.SQLState{},
			wantNonNil: true,
		},
		{
			name:    "wraps a failing open",
			openErr: openErr,
			wantErr: openErr,
		},
		{
			name:       "closes the connection when the ping fails",
			state:      &testkit.SQLState{PingErr: pingErr},
			wantErr:    pingErr,
			wantClosed: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			performer := ConnectionPerformer(
				func(context.Context, ConnectionConfig) (*sql.DB, error) {
					if testCase.openErr != nil {
						return nil, testCase.openErr
					}
					return testkit.OpenStubDB(t, t.Name(), testCase.state), nil
				},
			)

			conn, err := performer.Execute(t.Context(), ConnectionConfig{}, time.Second)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected error %v, got %v", testCase.wantErr, err)
			}
			if testCase.wantNonNil {
				if conn == nil {
					t.Fatal("expected a connection")
				}
				t.Cleanup(func() { _ = conn.Close() })
				return
			}
			if conn != nil {
				t.Fatal("expected no connection on failure")
			}
			if testCase.state == nil {
				return
			}
			if got := testCase.state.CloseCount(); got != testCase.wantClosed {
				t.Fatalf("expected %d close, got %d", testCase.wantClosed, got)
			}
		})
	}
}
