package mongochecker

import (
	"errors"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/wrapped-owls/testereiro/providers/mongotest/internal/mongoqueries"
	"github.com/wrapped-owls/testereiro/puppetest"
)

type mongoDoc struct {
	ID   int    `bson:"id"`
	Name string `bson:"name"`
}

func TestCustomValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validate  func(testing.TB, puppetest.Context, *Cursor) error
		wantError bool
	}{
		{name: "nil validation fn", validate: nil, wantError: true},
		{
			name:     "success",
			validate: func(testing.TB, puppetest.Context, *Cursor) error { return nil },
		},
	}

	ctx := testRunnerContext(t)
	cursor := mongoqueries.NewCursorCount(1)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := (&customValidator{validate: tc.validate}).Validate(t, ctx, cursor)
			if tc.wantError && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCountValidator_Validate(t *testing.T) {
	ctx := testRunnerContext(t)

	okValidator := &countValidator{expected: 2}
	if err := okValidator.Validate(t, ctx, mongoqueries.NewCursorCount(2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stored, found := puppetest.LoadFromCtx[int64](ctx)
	if !found || stored != 2 {
		t.Fatalf("unexpected stored count: %d found=%v", stored, found)
	}

	err := (&countValidator{expected: 1}).Validate(t, ctx, mongoqueries.NewCursorCount(3))
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestGenericValidator_ValidateAndExtract(t *testing.T) {
	ctx := testRunnerContext(t)
	docs := mustRawDocs(t,
		mongoDoc{ID: 1, Name: "A"},
		mongoDoc{ID: 2, Name: "B"},
	)

	compareCalls := 0
	validator := genericValidator[mongoDoc, []mongoDoc]{
		ExpectedList: []mongoDoc{{ID: 1, Name: "A"}, {ID: 2, Name: "B"}},
		Comparator: func(testing.TB, mongoDoc, mongoDoc) bool {
			compareCalls++
			return true
		},
	}

	err := validator.Validate(t, ctx, mongoqueries.NewCursorResult(docs...))
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if compareCalls != 2 {
		t.Fatalf("expected comparator to be called for each document, got %d", compareCalls)
	}

	storedList, ok := puppetest.LoadFromCtx[[]mongoDoc](ctx)
	if !ok || len(storedList) != 2 {
		t.Fatalf("expected stored list in context, got len=%d ok=%v", len(storedList), ok)
	}
}

func TestGenericValidator_ValidationErrors(t *testing.T) {
	ctx := testRunnerContext(t)
	docs := mustRawDocs(t, mongoDoc{ID: 1, Name: "A"})

	tests := []struct {
		name      string
		validator genericValidator[mongoDoc, []mongoDoc]
		cursor    *Cursor
		wantError string
	}{
		{
			name: "sanitizer error",
			validator: genericValidator[mongoDoc, []mongoDoc]{
				Expected: mongoDoc{ID: 1},
				Sanitizer: func(expected, actual *mongoDoc) error {
					return errors.New("sanitize")
				},
			},
			cursor:    mongoqueries.NewCursorResult(docs...),
			wantError: "failed to sanitize result",
		},
		{
			name: "comparator false",
			validator: genericValidator[mongoDoc, []mongoDoc]{
				Expected: mongoDoc{ID: 1},
				Comparator: func(testing.TB, mongoDoc, mongoDoc) bool {
					return false
				},
			},
			cursor:    mongoqueries.NewCursorResult(docs...),
			wantError: "mongo result mismatch",
		},
		{
			name: "extract empty cursor",
			validator: genericValidator[mongoDoc, []mongoDoc]{
				Expected: mongoDoc{ID: 1},
			},
			cursor:    mongoqueries.NewCursorResult(),
			wantError: "no documents found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.validator.Validate(t, ctx, tc.cursor)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tc.wantError != "" && !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func mustRawDocs(t *testing.T, docs ...mongoDoc) []bson.Raw {
	t.Helper()
	out := make([]bson.Raw, 0, len(docs))
	for _, doc := range docs {
		raw, err := bson.Marshal(doc)
		if err != nil {
			t.Fatalf("failed to marshal bson: %v", err)
		}
		out = append(out, raw)
	}
	return out
}

type captureContextRunner struct {
	ctx puppetest.Context
}

func (r *captureContextRunner) Run(t testing.TB, ctx puppetest.Context) error {
	r.ctx = ctx
	return nil
}

func testRunnerContext(t *testing.T) puppetest.Context {
	t.Helper()
	runner := &captureContextRunner{}
	engine := &puppetest.Engine{}
	if err := engine.Execute(t, runner); err != nil {
		t.Fatalf("failed to capture context: %v", err)
	}
	if runner.ctx == nil {
		t.Fatal("expected non-nil context")
	}
	return runner.ctx
}
