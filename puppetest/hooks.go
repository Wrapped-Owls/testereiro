package puppetest

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/pkg/atores"
)

type EngineCreateEvent struct {
	TB      testing.TB
	Factory *EngineFactory
	Engine  *Engine
}

type EngineRunEvent struct {
	TB     testing.TB
	Engine *Engine
	Runner atores.Runner
	Ctx    Context
}

type EngineTeardownEvent struct {
	Engine *Engine
}

type EngineSeedEvent struct {
	Engine        *Engine
	Seeds         []any
	ProviderSeeds []SeedProvider
}

type FactoryCloseEvent struct {
	Factory *EngineFactory
}

type (
	BeforeEngineCreateHook func(*EngineCreateEvent) error
	AfterEngineCreateHook  func(*EngineCreateEvent) error
)

type (
	BeforeEngineRunHook func(*EngineRunEvent) error
	AfterEngineRunHook  func(*EngineRunEvent) error
)

type (
	BeforeEngineTeardownHook func(*EngineTeardownEvent) error
	AfterEngineTeardownHook  func(*EngineTeardownEvent) error
)

type BeforeEngineSeedHook func(*EngineSeedEvent) error

type (
	BeforeFactoryCloseHook func(*FactoryCloseEvent) error
	AfterFactoryCloseHook  func(*FactoryCloseEvent) error
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
	var index uint16
	for _, value := range slices.Backward(hooks) {
		reversed[index] = value
		index++
	}
	return reversed
}
