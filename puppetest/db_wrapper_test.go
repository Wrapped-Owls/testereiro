package puppetest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestDBWrapper(t *testing.T) {
	cases := []struct {
		name              string
		dbName            string
		conn              *sql.DB
		state             *testkit.SQLState
		wantNormalized    string
		wantConnectionNil bool
		wantIsZero        bool
		wantCloseCount    int
	}{
		{
			name:           "normalizes db name and keeps non-nil connection",
			dbName:         "My Test-DB",
			state:          &testkit.SQLState{},
			wantNormalized: "my_test_db",
			wantIsZero:     false,
			wantCloseCount: 1,
		},
		{
			name:              "handles nil connection",
			dbName:            "DB",
			conn:              nil,
			wantNormalized:    "db",
			wantConnectionNil: true,
			wantIsZero:        true,
			wantCloseCount:    0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testkit.ResetSQLRegistry()

			conn := tc.conn
			if conn == nil && tc.state != nil {
				conn = testkit.OpenStubDB(t, "wrapper-"+tc.name, tc.state)
			}

			wrapper := NewDBWrapper(tc.dbName, conn)

			if wrapper.name != tc.wantNormalized {
				t.Fatalf("expected normalized name %q, got %q", tc.wantNormalized, wrapper.name)
			}

			gotConn := wrapper.Connection()
			if tc.wantConnectionNil && gotConn != nil {
				t.Fatalf("expected nil connection")
			}
			if !tc.wantConnectionNil && gotConn == nil {
				t.Fatalf("expected non-nil connection")
			}

			if wrapper.IsZero() != tc.wantIsZero {
				t.Fatalf("expected IsZero=%t, got %t", tc.wantIsZero, wrapper.IsZero())
			}

			if gotConn != nil {
				if pingErr := gotConn.PingContext(context.Background()); pingErr != nil {
					t.Fatalf("ping connection: %v", pingErr)
				}
			}

			if err := wrapper.Teardown(); err != nil {
				t.Fatalf("expected nil teardown error, got %v", err)
			}

			if tc.state != nil {
				if tc.state.CloseCount() != tc.wantCloseCount {
					t.Fatalf(
						"expected close count %d, got %d",
						tc.wantCloseCount,
						tc.state.CloseCount(),
					)
				}
			}
		})
	}
}
