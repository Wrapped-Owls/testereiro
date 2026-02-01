package reqrunner

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// jsonRequest represents a generic HTTP request with input body I.
type jsonRequest[I any] struct {
	Method string
	Path   string
	Body   I
}

// newJSONRequest creates a new jsonRequest.
func newJSONRequest[I any](method, path string, body I) jsonRequest[I] {
	return jsonRequest[I]{
		Method: method,
		Path:   path,
		Body:   body,
	}
}

func (r jsonRequest[I]) MakeRequest(baseURL string) (*http.Request, error) {
	var bodyBuffer bytes.Buffer
	if err := r.EncodeBody(&bodyBuffer); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(r.Method, baseURL+r.Path, &bodyBuffer)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return httpReq, nil
}

func (r jsonRequest[I]) EncodeBody(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	return encoder.Encode(r.Body)
}
