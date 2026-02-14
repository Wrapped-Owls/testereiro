package stgctx

import "github.com/wrapped-owls/testereiro/puppetest/internal/keydef"

type typeKey[V any] struct {
	keydef.TypedKey[V]
}

func (typeKey[V]) isStorageKey() {
	// Do nothing as it is a stub
}

func NewKey[V any](_ ...V) StorageKey {
	return typeKey[V]{TypedKey: keydef.NewKey[V]()}
}

func NewTaggedKey[V any](tag string, _ ...V) StorageKey {
	return typeKey[V]{TypedKey: keydef.NewTaggedKey[V](tag)}
}
