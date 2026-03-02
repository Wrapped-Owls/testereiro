package puppetest

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestConnectDB(t *testing.T) {
	cases := []struct {
		name        string
		conf        DBConnectionConfig
		build       func(*testing.T) (*sql.DB, error)
		wantErrText string
		wantConf    DBConnectionConfig
		assertDBSet bool
	}{
		{
			name: "passes config and returns db",
			conf: DBConnectionConfig{DBName: "games", AllowMultiStatements: true},
			build: func(t *testing.T) (*sql.DB, error) {
				return testkit.OpenStubDB(t, "connectdb-success", &testkit.SQLState{}), nil
			},
			wantConf:    DBConnectionConfig{DBName: "games", AllowMultiStatements: true},
			assertDBSet: true,
		},
		{
			name: "propagates connector error",
			conf: DBConnectionConfig{DBName: "games"},
			build: func(*testing.T) (*sql.DB, error) {
				return nil, errors.New("connector failed")
			},
			wantErrText: "connector failed",
			wantConf:    DBConnectionConfig{DBName: "games"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seenConf DBConnectionConfig
			performer := ConnectDB(func(conf DBConnectionConfig) (*sql.DB, error) {
				seenConf = conf
				return tc.build(t)
			})

			gotDB, err := performer(context.Background(), tc.conf)
			if tc.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrText) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrText, err)
				}
				if gotDB != nil {
					t.Fatalf("expected nil db when error happens")
				}
			} else if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if seenConf != tc.wantConf {
				t.Fatalf("unexpected config, expected %+v, got %+v", tc.wantConf, seenConf)
			}
			if tc.assertDBSet && gotDB == nil {
				t.Fatalf("expected non-nil db")
			}
			if gotDB != nil {
				if closeErr := gotDB.Close(); closeErr != nil {
					t.Fatalf("close db: %v", closeErr)
				}
			}
		})
	}
}

func TestConnectDBFromDSN(t *testing.T) {
	testkit.EnsureSQLDriver(t)
	cases := []struct {
		name          string
		driver        string
		dsn           string
		conf          DBConnectionConfig
		wantErrSubstr string
		wantOpenedDSN string
	}{
		{
			name:          "returns error for unknown driver",
			driver:        "driver-does-not-exist",
			dsn:           "missing-dsn",
			conf:          DBConnectionConfig{DBName: "abc"},
			wantErrSubstr: "unknown driver",
		},
		{
			name:          "opens using generated dsn",
			driver:        testkit.StubSQLDriverName,
			dsn:           "known-dsn",
			conf:          DBConnectionConfig{DBName: "db"},
			wantOpenedDSN: "known-dsn",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testkit.ResetSQLRegistry()
			testkit.RegisterSQLState(tc.dsn, &testkit.SQLState{})

			performer := ConnectDBFromDSN(tc.driver, func(conf DBConnectionConfig) string {
				if conf != tc.conf {
					t.Fatalf("unexpected config: %+v", conf)
				}
				return tc.dsn
			})

			db, err := performer(context.Background(), tc.conf)
			if tc.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if db == nil {
				t.Fatalf("expected non-nil db")
			}
			if pingErr := db.PingContext(context.Background()); pingErr != nil {
				t.Fatalf("ping db: %v", pingErr)
			}

			opened := testkit.OpenedDSNs()
			if len(opened) != 1 || opened[0] != tc.wantOpenedDSN {
				t.Fatalf("expected opened DSNs [%q], got %v", tc.wantOpenedDSN, opened)
			}

			if closeErr := db.Close(); closeErr != nil {
				t.Fatalf("close db: %v", closeErr)
			}
		})
	}
}
