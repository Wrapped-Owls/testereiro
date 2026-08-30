package dbastidor

import (
	"context"
	"testing"
)

func TestNoOpLifecycle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		invoke func(ctx context.Context, lifecycle DBLifecycle) error
	}{
		{name: "create does nothing", invoke: func(ctx context.Context, lifecycle DBLifecycle) error {
			return lifecycle.Create(ctx, "games")
		}},
		{name: "drop does nothing", invoke: func(ctx context.Context, lifecycle DBLifecycle) error {
			return lifecycle.Drop(ctx, "games")
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if err := testCase.invoke(context.Background(), NoOpLifecycle{}); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
