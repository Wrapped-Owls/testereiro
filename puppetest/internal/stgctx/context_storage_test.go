package stgctx

import (
	"context"
	"testing"
)

func TestTypedStorage_StoreAndLoad(t *testing.T) {
	tests := []struct {
		name      string
		initial   map[StorageKey]any
		storeKey  StorageKey
		storeVal  any
		loadKey   StorageKey
		wantFound bool
		wantVal   any
	}{
		{
			name:      "stores and loads value",
			initial:   map[StorageKey]any{},
			storeKey:  NewTaggedKey[int]("age"),
			storeVal:  30,
			loadKey:   NewTaggedKey[int]("age"),
			wantFound: true,
			wantVal:   30,
		},
		{
			name:      "load missing key returns not found",
			initial:   map[StorageKey]any{},
			storeKey:  NewTaggedKey[int]("age"),
			storeVal:  30,
			loadKey:   NewTaggedKey[int]("missing"),
			wantFound: false,
		},
		{
			name:      "store overwrites existing value",
			initial:   map[StorageKey]any{NewTaggedKey[int]("age"): 21},
			storeKey:  NewTaggedKey[int]("age"),
			storeVal:  35,
			loadKey:   NewTaggedKey[int]("age"),
			wantFound: true,
			wantVal:   35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &typedStorage{values: tt.initial}
			storage.Store(tt.storeKey, tt.storeVal)
			gotVal, gotFound := storage.Load(tt.loadKey)

			if gotFound != tt.wantFound {
				t.Fatalf("expected found=%t, got %t", tt.wantFound, gotFound)
			}
			if tt.wantFound && gotVal != tt.wantVal {
				t.Fatalf("expected value %#v, got %#v", tt.wantVal, gotVal)
			}
		})
	}
}

func TestSaveAndLoadFromCtx(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(RunnerContext)
		wantFoundInt bool
		wantInt      int
		wantFoundStr bool
		wantStr      string
	}{
		{
			name: "save and load same type",
			setup: func(ctx RunnerContext) {
				SaveOnCtx(ctx, 42)
			},
			wantFoundInt: true,
			wantInt:      42,
			wantFoundStr: false,
		},
		{
			name: "load missing type returns zero and false",
			setup: func(RunnerContext) {
			},
			wantFoundInt: false,
			wantInt:      0,
			wantFoundStr: false,
		},
		{
			name: "load fails when stored value has wrong concrete type",
			setup: func(ctx RunnerContext) {
				ctx.Storage().Store(NewKey[int](), "wrong")
			},
			wantFoundInt: false,
			wantInt:      0,
			wantFoundStr: false,
		},
		{
			name: "save overwrites existing typed value",
			setup: func(ctx RunnerContext) {
				SaveOnCtx(ctx, 1)
				SaveOnCtx(ctx, 7)
				SaveOnCtx(ctx, "token")
			},
			wantFoundInt: true,
			wantInt:      7,
			wantFoundStr: true,
			wantStr:      "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewRunnerContext(context.Background())
			tt.setup(ctx)

			gotInt, gotFoundInt := LoadFromCtx[int](ctx)
			if gotFoundInt != tt.wantFoundInt {
				t.Fatalf("expected int found=%t, got %t", tt.wantFoundInt, gotFoundInt)
			}
			if gotInt != tt.wantInt {
				t.Fatalf("expected int value %d, got %d", tt.wantInt, gotInt)
			}

			gotStr, gotFoundStr := LoadFromCtx[string](ctx)
			if gotFoundStr != tt.wantFoundStr {
				t.Fatalf("expected string found=%t, got %t", tt.wantFoundStr, gotFoundStr)
			}
			if gotStr != tt.wantStr {
				t.Fatalf("expected string value %q, got %q", tt.wantStr, gotStr)
			}
		})
	}
}
