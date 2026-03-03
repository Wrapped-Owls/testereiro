package netoche

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

type responseBody struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestJSONBodyValidator_DecodeBody(t *testing.T) {
	validator := jsonBodyValidator[responseBody]{}
	actual, err := validator.decodeBody(bytes.NewBufferString(`{"id":1,"name":"Prime"}`))
	if err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if actual.ID != 1 || actual.Name != "Prime" {
		t.Fatalf("unexpected decoded value: %#v", actual)
	}
}

func TestJSONBodyValidator_ValidateStoresBody(t *testing.T) {
	ctx := stgctx.NewRunnerContext(t.Context())
	resp := &http.Response{Body: io.NopCloser(bytes.NewBufferString(`{"id":5,"name":"Jazz"}`))}

	validator := jsonBodyValidator[responseBody]{
		Body: responseBody{ID: 5, Name: "Jazz"},
		Comparator: func(t testing.TB, expected, actual responseBody) bool {
			return expected == actual
		},
	}

	validator.Validate(t, ctx, resp)
	stored, ok := stgctx.LoadFromCtx[responseBody](ctx)
	if !ok {
		t.Fatal("expected decoded body in context")
	}
	if stored.ID != 5 || stored.Name != "Jazz" {
		t.Fatalf("unexpected stored value: %#v", stored)
	}
}

func TestExpectBodyOptionsAppendValidators(t *testing.T) {
	runner := New(
		"http://api",
		ExpectBody(responseBody{ID: 1}),
		ExpectBodyWithComparator(
			responseBody{ID: 2},
			func(t testing.TB, expected, actual responseBody) bool {
				return true
			},
		),
	)

	if len(runner.validators) != 2 {
		t.Fatalf("expected two validators, got %d", len(runner.validators))
	}
}

func TestDefaultComparator(t *testing.T) {
	if !defaultComparator(t, responseBody{ID: 1}, responseBody{ID: 1}) {
		t.Fatal("expected equal values to be true")
	}
}

func TestExpectBody_WithSanitizer(t *testing.T) {
	var called bool
	runner := New("http://api",
		ExpectBody(responseBody{ID: 9}, func(expected, actual *responseBody) error {
			called = true
			expected.Name = "x"
			actual.Name = "x"
			return nil
		}),
	)

	ctx := stgctx.NewRunnerContext(t.Context())
	resp := &http.Response{Body: io.NopCloser(bytes.NewBufferString(`{"id":9,"name":"n"}`))}

	asValidator, ok := runner.validators[0].(jsonBodyValidator[responseBody])
	if !ok {
		t.Fatalf("unexpected validator type: %T", runner.validators[0])
	}
	asValidator.Comparator = func(testing.TB, responseBody, responseBody) bool { return true }
	asValidator.Validate(t, ctx, resp)
	if !called {
		t.Fatal("expected sanitizer call")
	}
}

func TestJSONBodyValidator_DecodeError(t *testing.T) {
	validator := jsonBodyValidator[responseBody]{}
	_, err := validator.decodeBody(bytes.NewBufferString("not-json"))
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestExpectBody_SanitizerErrorShape(t *testing.T) {
	runner := New("http://api",
		ExpectBody(responseBody{ID: 1}, func(expected, actual *responseBody) error {
			return errors.New("bad sanitizer")
		}),
	)

	if len(runner.validators) != 1 {
		t.Fatalf("expected one validator, got %d", len(runner.validators))
	}
}
