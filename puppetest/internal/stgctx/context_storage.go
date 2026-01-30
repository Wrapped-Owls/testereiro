package stgctx

type typedStorage struct {
	values map[StorageKey]any
}

func (s *typedStorage) Store(t StorageKey, val any) {
	s.values[t] = val
}

func (s *typedStorage) Load(t StorageKey) (any, bool) {
	val, ok := s.values[t]
	return val, ok
}

// SaveOnCtx helper for type-safe storage.
func SaveOnCtx[V any](ctx RunnerContext, val V) {
	ctx.Storage().Store(NewKey[V](), val)
}

// LoadFromCtx helper for type-safe retrieval.
func LoadFromCtx[V any](ctx RunnerContext) (result V, ok bool) {
	val, found := ctx.Storage().Load(NewKey[V]())
	if !found {
		return result, false
	}
	result, ok = val.(V)
	return result, ok
}
