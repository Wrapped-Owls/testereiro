package providerstore

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestStore_Save(t *testing.T) {
	baseKey := NewTaggedKey[int]("base")

	tests := []struct {
		name          string
		setupStore    func() *Store
		key           Key
		value         any
		wantErrSubstr string
	}{
		{
			name:          "returns error when store is nil",
			setupStore:    func() *Store { return nil },
			key:           baseKey,
			value:         1,
			wantErrSubstr: "provider store is nil",
		},
		{
			name:          "returns error when key is nil",
			setupStore:    func() *Store { return New() },
			key:           nil,
			value:         1,
			wantErrSubstr: "provider key is nil",
		},
		{
			name:          "returns error when value is nil",
			setupStore:    func() *Store { return New() },
			key:           NewTaggedKey[int]("number"),
			value:         nil,
			wantErrSubstr: "value is nil",
		},
		{
			name:       "saves value successfully",
			setupStore: func() *Store { return New() },
			key:        baseKey,
			value:      9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.setupStore()
			err := store.Save(tt.key, tt.value, nil)

			if tt.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}

func TestStore_Load(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (*Store, Key)
		wantFound bool
		wantValue any
	}{
		{
			name: "returns not found when store is nil",
			setup: func() (*Store, Key) {
				return nil, NewKey[int]()
			},
			wantFound: false,
		},
		{
			name: "returns not found when key is nil",
			setup: func() (*Store, Key) {
				return New(), nil
			},
			wantFound: false,
		},
		{
			name: "returns not found when key is missing",
			setup: func() (*Store, Key) {
				return New(), NewKey[int]()
			},
			wantFound: false,
		},
		{
			name: "returns stored value when key exists",
			setup: func() (*Store, Key) {
				store := New()
				key := NewTaggedKey[string]("token")
				if err := store.Save(key, "abc", nil); err != nil {
					t.Fatalf("save setup: %v", err)
				}
				return store, key
			},
			wantFound: true,
			wantValue: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, key := tt.setup()
			got, found := store.Load(key)
			if found != tt.wantFound {
				t.Fatalf("expected found=%t, got %t", tt.wantFound, found)
			}
			if !tt.wantFound {
				if got != nil {
					t.Fatalf("expected nil value when not found, got %#v", got)
				}
				return
			}
			if got != tt.wantValue {
				t.Fatalf("expected value %#v, got %#v", tt.wantValue, got)
			}
		})
	}
}

func TestStore_Keys(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Store
		wantSize int
		wantNil  bool
	}{
		{
			name: "nil store returns nil",
			setup: func() *Store {
				return nil
			},
			wantNil: true,
		},
		{
			name: "returns insertion order without duplicates on overwrite",
			setup: func() *Store {
				store := New()
				keyA := NewTaggedKey[int]("A")
				keyB := NewTaggedKey[int]("B")
				if err := store.Save(keyA, 1, nil); err != nil {
					t.Fatalf("save A: %v", err)
				}
				if err := store.Save(keyB, 2, nil); err != nil {
					t.Fatalf("save B: %v", err)
				}
				if err := store.Save(keyA, 3, nil); err != nil {
					t.Fatalf("overwrite A: %v", err)
				}
				return store
			},
			wantSize: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.setup()
			keys := store.Keys()

			if tt.wantNil {
				if keys != nil {
					t.Fatalf("expected nil keys, got %#v", keys)
				}
				return
			}
			if len(keys) != tt.wantSize {
				t.Fatalf("expected %d keys, got %d", tt.wantSize, len(keys))
			}
		})
	}
}

