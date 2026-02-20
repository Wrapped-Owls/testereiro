package reqrunner

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// ExpectBody adds a validator for the response body.
func ExpectBody[O any](expected O, sanitizer ...respObjectSanitizer[O]) Option {
	return func(r *HttpRunner) {
		newValidator := jsonBodyValidator[O]{Body: expected, Comparator: defaultComparator[O]}
		if len(sanitizer) > 0 {
			newValidator.Sanitizer = sanitizer[0]
		}

		r.validators = append(r.validators, newValidator)
	}
}

func ExpectBodyWithComparator[O any](
	expected O, comparator respObjectComparator[O],
) Option {
	return func(r *HttpRunner) {
		r.validators = append(r.validators, jsonBodyValidator[O]{
			Body:       expected,
			Comparator: comparator,
		})
	}
}

// jsonBodyValidator defines expected results for a request body.
type jsonBodyValidator[O any] struct {
	Body       O
	Sanitizer  respObjectSanitizer[O]
	Comparator respObjectComparator[O]
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
	expectedBody := v.Body

	// If no comparator is provided, use sanitizer + default reflect comparator.
	if v.Sanitizer != nil {
		if err = v.Sanitizer(&expectedBody, &actualBody); err != nil {
			t.Fatalf("failed to sanitize response body: %v", err)
		}
	}

	if v.Comparator != nil {
		if !v.Comparator(t, expectedBody, actualBody) {
			t.Fatalf("response body mismatch")
		}
		return
	}
}

func defaultComparator[O any](t testing.TB, expected, actual O) bool {
	if reflect.DeepEqual(expected, actual) {
		return true
	}

	t.Errorf("response body mismatch: expected=%#v actual=%#v", expected, actual)
	return false
}
