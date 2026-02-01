package reqrunner

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type NoBody struct{}

// WithRequest creates an Option that sets the request for the runner.
func WithRequest[I any](method, path string, body I) Option {
	return func(r *HttpRunner) {
		req := newJSONRequest(method, path, body)
		r.makeRequest = func(baseURL string, _ testing.TB, _ stgctx.RunnerContext) (*http.Request, error) {
			return req.MakeRequest(baseURL)
		}
	}
}

// WithSubsequentRequest creates an Option that allows to generate a request with body changed using previous state objects.
func WithSubsequentRequest[T, I any](method, path string, bodyGen func(T) I) Option {
	return func(r *HttpRunner) {
		r.makeRequest = func(baseURL string, t testing.TB, rCtx stgctx.RunnerContext) (*http.Request, error) {
			if bodyGen == nil {
				t.Fatal("body generation is disabled")
			}

			previousObject, ok := stgctx.LoadFromCtx[T](rCtx)
			if !ok {
				return nil, fmt.Errorf(
					"no previous object in context, failed to find `%T`",
					previousObject,
				)
			}

			inputBody := bodyGen(previousObject)
			req := newJSONRequest(method, path, inputBody)
			return req.MakeRequest(baseURL)
		}
	}
}