func TestStore_Teardown(t *testing.T) {
	teardownErrA := errors.New("teardown A failed")
	teardownErrB := errors.New("teardown B failed")

	tests := []struct {
		name              string
		setup             func(*testing.T) (*Store, *[]string)
		wantErrs          []error
		wantErrSubstrings []string
		wantOrder         []string
		assertCleared     bool
	}{
		{
			name: "tears down providers in reverse insertion order",
			setup: func(t *testing.T) (*Store, *[]string) {
				store := New()
				order := []string{}
				for _, item := range []struct {
					tag   string
					value int
				}{
					{tag: "first", value: 1},
					{tag: "second", value: 2},
					{tag: "third", value: 3},
				} {
					item := item
					if err := store.Save(
						NewTaggedKey[int](item.tag),
						item.value,
						func(context.Context) error {
							order = append(order, item.tag)
							return nil
						},
					); err != nil {
						t.Fatalf("save %s: %v", item.tag, err)
					}
				}
				return store, &order
			},
			wantOrder:     []string{"third", "second", "first"},
			assertCleared: true,
		},
		{
			name: "joins teardown errors with provider context",
			setup: func(t *testing.T) (*Store, *[]string) {
				store := New()
				order := []string{}
				if err := store.Save(
					NewTaggedKey[int]("A"),
					1,
					func(context.Context) error {
						order = append(order, "A")
						return teardownErrA
					},
				); err != nil {
					t.Fatalf("save A: %v", err)
				}
				if err := store.Save(
					NewTaggedKey[int]("B"),
					2,
					func(context.Context) error {
						order = append(order, "B")
						return teardownErrB
					},
				); err != nil {
					t.Fatalf("save B: %v", err)
				}
				return store, &order
			},
			wantErrs:          []error{teardownErrA, teardownErrB},
			wantErrSubstrings: []string{"provider \"int(A)\"", "provider \"int(B)\""},
			wantOrder:         []string{"B", "A"},
			assertCleared:     true,
		},
		{
			name: "nil store teardown is a no-op",
			setup: func(*testing.T) (*Store, *[]string) {
				order := []string{}
				return nil, &order
			},
			assertCleared: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, order := tt.setup(t)
			err := store.Teardown(context.Background())

			if len(tt.wantErrs) == 0 {
				if err != nil {
					t.Fatalf("expected nil teardown error, got %v", err)
				}
			} else {
				for _, wantErr := range tt.wantErrs {
					if !errors.Is(err, wantErr) {
						t.Fatalf("expected joined error to include %v, got %v", wantErr, err)
					}
				}
				for _, wantText := range tt.wantErrSubstrings {
					if !strings.Contains(err.Error(), wantText) {
						t.Fatalf("expected error to contain %q, got %v", wantText, err)
					}
				}
			}

			if !slices.Equal(*order, tt.wantOrder) {
				t.Fatalf("expected teardown order %v, got %v", tt.wantOrder, *order)
			}

			if tt.assertCleared && store != nil {
				if len(store.Keys()) != 0 {
					t.Fatalf("expected store keys to be cleared after teardown")
				}
			}
		})
	}
}

func TestSaveProvider(t *testing.T) {
	tests := []struct {
		name          string
		value         *int
		teardown      func(context.Context, *int) error
		wantErrSubstr string
		wantLoaded    bool
		wantTeardown  bool
		wantNilValue  bool
	}{
		{
			name:         "saves typed nil pointer value",
			value:        nil,
			wantLoaded:   true,
			wantTeardown: true,
			wantNilValue: true,
		},
		{
			name:         "saves and tears down typed value",
			value:        func() *int { v := 11; return &v }(),
			wantLoaded:   true,
			wantTeardown: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := New()
			key := NewTaggedKey[int]("typed")
			teardownCalled := false
			teardownFn := tt.teardown
			if teardownFn == nil {
				teardownFn = func(_ context.Context, v *int) error {
					if tt.wantTeardown {
						teardownCalled = true
						if tt.wantNilValue {
							if v != nil {
								t.Fatalf("expected nil teardown value, got %#v", v)
							}
							return nil
						}
						if v == nil || *v != 11 {
							t.Fatalf("unexpected teardown value: %#v", v)
						}
					}
					return nil
				}
			}

			err := SaveProvider(store, key, tt.value, teardownFn)
			if tt.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			loaded, found := store.Load(key)
			if found != tt.wantLoaded {
				t.Fatalf("expected loaded=%t, got %t", tt.wantLoaded, found)
			}
			if tt.wantLoaded && loaded == nil {
				t.Fatalf("expected non-nil loaded value")
			}

			if teardownErr := store.Teardown(context.Background()); teardownErr != nil {
				t.Fatalf("unexpected teardown error: %v", teardownErr)
			}
			if tt.wantTeardown && !teardownCalled {
				t.Fatalf("expected teardown callback to be called")
			}
		})
	}
}
