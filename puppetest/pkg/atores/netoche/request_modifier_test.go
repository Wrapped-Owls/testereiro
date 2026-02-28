package netoche

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type strVal string

func (s strVal) String() string { return string(s) }

func TestResolveStringValue(t *testing.T) {
	ctx := stgctx.NewRunnerContext(context.Background())
	stgctx.SaveOnCtx(ctx, 21)

	tests := []struct {
		name      string
		value     any
		want      string
		wantError bool
	}{
		{name: "string", value: "v", want: "v"},
		{name: "func string", value: func() string { return "f" }, want: "f"},
		{name: "stringer", value: strVal("s"), want: "s"},
		{name: "ctx func", value: func(stgctx.RunnerContext) string { return "ctx" }, want: "ctx"},
		{
			name:  "ctx func with error",
			value: func(stgctx.RunnerContext) (string, error) { return "ok", nil },
			want:  "ok",
		},
		{name: "unsupported", value: 10, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveStringValue(ctx, tc.value)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected resolved value: got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestWithHeaderAndPathParamModifiers(t *testing.T) {
	type state struct{ ID int }

	ctx := stgctx.NewRunnerContext(context.Background())
	stgctx.SaveOnCtx(ctx, state{ID: 77})
	req, err := http.NewRequest(http.MethodGet, "http://api/items/{id}", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	runner := New("http://api",
		WithHeader("X-Fixed", "token"),
		WithHeaderFromCtx("X-ID", func(s state) string { return fmt.Sprintf("%d", s.ID) }),
		WithPathParam("id", "42"),
	)

	for _, modifier := range runner.requestModifiers {
		if err = modifier(t, ctx, req); err != nil {
			t.Fatalf("unexpected modifier error: %v", err)
		}
	}

	if got := req.Header.Get("X-Fixed"); got != "token" {
		t.Fatalf("unexpected fixed header: %s", got)
	}
	if got := req.Header.Get("X-ID"); got != "77" {
		t.Fatalf("unexpected ctx header: %s", got)
	}
	if got := req.URL.Path; got != "/items/42" {
		t.Fatalf("unexpected path: %s", got)
	}
}

func TestWithPathParamAndHeaderFromCtxErrors(t *testing.T) {
	type state struct{ Value string }
	ctx := stgctx.NewRunnerContext(context.Background())
	req, err := http.NewRequest(http.MethodGet, "http://api/items/{id}", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	tests := []struct {
		name    string
		option  Option
		wantErr string
	}{
		{
			name:    "path key empty",
			option:  WithPathParam("", "x"),
			wantErr: "path param key is empty",
		},
		{
			name:    "ctx mapper nil for header",
			option:  WithHeaderFromCtx[state]("X", nil),
			wantErr: "header value mapper is nil",
		},
		{
			name:    "ctx mapper nil for path",
			option:  WithPathParamFromCtx[state]("id", nil),
			wantErr: "path param mapper is nil",
		},
		{
			name:    "missing ctx value for header",
			option:  WithHeaderFromCtx[state]("X", func(s state) string { return s.Value }),
			wantErr: "no value in context",
		},
		{
			name:    "missing ctx value for path",
			option:  WithPathParamFromCtx[state]("id", func(s state) string { return s.Value }),
			wantErr: "no value in context",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := New("http://api", tc.option)
			err = runner.requestModifiers[0](t, ctx, req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
