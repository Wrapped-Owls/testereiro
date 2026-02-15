package puppetest

func WithBeforeEngineCreate(hook BeforeEngineCreateHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineCreateHooks = append(
			fac.hookLifecycle.beforeEngineCreateHooks,
			hook,
		)
		return nil
	}
}

func WithAfterEngineCreate(hook AfterEngineCreateHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterEngineCreateHooks = append(
			fac.hookLifecycle.afterEngineCreateHooks,
			hook,
		)
		return nil
	}
}

func WithBeforeEngineSeed(hook BeforeEngineSeedHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineSeedHooks = append(
			fac.hookLifecycle.beforeEngineSeedHooks,
			hook,
		)
		return nil
	}
}

func WithBeforeEngineRun(hook BeforeEngineRunHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineRunHooks = append(
			fac.hookLifecycle.beforeEngineRunHooks,
			hook,
		)
		return nil
	}
}

func WithAfterEngineRun(hook AfterEngineRunHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterEngineRunHooks = append(
			fac.hookLifecycle.afterEngineRunHooks,
			hook,
		)
		return nil
	}
}

func WithBeforeEngineTeardown(hook BeforeEngineTeardownHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeEngineTeardownHooks = append(
			fac.hookLifecycle.beforeEngineTeardownHooks,
			hook,
		)
		return nil
	}
}

func WithAfterEngineTeardown(hook AfterEngineTeardownHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterEngineTeardownHooks = append(
			fac.hookLifecycle.afterEngineTeardownHooks,
			hook,
		)
		return nil
	}
}

func WithBeforeFactoryClose(hook BeforeFactoryCloseHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.beforeFactoryCloseHooks = append(
			fac.hookLifecycle.beforeFactoryCloseHooks,
			hook,
		)
		return nil
	}
}

func WithAfterFactoryClose(hook AfterFactoryCloseHook) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.hookLifecycle.afterFactoryCloseHooks = append(
			fac.hookLifecycle.afterFactoryCloseHooks,
			hook,
		)
		return nil
	}
}
