package puppetest

// WithBeforeEngineCreate registers a hook executed before each engine is created.
func WithBeforeEngineCreate(hook BeforeEngineCreateHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineCreateHooks = append(
			fac.hookLifecycle.beforeEngineCreateHooks,
			hook,
		)
		return nil
	}
}

// WithAfterEngineCreate registers a hook executed after each engine is created.
func WithAfterEngineCreate(hook AfterEngineCreateHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterEngineCreateHooks = append(
			fac.hookLifecycle.afterEngineCreateHooks,
			hook,
		)
		return nil
	}
}

// WithBeforeEngineSeed registers a hook executed before Engine.Seed and Engine.SeedWithProvider.
func WithBeforeEngineSeed(hook BeforeEngineSeedHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineSeedHooks = append(
			fac.hookLifecycle.beforeEngineSeedHooks,
			hook,
		)
		return nil
	}
}

// WithBeforeEngineRun registers a hook executed before Engine.Execute.
func WithBeforeEngineRun(hook BeforeEngineRunHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineRunHooks = append(
			fac.hookLifecycle.beforeEngineRunHooks,
			hook,
		)
		return nil
	}
}

// WithAfterEngineRun registers a hook executed after Engine.Execute.
func WithAfterEngineRun(hook AfterEngineRunHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterEngineRunHooks = append(
			fac.hookLifecycle.afterEngineRunHooks,
			hook,
		)
		return nil
	}
}

// WithBeforeEngineTeardown registers a hook executed before Engine.Teardown.
func WithBeforeEngineTeardown(hook BeforeEngineTeardownHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineTeardownHooks = append(
			fac.hookLifecycle.beforeEngineTeardownHooks,
			hook,
		)
		return nil
	}
}

// WithAfterEngineTeardown registers a hook executed after Engine.Teardown.
func WithAfterEngineTeardown(hook AfterEngineTeardownHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterEngineTeardownHooks = append(
			fac.hookLifecycle.afterEngineTeardownHooks,
			hook,
		)
		return nil
	}
}

// WithBeforeFactoryClose registers a hook executed before EngineFactory.Close.
func WithBeforeFactoryClose(hook BeforeFactoryCloseHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeFactoryCloseHooks = append(
			fac.hookLifecycle.beforeFactoryCloseHooks,
			hook,
		)
		return nil
	}
}

// WithAfterFactoryClose registers a hook executed after EngineFactory.Close.
func WithAfterFactoryClose(hook AfterFactoryCloseHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterFactoryCloseHooks = append(
			fac.hookLifecycle.afterFactoryCloseHooks,
			hook,
		)
		return nil
	}
}
