package mongoseeder

import (
	"testing"
)

func TestNew_Defaults(t *testing.T) {
	runner := New()
	if runner == nil {
		t.Fatal("expected seed runner")
	}
	if !runner.clearBefore {
		t.Fatal("expected clear-before default to be true")
	}
	if !runner.ordered {
		t.Fatal("expected ordered default to be true")
	}
	if runner.mode != SeedModeInsertMany {
		t.Fatalf("expected default mode %d, got %d", SeedModeInsertMany, runner.mode)
	}
}

func TestWithClearAndSeed_EnablesClearAndAppendsPlan(t *testing.T) {
	runner := New().
		WithClearBeforeSeed(false).
		WithClearAndSeed("users", map[string]any{"name": "a"})

	if !runner.clearBefore {
		t.Fatal("expected clear-before to be enabled by WithClearAndSeed")
	}
	if len(runner.plans) != 1 {
		t.Fatalf("expected one plan, got %d", len(runner.plans))
	}
	if runner.plans[0].Collection != "users" {
		t.Fatalf("unexpected collection name: %s", runner.plans[0].Collection)
	}
}

func TestSeedWithProvider_ValidatesInput(t *testing.T) {
	runner := New()

	if err := runner.ExecuteSeed(nil); err == nil {
		t.Fatal("expected nil engine error")
	}
}

func TestWithSeedOperationMode_SetsMode(t *testing.T) {
	runner := New().WithSeedOperationMode(SeedModeClientBulkWrite)
	if runner.mode != SeedModeClientBulkWrite {
		t.Fatalf("expected mode %d, got %d", SeedModeClientBulkWrite, runner.mode)
	}
}

func TestWithSeedDocuments_AppendsPlan(t *testing.T) {
	runner := New().WithSeedDocuments("jokers", map[string]any{"name": "cavendish"})
	if len(runner.plans) != 1 {
		t.Fatalf("expected one seed plan, got %d", len(runner.plans))
	}
	if runner.plans[0].Collection != "jokers" {
		t.Fatalf("unexpected collection: %s", runner.plans[0].Collection)
	}
	if len(runner.plans[0].Documents) != 1 {
		t.Fatalf("expected one document, got %d", len(runner.plans[0].Documents))
	}
}
