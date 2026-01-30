package reqrunner

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// WithRequestModifier creates an Option that modifies the generated request.
func WithRequestModifier(modifier RequestModifier) Option {
	return func(r *HttpRunner) {
		if modifier == nil {
			return
		}
		r.requestModifiers = append(r.requestModifiers, modifier)
	}
}

// WithHeader adds a header to the request.
func WithHeader(key string, value any) Option {
	return WithRequestModifier(
		func(_ testing.TB, rCtx stgctx.RunnerContext, req *http.Request) error {
			val, err := resolveStringValue(rCtx, value)
			if err != nil {
				return err
			}
			req.Header.Set(key, val)
			return nil
		},
	)
}

// WithHeaderFromCtx adds a header using a value loaded from the runner context.
func WithHeaderFromCtx[T any](key string, mapper func(T) string) Option {
	return WithHeader(
		key, func(rCtx stgctx.RunnerContext) (string, error) {
			if mapper == nil {
				return "", fmt.Errorf("header value mapper is nil")
			}
			val, ok := stgctx.LoadFromCtx[T](rCtx)
			if !ok {
				return "", fmt.Errorf("no value in context for %T", val)
			}
			return mapper(val), nil
		},
	)
}

// WithPathParam replaces a path variable in the request path.
func WithPathParam(key string, value any) Option {
	return WithRequestModifier(
		func(_ testing.TB, rCtx stgctx.RunnerContext, req *http.Request) error {
			if key == "" {
				return fmt.Errorf("path param key is empty")
			}

			val, err := resolveStringValue(rCtx, value)
			if err != nil {
				return err
			}

			bracedToken := "{" + key + "}"
			req.URL.Path = strings.ReplaceAll(req.URL.Path, bracedToken, val)
			return nil
		},
	)
}

// WithPathParamFromCtx replaces a path variable using a value loaded from the runner context.
func WithPathParamFromCtx[T any](key string, mapper func(T) string) Option {
	return WithPathParam(
		key, func(rCtx stgctx.RunnerContext) (string, error) {
			if mapper == nil {
				return "", fmt.Errorf("path param mapper is nil")
			}
			val, ok := stgctx.LoadFromCtx[T](rCtx)
			if !ok {
				return "", fmt.Errorf("no value in context for %T", val)
			}
			return mapper(val), nil
		},
	)
}

func resolveStringValue(rCtx stgctx.RunnerContext, value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case func() string:
		return v(), nil
	case fmt.Stringer:
		return v.String(), nil
	case func(stgctx.RunnerContext) string:
		return v(rCtx), nil
	case func(stgctx.RunnerContext) (string, error):
		return v(rCtx)
	default:
		return "", fmt.Errorf("unsupported value type %T", value)
	}
}
