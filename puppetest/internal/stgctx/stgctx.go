package stgctx

import "context"

type StorageKey interface {
	isKey()
}

// Storage handles type-safe storage of values.
type Storage interface {
	Store(StorageKey, any)
	Load(StorageKey) (any, bool)
}

// RunnerContext provides testing capabilities and access to shared storage.
type RunnerContext interface {
	context.Context
	Storage() Storage
}

// runnerCtx implements RunnerContext.
type runnerCtx struct {
	context.Context
	storage *typedStorage
}

func NewRunnerContext(ctx context.Context) RunnerContext {
	return &runnerCtx{
		Context: ctx,
		storage: &typedStorage{
			values: make(map[StorageKey]any),
		},
	}
}

func (c *runnerCtx) Storage() Storage {
	return c.storage
}

func (c *runnerCtx) Value(key any) any {
	if asStorageKey, ok := key.(StorageKey); ok {
		if checkVal, found := c.storage.Load(asStorageKey); found {
			return checkVal
		}
		return nil
	}

	return c.Context.Value(key)
}
