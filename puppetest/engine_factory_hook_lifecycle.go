package puppetest

import (
	"errors"
	"fmt"
	"testing"
)

type engineFactoryHookLifecycle struct {
	factory                   *EngineFactory
	beforeEngineCreateHooks   []BeforeEngineCreateHook
	afterEngineCreateHooks    []AfterEngineCreateHook
	beforeEngineSeedHooks     []BeforeEngineSeedHook
	beforeEngineRunHooks      []BeforeEngineRunHook
	afterEngineRunHooks       []AfterEngineRunHook
	beforeEngineTeardownHooks []BeforeEngineTeardownHook
	afterEngineTeardownHooks  []AfterEngineTeardownHook
	beforeFactoryCloseHooks   []BeforeFactoryCloseHook
	afterFactoryCloseHooks    []AfterFactoryCloseHook
}

func (l *engineFactoryHookLifecycle) bind(fac *EngineFactory) {
	l.factory = fac
}

func (l *engineFactoryHookLifecycle) handleEngineCreation(t testing.TB, engine *Engine) error {
	if l.factory == nil {
		return errors.New("factory lifecycle is not bound")
	}

	engine.hooks = engineLifecycleHooks{
		beforeSeedHooks: append([]BeforeEngineSeedHook{}, l.beforeEngineSeedHooks...),
		beforeRunHooks:  append([]BeforeEngineRunHook{}, l.beforeEngineRunHooks...),
		afterRunHooks:   append([]AfterEngineRunHook{}, l.afterEngineRunHooks...),
		beforeTeardownHooks: append(
			[]BeforeEngineTeardownHook{}, l.beforeEngineTeardownHooks...,
		),
		afterTeardownHooks: append(
			[]AfterEngineTeardownHook{}, l.afterEngineTeardownHooks...,
		),
	}
	createEvent := &EngineCreateEvent{
		TB:      t,
		Factory: l.factory,
		Engine:  engine,
	}

	beforeHookErr := runHooks(createEvent, l.beforeEngineCreateHooks)
	if beforeHookErr != nil {
		return fmt.Errorf("before-engine-create hooks failed: %v", beforeHookErr)
	}

	if initErr := l.factory.initEngine(t, engine); initErr != nil {
		return fmt.Errorf("init engine failed: %w", initErr)
	}

	if afterHookErr := runHooks(createEvent, reverseHooks(l.afterEngineCreateHooks)); afterHookErr != nil {
		return fmt.Errorf("after-engine-create-hooks failed: %w", afterHookErr)
	}

	return nil
}

func (l *engineFactoryHookLifecycle) closeFactory() error {
	if l.factory == nil {
		return errors.New("factory lifecycle is not bound")
	}

	closeEvent := &FactoryCloseEvent{Factory: l.factory}

	beforeHookErr := runHooks(closeEvent, l.beforeFactoryCloseHooks)
	if beforeHookErr != nil {
		return beforeHookErr
	}

	closeErr := l.factory.closeOperation()
	afterHookErr := runHooks(closeEvent, reverseHooks(l.afterFactoryCloseHooks))
	return errors.Join(closeErr, afterHookErr)
}
