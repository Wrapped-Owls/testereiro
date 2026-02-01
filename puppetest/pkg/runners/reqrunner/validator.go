package reqrunner

import (
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// Validator defines how to validate an HTTP response.
type (
	respObjectSanitizer[O any] func(expected, actual *O) error
	Validator                  interface {
		Validate(t testing.TB, rCtx stgctx.RunnerContext, resp *http.Response)
	}
)

// ExtractToState adds a validator that extracts a value from a previously stored response
// and saves it to storage.
func ExtractToState[R any, V any](extractor func(resp R) V) Option {
	return func(r *HttpRunner) {
		r.validators = append(r.validators, &extractorValidator[R, V]{extractor: extractor})
	}
}

type extractorValidator[R any, V any] struct {
	extractor func(resp R) V
}

func (v *extractorValidator[R, V]) Validate(
	t testing.TB,
	rCtx stgctx.RunnerContext,
	_ *http.Response,
) {
	respBody, ok := stgctx.LoadFromCtx[R](rCtx)
	if !ok {
		t.Fatalf("Failed to load response of type %T from storage for extraction", respBody)
	}

	val := v.extractor(respBody)
	stgctx.SaveOnCtx(rCtx, val)
}
