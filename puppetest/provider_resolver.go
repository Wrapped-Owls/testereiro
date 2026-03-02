package puppetest

import "github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"

// ProviderKey identifies values stored in engine or factory provider stores.
type ProviderKey = providerstore.Key

// NewProviderKey creates a provider key bound to type V.
func NewProviderKey[V any](_ ...V) ProviderKey {
	return providerstore.NewKey[V]()
}

// NewTaggedProviderKey creates a provider key bound to type V and an explicit tag.
func NewTaggedProviderKey[V any](tag string, _ ...V) ProviderKey {
	return providerstore.NewTaggedKey[V](tag)
}

// ProviderResolver exposes access to a provider store.
type ProviderResolver interface {
	providerStore() *providerstore.Store
}

// ResolveProvider loads and type-asserts a provider by key.
func ResolveProvider[T any](resolver ProviderResolver, key ProviderKey) (*T, bool) {
	if resolver == nil || key == nil {
		return nil, false
	}

	ps := resolver.providerStore()
	value, found := ps.Load(key)
	if !found || value == nil {
		return nil, false
	}

	casted, ok := value.(*T)
	if !ok || casted == nil {
		return nil, false
	}
	return casted, true
}
