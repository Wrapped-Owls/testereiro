package reqrunner

import (
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// ExpectStatus adds a validator for the response status code.
func ExpectStatus(code int) Option {
	return func(r *HttpRunner) {
		r.validators = append(r.validators, statusValidator(code))
	}
}

type statusValidator int

func (v statusValidator) Validate(t testing.TB, rCtx stgctx.RunnerContext, resp *http.Response) {
	if int(v) != resp.StatusCode {
		t.Fatalf("Unexpected status code: got %v want %v", resp.StatusCode, int(v))
	}
}
