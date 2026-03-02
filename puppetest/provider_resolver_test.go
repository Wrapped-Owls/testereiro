package puppetest

import (
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/providerstore"
)

type resolverStub struct {
	store *providerstore.Store
}

func (r resolverStub) providerStore() *providerstore.Store {
	return r.store
}

func TestResolveProvider(t *testing.T) {
	type sample struct {
		Name string
	}
	goodKey := NewProviderKey[sample]()
	wrongTypeKey := NewProviderKey[string]()
	typedNilKey := NewProviderKey[sample](sample{})

	goodStore := providerstore.New()
	goodValue := &sample{Name: "resource"}
	if err := providerstore.SaveProvider(goodStore, goodKey, goodValue, nil); err != nil {
		t.Fatalf("save provider: %v", err)
	}

	wrongTypeStore := providerstore.New()
	wrongTypeValue := "text"
	if err := providerstore.SaveProvider(wrongTypeStore, wrongTypeKey, &wrongTypeValue, nil); err != nil {
		t.Fatalf("save wrong type provider: %v", err)
	}

	typedNilStore := providerstore.New()
	var typedNil *sample
	if err := providerstore.SaveProvider(typedNilStore, typedNilKey, typedNil, nil); err != nil {
		t.Fatalf("save typed nil provider: %v", err)
	}

	cases := []struct {
		name      string
		resolver  ProviderResolver
		key       ProviderKey
		wantFound bool
		wantName  string
	}{
		{
			name:      "returns false when resolver is nil",
			resolver:  nil,
			key:       goodKey,
			wantFound: false,
		},
		{
			name:      "returns false when key is nil",
			resolver:  resolverStub{store: goodStore},
			key:       nil,
			wantFound: false,
		},
		{
			name:      "returns false when key is missing",
			resolver:  resolverStub{store: providerstore.New()},
			key:       goodKey,
			wantFound: false,
		},
		{
			name:      "returns false when stored type does not match",
			resolver:  resolverStub{store: wrongTypeStore},
			key:       wrongTypeKey,
			wantFound: false,
		},
		{
			name:      "returns false for typed nil pointer",
			resolver:  resolverStub{store: typedNilStore},
			key:       typedNilKey,
			wantFound: false,
		},
		{
			name:      "returns typed provider",
			resolver:  resolverStub{store: goodStore},
			key:       goodKey,
			wantFound: true,
			wantName:  "resource",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := ResolveProvider[sample](tc.resolver, tc.key)
			if found != tc.wantFound {
				t.Fatalf("expected found=%t, got %t", tc.wantFound, found)
			}
			if !tc.wantFound {
				if got != nil {
					t.Fatalf("expected nil result when not found")
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil provider")
			}
			if got.Name != tc.wantName {
				t.Fatalf("expected provider name %q, got %q", tc.wantName, got.Name)
			}
		})
	}
}
