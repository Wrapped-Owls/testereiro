package puppetest

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type testRunner struct {
	run func(testing.TB, stgctx.RunnerContext) error
}

func (r testRunner) Run(t testing.TB, ctx stgctx.RunnerContext) error {
	return r.run(t, ctx)
}

type testSeedProvider struct {
	seed func(*Engine) error
}

func (p testSeedProvider) ExecuteSeed(engine *Engine) error {
	return p.seed(engine)
}

func TestEngineCreateHooks_Order(t *testing.T) {
	var order []string

	fac, err := NewEngineFactory(
		WithBeforeEngineCreate(func(_ *EngineCreateEvent) error {
			order = append(order, "before-create-1")
			return nil
		}),
		WithBeforeEngineCreate(func(_ *EngineCreateEvent) error {
			order = append(order, "before-create-2")
			return nil
		}),
		WithExtensions(func(_ *Engine) error {
			order = append(order, "extension")
			return nil
		}),
		WithAfterEngineCreate(func(_ *EngineCreateEvent) error {
			order = append(order, "after-create-1")
			return nil
		}),
		WithAfterEngineCreate(func(_ *EngineCreateEvent) error {
			order = append(order, "after-create-2")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	fac.NewEngine(t)

	expected := []string{
		"before-create-1",
		"before-create-2",
		"extension",
		"after-create-2",
		"after-create-1",
	}
	if !reflect.DeepEqual(expected, order) {
		t.Fatalf("unexpected hook order: expected %v, got %v", expected, order)
	}
}

func TestEngineRunHooks_BeforeAndRunnerError(t *testing.T) {
	var order []string
	runnerErr := errors.New("runner failed")

	fac, err := NewEngineFactory(
		WithBeforeEngineRun(func(event *EngineRunEvent) error {
			order = append(order, "before-run")
			if event.Ctx == nil {
				return errors.New("expected run context")
			}
			return nil
		}),
		WithAfterEngineRun(func(_ *EngineRunEvent) error {
			order = append(order, "after-run-1")
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	engine := fac.NewEngine(t)
	err = engine.Execute(
		t,
		testRunner{
			run: func(_ testing.TB, _ stgctx.RunnerContext) error {
				order = append(order, "runner")
				return runnerErr
			},
		},
	)
	if !errors.Is(err, runnerErr) {
		t.Fatalf("expected runner error, got: %v", err)
	}

	expected := []string{
		"before-run",
		"runner",
	}
	if !reflect.DeepEqual(expected, order) {
		t.Fatalf("unexpected hook order: expected %v, got %v", expected, order)
	}
}

func TestEngineTeardownHooks_BeforeAbort(t *testing.T) {
	beforeErr := errors.New("before teardown failed")
	afterCalled := false

	engine := &Engine{
		hooks: engineLifecycleHooks{
			beforeTeardownHooks: []BeforeEngineTeardownHook{
				func(_ *EngineTeardownEvent) error {
					return beforeErr
				},
			},
			afterTeardownHooks: []AfterEngineTeardownHook{
				func(_ *EngineTeardownEvent) error {
					afterCalled = true
					return nil
				},
			},
		},
	}

	err := engine.Teardown()
	if !errors.Is(err, beforeErr) {
		t.Fatalf("expected before error, got %v", err)
	}
	if afterCalled {
		t.Fatal("expected after teardown hooks to be skipped when before hook fails")
	}
}

func TestFactoryCloseHooks_OrderAndErrorJoin(t *testing.T) {
	var order []string
	closeHookErr := errors.New("close hook failed")
	afterHookErr := errors.New("after close hook failed")

	fac, err := NewEngineFactory(
		WithBeforeFactoryClose(func(_ *FactoryCloseEvent) error {
			order = append(order, "before-close")
			return nil
		}),
		WithAfterFactoryClose(func(_ *FactoryCloseEvent) error {
			order = append(order, "after-close-1")
			return nil
		}),
		WithAfterFactoryClose(func(_ *FactoryCloseEvent) error {
			order = append(order, "after-close-2")
			return afterHookErr
		}),
		WithAfterFactoryClose(func(_ *FactoryCloseEvent) error {
			order = append(order, "close-hook")
			return closeHookErr
		}),
	)
	if err != nil {
		t.Fatalf("failed to create factory: %v", err)
	}

	err = fac.Close()
	if !errors.Is(err, closeHookErr) {
		t.Fatalf("expected close hook error, got %v", err)
	}
	if !errors.Is(err, afterHookErr) {
		t.Fatalf("expected after close error, got %v", err)
	}

	expected := []string{
		"before-close",
		"close-hook",
		"after-close-2",
		"after-close-1",
	}
	if !reflect.DeepEqual(expected, order) {
		t.Fatalf("unexpected hook order: expected %v, got %v", expected, order)
	}
}

func TestFactoryCreateLifecycle_BeforeErrorSkipsOperationAndAfter(t *testing.T) {
	beforeErr := errors.New("before create failed")
	initCalled := false
	afterCalled := false
	fac := &EngineFactory{
		extensions: []EngineExtension{
			func(_ *Engine) error {
				initCalled = true
				return nil
			},
		},
	}

	lifecycle := engineFactoryHookLifecycle{
		factory: fac,
		beforeEngineCreateHooks: []BeforeEngineCreateHook{
			func(_ *EngineCreateEvent) error {
				return beforeErr
			},
		},
		afterEngineCreateHooks: []AfterEngineCreateHook{
			func(_ *EngineCreateEvent) error {
				afterCalled = true
				return nil
			},
		},
	}

	err := lifecycle.handleEngineCreation(t, &Engine{})
	if err == nil || !strings.Contains(err.Error(), beforeErr.Error()) {
		t.Fatalf("expected before error, got %v", err)
	}
	if initCalled {
		t.Fatal("expected init operation to be skipped when before hook fails")
	}
	if afterCalled {
		t.Fatal("expected after hooks to be skipped when before hook fails")
	}
}

func TestFactoryCreateLifecycle_AfterHookRunsAndPropagatesError(t *testing.T) {
	afterErr := errors.New("after create failed")

	lifecycle := engineFactoryHookLifecycle{
		factory: &EngineFactory{},
		afterEngineCreateHooks: []AfterEngineCreateHook{
			func(_ *EngineCreateEvent) error {
				return afterErr
			},
		},
	}

	err := lifecycle.handleEngineCreation(t, &Engine{})
	if !errors.Is(err, afterErr) {
		t.Fatalf("expected after hook error, got %v", err)
	}
}

func TestFactoryCloseLifecycle_BeforeErrorSkipsOperationAndAfter(t *testing.T) {
	beforeErr := errors.New("before close failed")
	afterCalled := false

	lifecycle := engineFactoryHookLifecycle{
		factory: &EngineFactory{},
		beforeFactoryCloseHooks: []BeforeFactoryCloseHook{
			func(_ *FactoryCloseEvent) error {
				return beforeErr
			},
		},
		afterFactoryCloseHooks: []AfterFactoryCloseHook{
			func(_ *FactoryCloseEvent) error {
				afterCalled = true
				return nil
			},
		},
	}

	err := lifecycle.closeFactory()
	if !errors.Is(err, beforeErr) {
		t.Fatalf("expected before error, got %v", err)
	}
	if afterCalled {
		t.Fatal("expected after hooks to be skipped when before hook fails")
	}
}

func TestEngineSeedHooks_BeforeHookIsExecuted(t *testing.T) {
	hookErr := errors.New("before seed failed")
	hookCalled := false
	engine := &Engine{
		hooks: engineLifecycleHooks{
			beforeSeedHooks: []BeforeEngineSeedHook{
				func(event *EngineSeedEvent) error {
					hookCalled = true
					if len(event.Seeds) != 1 {
						t.Fatalf("expected 1 seed item, got %d", len(event.Seeds))
					}
					if len(event.ProviderSeeds) != 0 {
						t.Fatalf("expected 0 provider seed items, got %d", len(event.ProviderSeeds))
					}
					return hookErr
				},
			},
		},
	}

	err := engine.Seed(struct{}{})
	if !errors.Is(err, hookErr) {
		t.Fatalf("expected before seed error, got %v", err)
	}
	if !hookCalled {
		t.Fatal("expected before seed hook to run")
	}
}

func TestEngineSeedHooks_FromFactoryOption(t *testing.T) {
	hookCalled := false

	fac, err := NewEngineFactory(
		WithBeforeEngineSeed(func(_ *EngineSeedEvent) error {
			hookCalled = true
			return nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to create engine factory: %v", err)
	}

	engine := fac.NewEngine(t)
	err = engine.Seed(struct{}{})
	if err == nil {
		t.Fatal("expected seed to fail without database")
	}
	if err.Error() != "database not initialized" {
		t.Fatalf("unexpected seed error: %v", err)
	}
	if !hookCalled {
		t.Fatal("expected before seed hook to run")
	}
}

func TestEngineSeedWithProvider_BeforeHookIsExecuted(t *testing.T) {
	hookErr := errors.New("before seed failed")
	hookCalled := false
	engine := &Engine{
		hooks: engineLifecycleHooks{
			beforeSeedHooks: []BeforeEngineSeedHook{
				func(event *EngineSeedEvent) error {
					hookCalled = true
					if len(event.Seeds) != 0 {
						t.Fatalf("expected 0 seed items, got %d", len(event.Seeds))
					}
					if len(event.ProviderSeeds) != 1 {
						t.Fatalf("expected 1 seed provider item, got %d", len(event.ProviderSeeds))
					}
					return hookErr
				},
			},
		},
	}

	err := engine.SeedWithProvider(testSeedProvider{
		seed: func(_ *Engine) error { return nil },
	})
	if !errors.Is(err, hookErr) {
		t.Fatalf("expected before seed error, got %v", err)
	}
	if !hookCalled {
		t.Fatal("expected before seed hook to run")
	}
}

func TestEngineSeedWithProvider_CollectsErrors(t *testing.T) {
	providerErr := errors.New("provider failed")
	engine := &Engine{}

	err := engine.SeedWithProvider(
		nil,
		testSeedProvider{seed: func(_ *Engine) error { return providerErr }},
	)
	if err == nil {
		t.Fatal("expected seed-with-provider to fail")
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
	if !strings.Contains(err.Error(), "seed provider at index 0 is nil") {
		t.Fatalf("expected nil provider error, got %v", err)
	}
}
