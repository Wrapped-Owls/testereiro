package providerstore

import "github.com/wrapped-owls/testereiro/puppetest/internal/keydef"

type Key interface {
	keydef.Key
	isProviderKey()
}

type typeKey[V any] struct {
	keydef.TypedKey[V]
}

func (typeKey[V]) isProviderKey() {}

func NewKey[V any](_ ...V) Key {
	return typeKey[V]{TypedKey: keydef.NewKey[V]()}
}

func NewTaggedKey[V any](tag string, _ ...V) Key {
	return typeKey[V]{TypedKey: keydef.NewTaggedKey[V](tag)}
}
