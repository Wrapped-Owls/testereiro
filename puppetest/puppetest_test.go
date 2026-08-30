package puppetest

import (
	"fmt"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/dbastidor"
	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestEngineAccessors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		withDB    bool
		nameSeed  string
		wantDBSet bool
	}{
		{
			name:     "no connection factory",
			nameSeed: "%s_puppetest",
		},
		{
			name:      "with connection factory",
			withDB:    true,
			nameSeed:  "%s",
			wantDBSet: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var options []EngineFactoryOption
			if testCase.withDB {
				options = append(options, WithConnectionFactory(
					stubConnectionPerformer(t, &testkit.SQLState{}, &testkit.SQLState{}),
				))
			}

			fac, err := NewEngineFactory(options...)
			if err != nil {
				t.Fatalf("create factory: %v", err)
			}
			t.Cleanup(func() { _ = fac.Close() })

			engine := fac.NewEngine(t)
			if engine.Context() == nil {
				t.Fatal("expected a non-nil engine context")
			}
			wantDBName := dbastidor.NormalizeDBName(fmt.Sprintf(testCase.nameSeed, t.Name()))
			if got := engine.DBName(); got != wantDBName {
				t.Fatalf("expected database name %q, got %q", wantDBName, got)
			}
			if hasDB := engine.DB() != nil; hasDB != testCase.wantDBSet {
				t.Fatalf("expected DB set %v, got %v", testCase.wantDBSet, hasDB)
			}
			if engine.BaseURL() != "" {
				t.Fatalf("expected no base URL without a server, got %q", engine.BaseURL())
			}
		})
	}
}

func TestRunnerContextValueRoundTrip(t *testing.T) {
	t.Parallel()

	type payload struct {
		Label string
	}

	ctx := stgctx.NewRunnerContext(t.Context())

	if _, found := LoadFromCtx[payload](ctx); found {
		t.Fatal("expected no value before it is saved")
	}

	SaveOnCtx(ctx, payload{Label: "jimbo"})

	loaded, found := LoadFromCtx[payload](ctx)
	if !found {
		t.Fatal("expected the saved value to be found")
	}
	if loaded.Label != "jimbo" {
		t.Fatalf("expected label %q, got %q", "jimbo", loaded.Label)
	}

	if _, found = LoadFromCtx[string](ctx); found {
		t.Fatal("expected a different type to miss")
	}
}
