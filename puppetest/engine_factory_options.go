package puppetest

import (
	"errors"
	"time"
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

// WithPlaceholderStyle sets the bind markers Engine.Seed generates; it defaults to
// QuestionPlaceholder, so MySQL and SQLite consumers need not set it.
func WithPlaceholderStyle(style PlaceholderStyle) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		if style == nil {
			return errors.New("nil placeholder style")
		}
		fac.placeholderStyle = style
		return nil
	}
}

// WithConnectionTimeout bounds the whole connect path. The ConnectionPerformer call and the ping
// that follows it SHARE this budget, so a performer that warms a connection eats into the ping's
// share rather than getting its own. Defaults to one second.
func WithConnectionTimeout(timeout time.Duration) EngineFactoryOption {
	return func(fac *EngineFactory) error {
		if timeout <= 0 {
			return errors.New("connection timeout must be positive")
		}
		fac.connTimeout = timeout
		return nil
	}
}
