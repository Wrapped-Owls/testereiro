package atores

import (
	"errors"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type runnerFunc func(testing.TB, stgctx.RunnerContext) error

func (f runnerFunc) Run(t testing.TB, ctx stgctx.RunnerContext) error {
	return f(t, ctx)
}

func TestMultiRunner_Run(t *testing.T) {
	var firstCtx stgctx.RunnerContext
	calledSecond := false

	mr := MultiRunner{Runners: []Runner{
		runnerFunc(func(_ testing.TB, ctx stgctx.RunnerContext) error {
			firstCtx = ctx
			stgctx.SaveOnCtx(ctx, "saved")
			return nil
		}),
		runnerFunc(func(_ testing.TB, ctx stgctx.RunnerContext) error {
			calledSecond = true
			if firstCtx != ctx {
				t.Fatal("expected the same context instance across runners")
			}
			if val, ok := stgctx.LoadFromCtx[string](ctx); !ok || val != "saved" {
				t.Fatalf("expected shared context value, got %q ok=%v", val, ok)
			}
			return nil
		}),
	}}

	ctx := stgctx.NewRunnerContext(t.Context())
	if err := mr.Run(t, ctx); err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if !calledSecond {
		t.Fatal("expected second runner to be called")
	}
}

func TestMultiRunner_RunStopsOnError(t *testing.T) {
	wantErr := errors.New("stop")
	calledSecond := false

	mr := MultiRunner{Runners: []Runner{
		runnerFunc(func(testing.TB, stgctx.RunnerContext) error { return wantErr }),
		runnerFunc(func(testing.TB, stgctx.RunnerContext) error {
			calledSecond = true
			return nil
		}),
	}}

	ctx := stgctx.NewRunnerContext(t.Context())
	err := mr.Run(t, ctx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
	if calledSecond {
		t.Fatal("expected execution to stop after first error")
	}
}
