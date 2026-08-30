package puppetest

import (
	"errors"
)

// WithConnectionFactory on its own creates no database and reuses the connected one; pair it with
// WithDatabaseLifecycle for an isolated database per test.
func WithConnectionFactory(connPerformer ConnectionPerformer) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		if connPerformer == nil {
			return errors.New("nil connection performer")
		}
		fac.connPerformer = connPerformer
		return nil
	}
}

// WithDatabaseLifecycle is order-independent: the lifecycle is built once every option has run,
// over the root connection WithConnectionFactory opened.
func WithDatabaseLifecycle(buildLifecycle DBLifecycleBuilder) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		if buildLifecycle == nil {
			return errors.New("nil database lifecycle builder")
		}
		fac.lifecycleBuilder = buildLifecycle
		return nil
	}
}

func WithExtensions(extensions ...EngineExtension) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		fac.extensions = append(fac.extensions, extensions...)
		return nil
	}
}
