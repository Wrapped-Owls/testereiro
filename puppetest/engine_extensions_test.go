package puppetest

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestWithTestServer(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
	}{
		{name: "returns 200", statusCode: http.StatusOK, body: "ok"},
		{name: "returns 201", statusCode: http.StatusCreated, body: "created"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := &Engine{}
			ext := WithTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(tc.body))
			}))

			panicValue, err := testkit.ApplyWithPanicCapture(t, func() error { return ext(engine) })
			if panicValue != nil {
				panicText := fmt.Sprint(panicValue)
				if strings.Contains(panicText, "operation not permitted") {
					t.Skipf("socket creation is blocked in this environment: %v", panicValue)
				}
				t.Fatalf("unexpected panic: %v", panicValue)
			}
			if err != nil {
				t.Fatalf("apply extension: %v", err)
			}
			t.Cleanup(func() { _ = engine.Teardown() })

			if engine.BaseURL() == "" {
				t.Fatalf("expected non-empty base url")
			}

			resp, err := http.Get(engine.BaseURL())
			if err != nil {
				t.Fatalf("http get: %v", err)
			}
			defer func() {
				if closeErr := resp.Body.Close(); closeErr != nil {
					t.Errorf("close response body: %v", closeErr)
				}
			}()

			if resp.StatusCode != tc.statusCode {
				t.Fatalf("expected status %d, got %d", tc.statusCode, resp.StatusCode)
			}
		})
	}
}

func TestWithTestServerFromEngine(t *testing.T) {
	factoryErr := errors.New("factory failed")
	cases := []struct {
		name           string
		handlerFactory func(*Engine) (http.Handler, error)
		wantErr        error
		wantErrSubstr  string
	}{
		{
			name: "propagates wrapped handler error",
			handlerFactory: func(*Engine) (http.Handler, error) {
				return nil, factoryErr
			},
			wantErr:       factoryErr,
			wantErrSubstr: "could not create main handler",
		},
		{
			name: "sets server when factory succeeds",
			handlerFactory: func(e *Engine) (http.Handler, error) {
				if e == nil {
					return nil, fmt.Errorf("engine is nil")
				}
				return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					_, _ = w.Write([]byte("ok"))
				}), nil
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine := &Engine{}
			ext := WithTestServerFromEngine(tc.handlerFactory)
			panicValue, err := testkit.ApplyWithPanicCapture(t, func() error { return ext(engine) })

			if panicValue != nil {
				panicText := fmt.Sprint(panicValue)
				if strings.Contains(panicText, "operation not permitted") {
					t.Skipf("socket creation is blocked in this environment: %v", panicValue)
				}
				t.Fatalf("unexpected panic: %v", panicValue)
			}

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("expected error to contain %q, got %v", tc.wantErrSubstr, err)
				}
				if engine.ts != nil {
					t.Fatalf("expected test server not to be created on error")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			t.Cleanup(func() { _ = engine.Teardown() })

			if engine.ts == nil {
				t.Fatalf("expected test server to be created")
			}
		})
	}
}

func TestWithMigrationRunner(t *testing.T) {
	execErr := errors.New("migration execution failed")
	cases := []struct {
		name          string
		setup         func(*testing.T) (*Engine, *testkit.SQLState)
		migrations    fs.FS
		wantErr       error
		wantErrSubstr string
		wantExecStmts []string
	}{
		{
			name:          "returns error when db wrapper is nil",
			setup:         func(*testing.T) (*Engine, *testkit.SQLState) { return &Engine{}, nil },
			migrations:    fstest.MapFS{},
			wantErrSubstr: "database not initialized",
		},
		{
			name:          "returns error when db connection is nil",
			setup:         func(*testing.T) (*Engine, *testkit.SQLState) { return &Engine{db: NewDBWrapper("x", nil)}, nil },
			migrations:    fstest.MapFS{},
			wantErrSubstr: "database not initialized",
		},
		{
			name: "returns read error from migration fs",
			setup: func(t *testing.T) (*Engine, *testkit.SQLState) {
				state := &testkit.SQLState{}
				return &Engine{
					db: NewDBWrapper("x", testkit.OpenStubDB(t, "migration-read-error", state)),
				}, state
			},
			migrations:    testkit.BrokenFS{},
			wantErrSubstr: "error reading migrations",
		},
		{
			name: "executes only root sql migrations",
			setup: func(t *testing.T) (*Engine, *testkit.SQLState) {
				state := &testkit.SQLState{}
				return &Engine{
					db: NewDBWrapper("x", testkit.OpenStubDB(t, "migration-success", state)),
				}, state
			},
			migrations: fstest.MapFS{
				"001_create.sql": {Data: []byte("CREATE TABLE t(id INT)")},
				"README.txt":     {Data: []byte("ignore")},
				"nested/002.sql": {Data: []byte("CREATE TABLE ignored(id INT)")},
			},
			wantExecStmts: []string{"CREATE TABLE t(id INT)"},
		},
		{
			name: "returns execution error from migration",
			setup: func(t *testing.T) (*Engine, *testkit.SQLState) {
				state := &testkit.SQLState{ExecErr: execErr}
				return &Engine{
					db: NewDBWrapper("x", testkit.OpenStubDB(t, "migration-exec-error", state)),
				}, state
			},
			migrations:    fstest.MapFS{"001.sql": {Data: []byte("CREATE TABLE fail(id INT)")}},
			wantErr:       execErr,
			wantErrSubstr: "failed to execute migration file",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testkit.ResetSQLRegistry()
			engine, state := tc.setup(t)

			ext := WithMigrationRunner(tc.migrations)
			err := ext(engine)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			}
			if tc.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrSubstr, err)
				}
			}
			if tc.wantErr == nil && tc.wantErrSubstr == "" && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(tc.wantExecStmts) > 0 {
				if state == nil {
					t.Fatalf("test case requires non-nil state")
				}

				executed := state.ExecStatements()
				if len(executed) != len(tc.wantExecStmts) {
					t.Fatalf("expected exec statements %v, got %v", tc.wantExecStmts, executed)
				}
				for i := range tc.wantExecStmts {
					if executed[i] != tc.wantExecStmts[i] {
						t.Fatalf("expected exec statements %v, got %v", tc.wantExecStmts, executed)
					}
				}
			}
		})
	}
}
