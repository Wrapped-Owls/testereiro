package stgctx

import (
	"reflect"
	"testing"
)

func TestContextKeyCreation(t *testing.T) {
	tests := []struct {
		name     string
		key      StorageKey
		wantTag  string
		wantType reflect.Type
	}{
		{
			name:     "new key keeps type and empty tag",
			key:      NewKey[int](),
			wantTag:  "",
			wantType: reflect.TypeOf(int(0)),
		},
		{
			name:     "new tagged key keeps type and tag",
			key:      NewTaggedKey[string]("token"),
			wantTag:  "token",
			wantType: reflect.TypeOf(""),
		},
		{
			name:     "pointer type key",
			key:      NewKey[*int](),
			wantTag:  "",
			wantType: reflect.TypeOf((*int)(nil)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTag := tt.key.Tag(); gotTag != tt.wantTag {
				t.Fatalf("expected tag %q, got %q", tt.wantTag, gotTag)
			}
			if gotType := tt.key.Type(); gotType != tt.wantType {
				t.Fatalf("expected type %v, got %v", tt.wantType, gotType)
			}
		})
	}
}
