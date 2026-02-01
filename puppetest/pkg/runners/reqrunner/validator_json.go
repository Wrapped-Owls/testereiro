package reqrunner

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// ExpectBody adds a validator for the response body.
func ExpectBody[O any](expected O, sanitizer ...respObjectSanitizer[O]) Option {
	return func(r *HttpRunner) {
		newValidator := jsonBodyValidator[O]{Body: expected}
		if len(sanitizer) > 0 {
			newValidator.Sanitizer = sanitizer[0]
		}

		r.validators = append(r.validators, newValidator)
	}
}

// jsonBodyValidator defines expected results for a request body.
type jsonBodyValidator[O any] struct {
	Body      O
	Sanitizer respObjectSanitizer[O]
}

func (v jsonBodyValidator[O]) decodeBody(reader io.Reader) (O, error) {
	decoder := json.NewDecoder(reader)

	output := new(O)
	err := decoder.Decode(output)
	return *output, err
}

func (v jsonBodyValidator[O]) Validate(
	t testing.TB,
	rCtx stgctx.RunnerContext,
	resp *http.Response,
) {
	// Unmarshal response
	actualBody, err := v.decodeBody(resp.Body)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Persist the decoded body to storage automatically
	stgctx.SaveOnCtx(rCtx, actualBody)

	// If Sanitizer is provided, use it
	expectedBody := v.Body
	if v.Sanitizer != nil {
		if err = v.Sanitizer(&expectedBody, &actualBody); err != nil {
			t.Fatalf("Failed to sanitize response body: %v", err)
		}
	}

	assert.Equal(t, expectedBody, actualBody, "response body mismatch")
}
