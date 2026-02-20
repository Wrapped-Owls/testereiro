package puppetest

import (
	"context"
	"errors"
	"testing"
)

func TestEngineFactory_ProviderStorage(t *testing.T) {
	type sample struct {
		Name string
	}

	factory, err := NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}
	key := NewTaggedProviderKey[sample]("sample.factory.provider")
	value := &sample{Name: "resource"}

	if err = RegisterFactoryProvider(factory, key, value, nil, nil); err != nil {
		t.Fatalf("failed to set factory provider: %v", err)
	}

	got, found := FactoryProvider[sample](factory, key)
	if !found {
		t.Fatal("expected factory provider to be found")
	}
	if got.Name != "resource" {
		t.Fatalf("unexpected factory provider value: %q", got.Name)
	}
}

func TestEngineFactory_ProviderTeardownOnFactoryClose(t *testing.T) {
	type sample struct {
		Name string
	}

	factory, err := NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}
	key := NewTaggedProviderKey[sample]("sample.factory.provider")
	value := &sample{Name: "resource"}

	teardownCalled := false
	if err = RegisterFactoryProvider(
		factory,
		key,
		value,
		nil,
		func(_ context.Context, _ *sample) error {
			teardownCalled = true
			return nil
		},
	); err != nil {
		t.Fatalf("failed to set provider: %v", err)
	}

	if err = factory.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if !teardownCalled {
		t.Fatal("expected factory provider teardown to be called")
	}
}

func TestEngineFactory_ProviderTeardownError(t *testing.T) {
	type sample struct {
		Name string
	}

	factory, err := NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}
	key := NewTaggedProviderKey[sample]("sample.factory.provider")
	value := &sample{Name: "resource"}
	expectedErr := errors.New("factory provider teardown failed")

	if err = RegisterFactoryProvider(
		factory,
		key,
		value,
		nil,
		func(_ context.Context, _ *sample) error {
			return expectedErr
		},
	); err != nil {
		t.Fatalf("failed to set provider: %v", err)
	}

	err = factory.Close()
	if err == nil {
		t.Fatal("expected close error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to include provider teardown failure, got: %v", err)
	}
}

func TestEngineFactory_RegisterFactoryProvider_BindsOnNewEngine(t *testing.T) {
	type factorySample struct {
		Name string
	}
	type engineSample struct {
		Name string
	}

	factoryProviderKey := NewTaggedProviderKey[factorySample]("sample.factory.provider")
	engineProviderKey := NewTaggedProviderKey[engineSample]("sample.engine.provider")
	value := &factorySample{Name: "resource"}
	binderCalled := false

	factory, err := NewEngineFactory(func(fac *EngineFactory) error {
		return RegisterFactoryProvider(
			fac,
			factoryProviderKey,
			value,
			func(
				_ context.Context,
				_ *EngineFactory,
				engine *Engine,
				resource *factorySample,
			) error {
				binderCalled = true
				derived := &engineSample{Name: resource.Name}
				return SetProvider(engine, engineProviderKey, derived, nil)
			},
			nil,
		)
	})
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	engine := factory.NewEngine(t)
	if !binderCalled {
		t.Fatal("expected factory provider binder to be called on engine creation")
	}
	bound, found := Provider[engineSample](engine, engineProviderKey)
	if !found || bound == nil {
		t.Fatal("expected bound engine provider to be available")
	}
	if bound.Name != value.Name {
		t.Fatalf("unexpected bound engine provider value: %q", bound.Name)
	}
}

func TestEngineFactory_FactoryProvidersBindBeforeExtensions(t *testing.T) {
	factoryProviderKey := NewTaggedProviderKey[int]("sample.factory.provider")
	engineProviderKey := NewTaggedProviderKey[int]("sample.engine.provider")
	value := 42
	seenInExtension := false

	factory, err := NewEngineFactory(
		func(fac *EngineFactory) error {
			return RegisterFactoryProvider(
				fac,
				factoryProviderKey,
				&value,
				func(_ context.Context, _ *EngineFactory, engine *Engine, resource *int) error {
					return SetProvider(engine, engineProviderKey, resource, nil)
				},
				nil,
			)
		},
		WithExtensions(func(engine *Engine) error {
			_, seenInExtension = Provider[int](engine, engineProviderKey)
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	_ = factory.NewEngine(t)
	if !seenInExtension {
		t.Fatal("expected factory-bound provider to be available during extension execution")
	}
}

func TestEngineFactory_RegisterFactoryProvider_DuplicateKeyFails(t *testing.T) {
	factory, err := NewEngineFactory()
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}
	key := NewTaggedProviderKey[int]("dup.provider")
	first := 1
	second := 2

	if err = RegisterFactoryProvider(factory, key, &first, nil, nil); err != nil {
		t.Fatalf("failed to register first provider: %v", err)
	}
	if err = RegisterFactoryProvider(factory, key, &second, nil, nil); err == nil {
		t.Fatal("expected duplicate provider key registration to fail")
	}
}

func TestEngineFactory_BindFactoryProviders_ReturnsError(t *testing.T) {
	factoryProviderKey := NewTaggedProviderKey[int]("sample.factory.provider")
	value := 42
	expectedErr := errors.New("bind failed")

	factory, err := NewEngineFactory(func(fac *EngineFactory) error {
		return RegisterFactoryProvider(
			fac,
			factoryProviderKey,
			&value,
			func(_ context.Context, _ *EngineFactory, _ *Engine, _ *int) error {
				return expectedErr
			},
			nil,
		)
	})
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	err = factory.bindFactoryProviders(context.Background(), new(Engine))
	if err == nil {
		t.Fatal("expected bind error")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected bind error to include original error, got: %v", err)
	}
}
