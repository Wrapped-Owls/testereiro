package puppetest

import (
	"context"
	"errors"
	"testing"
)

func TestEngine_ProviderStorage(t *testing.T) {
	type sample struct {
		Name string
	}

	engine := &Engine{}
	key := NewTaggedProviderKey[sample]("sample.provider")
	value := &sample{Name: "resource"}

	if err := SetProvider(engine, key, value, nil); err != nil {
		t.Fatalf("failed to set provider: %v", err)
	}

	got, found := Provider[sample](engine, key)
	if !found {
		t.Fatal("expected provider to be found")
	}
	if got.Name != "resource" {
		t.Fatalf("unexpected provider value: %q", got.Name)
	}
}

func TestEngine_ProviderTeardownOnEngineTeardown(t *testing.T) {
	type sample struct {
		Name string
	}

	engine := &Engine{}
	key := NewTaggedProviderKey[sample]("sample.provider")
	value := &sample{Name: "resource"}

	teardownCalled := false
	if err := SetProvider(
		engine,
		key,
		value,
		func(_ context.Context, _ *sample) error {
			teardownCalled = true
			return nil
		},
	); err != nil {
		t.Fatalf("failed to set provider: %v", err)
	}

	if err := engine.Teardown(); err != nil {
		t.Fatalf("teardown failed: %v", err)
	}
	if !teardownCalled {
		t.Fatal("expected provider teardown to be called")
	}
}

func TestEngine_ProviderTeardownError(t *testing.T) {
	type sample struct {
		Name string
	}

	engine := &Engine{}
	key := NewTaggedProviderKey[sample]("sample.provider")
	value := &sample{Name: "resource"}
	expectedErr := errors.New("provider teardown failed")

	if err := SetProvider(
		engine,
		key,
		value,
		func(_ context.Context, _ *sample) error {
			return expectedErr
		},
	); err != nil {
		t.Fatalf("failed to set provider: %v", err)
	}

	err := engine.Teardown()
	if err == nil {
		t.Fatal("expected teardown error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to include provider teardown failure, got: %v", err)
	}
}
