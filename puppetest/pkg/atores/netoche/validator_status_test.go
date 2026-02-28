package netoche

import (
	"context"
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

func TestStatusValidator_MatchingStatus(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
		respStatus int
	}{
		{
			name:       "accepts 201 when expected is 201",
			wantStatus: http.StatusCreated,
			respStatus: http.StatusCreated,
		},
		{
			name:       "accepts 200 when expected is 200",
			wantStatus: http.StatusOK,
			respStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := statusValidator(tt.wantStatus)
			validator.Validate(
				t,
				stgctx.NewRunnerContext(context.Background()),
				&http.Response{StatusCode: tt.respStatus},
			)
		})
	}
}

func TestExtractToState_StoresExtractedValue(t *testing.T) {
	type raw struct{ Token string }

	tests := []struct {
		name       string
		setupCtx   func(stgctx.RunnerContext)
		want       string
		wantStored bool
	}{
		{
			name: "stores extracted token when source exists",
			setupCtx: func(ctx stgctx.RunnerContext) {
				stgctx.SaveOnCtx(ctx, raw{Token: "abc"})
			},
			want:       "abc",
			wantStored: true,
		},
		{
			name: "overwrites existing stored target type with extracted value",
			setupCtx: func(ctx stgctx.RunnerContext) {
				stgctx.SaveOnCtx(ctx, raw{Token: "new-token"})
				stgctx.SaveOnCtx(ctx, "old-token")
			},
			want:       "new-token",
			wantStored: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := stgctx.NewRunnerContext(context.Background())
			tt.setupCtx(ctx)

			runner := New("http://api", ExtractToState(func(v raw) string { return v.Token }))
			if len(runner.validators) != 1 {
				t.Fatalf("expected one validator, got %d", len(runner.validators))
			}

			runner.validators[0].Validate(t, ctx, nil)

			stored, ok := stgctx.LoadFromCtx[string](ctx)
			if ok != tt.wantStored {
				t.Fatalf("expected stored=%t, got %t", tt.wantStored, ok)
			}
			if tt.wantStored && stored != tt.want {
				t.Fatalf("expected extracted value %q, got %q", tt.want, stored)
			}
		})
	}
}
