package keydef

import "reflect"

type Key interface {
	Type() reflect.Type
	Tag() string
}

type TypedKey[V any] struct {
	tag string
}

func (k TypedKey[V]) Tag() string {
	return k.tag
}

func (TypedKey[V]) Type() reflect.Type {
	var v V
	return reflect.TypeOf(v)
}

func NewKey[V any]() TypedKey[V] {
	return TypedKey[V]{}
}

func NewTaggedKey[V any](tag string) TypedKey[V] {
	return TypedKey[V]{tag: tag}
}
