package puppetest

import (
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/testkit"
)

func TestEngineSeed_FailurePaths(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		withDB      bool
		seeds       []any
		wantErrText string
	}{
		{
			name:        "no database configured",
			seeds:       []any{seedRow{Label: "jimbo"}},
			wantErrText: "database not initialized",
		},
		{
			name:        "wraps a failing seed struct",
			withDB:      true,
			seeds:       []any{struct{ Untagged string }{}},
			wantErrText: "failed to seed data",
		},
		{
			name:   "no seeds is not an error",
			withDB: true,
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

			seedErr := fac.NewEngine(t).Seed(testCase.seeds...)
			if testCase.wantErrText == "" {
				if seedErr != nil {
					t.Fatalf("expected no error, got %v", seedErr)
				}
				return
			}
			if seedErr == nil || !strings.Contains(seedErr.Error(), testCase.wantErrText) {
				t.Fatalf("expected error containing %q, got %v", testCase.wantErrText, seedErr)
			}
		})
	}
}

type seedRow struct {
	Label string `db:"label"`
}
