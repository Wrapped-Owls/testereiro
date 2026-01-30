package reqrunner

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// RequestMaker defines how to construct an HTTP request.
type RequestMaker func(baseURL string, t testing.TB, ctx stgctx.RunnerContext) (*http.Request, error)

// RequestModifier updates a request after it is created.
type RequestModifier func(t testing.TB, ctx stgctx.RunnerContext, req *http.Request) error

// RequestExecutor defines the interface for executing HTTP requests.
// It allows replacing the standard http.Client with custom implementations or mocks.
type RequestExecutor interface {
	Do(*http.Request) (*http.Response, error)
}

// HttpRunner is a runner that performs HTTP requests.
type HttpRunner struct {
	BaseURL          string
	reqExec          RequestExecutor
	makeRequest      RequestMaker
	requestModifiers []RequestModifier
	validators       []Validator
}

// Option is a functional option for configuring the HttpRunner.
type Option func(*HttpRunner)

// NewHttpRunner creates a new HttpRunner with the given options.
func NewHttpRunner(baseURL string, opts ...Option) *HttpRunner {
	r := &HttpRunner{
		BaseURL: baseURL,
		reqExec: &http.Client{Timeout: 30 * time.Second}, // Default client
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *HttpRunner) Run(t testing.TB, rCtx stgctx.RunnerContext) error {
	if r.makeRequest == nil {
		t.Fatal("No request configured for HttpRunner")
	}

	// Execute Request
	httpReq, err := r.makeRequest(r.BaseURL, t, rCtx)
	if err != nil {
		t.Fatalf("Impossible to perform request: %v", err)
	}

	for _, modifier := range r.requestModifiers {
		if modifier == nil {
			continue
		}
		if modErr := modifier(t, rCtx, httpReq); modErr != nil {
			t.Fatalf("Failed to modify request: %v", modErr)
		}
	}

	t.Logf("Performing request: %s %s", httpReq.Method, httpReq.URL.String())
	start := time.Now()

	resp, reqErr := r.reqExec.Do(httpReq)
	if reqErr != nil {
		t.Fatalf("Failed to perform request: %s", reqErr.Error())
	}
	defer func(Body io.ReadCloser) {
		if closeErr := Body.Close(); closeErr != nil {
			t.Errorf("Failed to close response body: %v", closeErr)
		}
	}(resp.Body)

	t.Logf("Request performed in %s", time.Since(start))

	// Run validators
	for _, validator := range r.validators {
		validator.Validate(t, rCtx, resp)
	}

	return nil
}
