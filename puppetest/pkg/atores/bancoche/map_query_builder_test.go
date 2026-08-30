package bancoche

import (
	"strconv"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

func TestMapQueryBuilder_Build(t *testing.T) {
	ctx := stgctx.NewRunnerContext(t.Context())

	tests := []struct {
		name          string
		table         string
		filters       map[string]any
		expectedQuery string
		expectedArgs  []any
	}{
		{
			name:          "basic_query",
			table:         "users",
			filters:       map[string]any{"id": 1},
			expectedQuery: "SELECT * FROM users WHERE id = ?",
			expectedArgs:  []any{1},
		},
		{
			name:          "single_filter_name",
			table:         "users",
			filters:       map[string]any{"name": "john"},
			expectedQuery: "SELECT * FROM users WHERE name = ?",
			expectedArgs:  []any{"john"},
		},
		{
			name:          "no_filters",
			table:         "users",
			filters:       nil,
			expectedQuery: "SELECT * FROM users",
			expectedArgs:  nil,
		},
	}

	for _, tCase := range tests {
		t.Run(tCase.name, func(t *testing.T) {
			qb := NewMapQuery(tCase.table, tCase.filters)

			query, args, err := qb.Build(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if query != tCase.expectedQuery {
				t.Errorf("expected query %q, got %q", tCase.expectedQuery, query)
			}

			if len(args) != len(tCase.expectedArgs) {
				t.Fatalf("expected %d args, got %d (%v)", len(tCase.expectedArgs), len(args), args)
			}

			for index := range args {
				if args[index] != tCase.expectedArgs[index] {
					t.Errorf(
						"expected arg[%d] = %v, got %v",
						index,
						tCase.expectedArgs[index],
						args[index],
					)
				}
			}
		})
	}
}

func TestMapQueryBuilder_PlaceholderStyle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		opts  []MapQueryOption
		want  string
		wantN int
	}{
		{
			name:  "defaults to question marks",
			want:  "SELECT * FROM jokers WHERE rarity = ?",
			wantN: 1,
		},
		{
			name:  "ordinal style numbers the clause",
			opts:  []MapQueryOption{WithPlaceholderStyle(ordinal)},
			want:  "SELECT * FROM jokers WHERE rarity = $1",
			wantN: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			builder := NewMapQuery("jokers", map[string]any{"rarity": "Common"}, testCase.opts...)
			query, args, err := builder.Build(stgctx.NewRunnerContext(t.Context()))
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if query != testCase.want {
				t.Fatalf("expected query %q, got %q", testCase.want, query)
			}
			if len(args) != testCase.wantN {
				t.Fatalf("expected %d args, got %v", testCase.wantN, args)
			}
		})
	}
}

func TestMapQueryBuilder_OrdinalsNumberEachFilter(t *testing.T) {
	t.Parallel()

	// Map iteration order is random, so assert the marker SET rather than a fixed clause order.
	builder := NewMapQuery(
		"jokers",
		map[string]any{"rarity": "Common", "name": "Jimbo", "edition": "foil"},
		WithPlaceholderStyle(ordinal),
	)
	query, args, err := builder.Build(stgctx.NewRunnerContext(t.Context()))
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %v", args)
	}
	for _, marker := range []string{"$1", "$2", "$3"} {
		if !strings.Contains(query, "= "+marker) {
			t.Fatalf("expected %q in query, got %q", marker, query)
		}
	}
}

func ordinal(index int) string {
	return "$" + strconv.Itoa(index)
}
