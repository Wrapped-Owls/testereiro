package puppetest

import (
	"errors"
	"fmt"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/atores"
)

// EngineCreateEvent contains data available to engine creation hooks.
type EngineCreateEvent struct {
	TB      testing.TB
	Factory *EngineFactory
	Engine  *Engine
}

// EngineRunEvent contains data available to engine run hooks.
type EngineRunEvent struct {
	TB     testing.TB
	Engine *Engine
	Runner atores.Runner
	Ctx    Context
}

// EngineTeardownEvent contains data available to engine teardown hooks.
type EngineTeardownEvent struct {
	Engine *Engine
}

// EngineSeedEvent contains data available to engine seed hooks.
type EngineSeedEvent struct {
	Engine        *Engine
	Seeds         []any
	ProviderSeeds []SeedProvider
}

// FactoryCloseEvent contains data available to factory close hooks.
type FactoryCloseEvent struct {
	Factory *EngineFactory
}

type (
	// BeforeEngineCreateHook runs before engine initialization.
	BeforeEngineCreateHook func(*EngineCreateEvent) error
	// AfterEngineCreateHook runs after engine initialization.
	AfterEngineCreateHook func(*EngineCreateEvent) error
)

type (
	// BeforeEngineRunHook runs before a runner is executed.
	BeforeEngineRunHook func(*EngineRunEvent) error
	// AfterEngineRunHook runs after a runner is executed.
	AfterEngineRunHook func(*EngineRunEvent) error
)

type (
	// BeforeEngineTeardownHook runs before engine teardown starts.
	BeforeEngineTeardownHook func(*EngineTeardownEvent) error
	// AfterEngineTeardownHook runs after engine teardown completes.
	AfterEngineTeardownHook func(*EngineTeardownEvent) error
)

// BeforeEngineSeedHook runs before SQL or provider seed operations.
type BeforeEngineSeedHook func(*EngineSeedEvent) error

type (
	// BeforeFactoryCloseHook runs before factory close operations.
	BeforeFactoryCloseHook func(*FactoryCloseEvent) error
	// AfterFactoryCloseHook runs after factory close operations.
	AfterFactoryCloseHook func(*FactoryCloseEvent) error
)

type engineLifecycleHooks struct {
	beforeSeedHooks     []BeforeEngineSeedHook
	beforeRunHooks      []BeforeEngineRunHook
	afterRunHooks       []AfterEngineRunHook
	beforeTeardownHooks []BeforeEngineTeardownHook
	afterTeardownHooks  []AfterEngineTeardownHook
}

func runHooks[T any, F ~func(T) error](event T, hooks []F) error {
	var hookErrs []error
	for index, hook := range hooks {
		if hook == nil {
			hookErrs = append(hookErrs, fmt.Errorf("hook at index %d is nil", index))
			continue
		}
		if err := hook(event); err != nil {
			hookErrs = append(hookErrs, err)
		}
	}
	return errors.Join(hookErrs...)
}

func reverseHooks[F any](hooks []F) []F {
	reversed := make([]F, len(hooks))
	totalHooks := len(hooks) - 1
	for index := totalHooks; index >= 0; index-- {
		value := hooks[index]
		reversed[totalHooks-index] = value
	}
	return reversed
}
