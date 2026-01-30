package stgctx

import "reflect"

type typeKey[V any] struct {
	Tag string
}

func (typeKey[V]) isKey() {
	// Do nothing as it is a stub
}

func (typeKey[V]) Type() reflect.Type {
	var v V
	return reflect.TypeOf(v)
}

func NewKey[V any](_ ...V) StorageKey {
	return typeKey[V]{}
}

func NewTaggedKey[V any](tag string, _ ...V) StorageKey {
	return typeKey[V]{Tag: tag}
}
