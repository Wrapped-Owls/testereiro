package netoche

import (
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type requestBody struct {
	ID int `json:"id"`
}

func TestWithRequest_SetsStaticRequestMaker(t *testing.T) {
	runner := New("http://example.test", WithRequest("PATCH", "/path", requestBody{ID: 9}))
	ctx := stgctx.NewRunnerContext(t.Context())

	req, err := runner.makeRequest(runner.BaseURL, t, ctx)
	if err != nil {
		t.Fatalf("unexpected error making request: %v", err)
	}
	if req.Method != http.MethodPatch {
		t.Fatalf("unexpected method: %s", req.Method)
	}
	if req.URL.String() != "http://example.test/path" {
		t.Fatalf("unexpected url: %s", req.URL.String())
	}
}

func TestWithSubsequentRequest(t *testing.T) {
	type state struct{ Token string }
	type payload struct {
		Auth string `json:"auth"`
	}

	tests := []struct {
		name      string
		setupCtx  func(stgctx.RunnerContext)
		bodyGen   func(state) payload
		wantError bool
		wantPath  string
	}{
		{
			name:      "missing state in context",
			setupCtx:  func(stgctx.RunnerContext) {},
			bodyGen:   func(s state) payload { return payload{Auth: s.Token} },
			wantError: true,
		},
		{
			name: "state present",
			setupCtx: func(ctx stgctx.RunnerContext) {
				stgctx.SaveOnCtx(ctx, state{Token: "abc"})
			},
			bodyGen:   func(s state) payload { return payload{Auth: s.Token} },
			wantPath:  "http://api.test/sessions",
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := New("http://api.test", WithSubsequentRequest("POST", "/sessions", tc.bodyGen))
			ctx := stgctx.NewRunnerContext(t.Context())
			tc.setupCtx(ctx)

			req, err := runner.makeRequest(runner.BaseURL, t, ctx)
			if tc.wantError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if req.URL.String() != tc.wantPath {
				t.Fatalf("unexpected request url: %s", req.URL.String())
			}
		})
	}
}
