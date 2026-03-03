package netoche

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type closeSpyBody struct {
	io.Reader
	closed bool
}

func (b *closeSpyBody) Close() error {
	b.closed = true
	return nil
}

type fakeExecutor struct {
	resp       *http.Response
	calledWith *http.Request
}

func (f *fakeExecutor) Do(req *http.Request) (*http.Response, error) {
	f.calledWith = req
	return f.resp, nil
}

type fakeValidator struct {
	calls int
}

func (f *fakeValidator) Validate(t testing.TB, rCtx stgctx.RunnerContext, resp *http.Response) {
	f.calls++
}

func TestHttpRunner_RunSuccess(t *testing.T) {
	body := &closeSpyBody{Reader: bytes.NewBufferString(`{"ok":true}`)}
	exec := &fakeExecutor{resp: &http.Response{StatusCode: http.StatusOK, Body: body}}
	validator := &fakeValidator{}

	runner := New("http://api.test",
		WithRequest(http.MethodGet, "/health", NoBody{}),
		WithHeader("X-Test", "true"),
	)
	runner.reqExec = exec
	runner.validators = append(runner.validators, validator)

	ctx := stgctx.NewRunnerContext(t.Context())
	if err := runner.Run(t, ctx); err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}

	if exec.calledWith == nil {
		t.Fatal("expected request execution")
	}
	if got := exec.calledWith.Header.Get("X-Test"); got != "true" {
		t.Fatalf("unexpected header value: %s", got)
	}
	if validator.calls != 1 {
		t.Fatalf("expected validator call, got %d", validator.calls)
	}
	if !body.closed {
		t.Fatal("expected response body to be closed")
	}
}

func TestWithRequestModifier_NilModifierIgnored(t *testing.T) {
	runner := New("http://api", WithRequestModifier(nil))
	if len(runner.requestModifiers) != 0 {
		t.Fatalf("expected no modifiers, got %d", len(runner.requestModifiers))
	}
}
