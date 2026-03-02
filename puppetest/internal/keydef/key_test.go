package keydef

import (
	"reflect"
	"testing"
)

func TestTypedKey_Tag(t *testing.T) {
	tests := []struct {
		name    string
		key     Key
		wantTag string
	}{
		{
			name:    "new key has empty tag",
			key:     NewKey[int](),
			wantTag: "",
		},
		{
			name:    "new tagged key keeps tag",
			key:     NewTaggedKey[int]("my-tag"),
			wantTag: "my-tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.Tag(); got != tt.wantTag {
				t.Fatalf("expected tag %q, got %q", tt.wantTag, got)
			}
		})
	}
}

func TestTypedKey_Type(t *testing.T) {
	tests := []struct {
		name     string
		keyType  func() reflect.Type
		wantType reflect.Type
	}{
		{
			name: "int type",
			keyType: func() reflect.Type {
				return NewKey[int]().Type()
			},
			wantType: reflect.TypeOf(int(0)),
		},
		{
			name: "pointer type",
			keyType: func() reflect.Type {
				return NewKey[*int]().Type()
			},
			wantType: reflect.TypeOf((*int)(nil)),
		},
		{
			name: "interface type returns nil reflect type",
			keyType: func() reflect.Type {
				return NewKey[any]().Type()
			},
			wantType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.keyType(); got != tt.wantType {
				t.Fatalf("expected type %v, got %v", tt.wantType, got)
			}
		})
	}
}
