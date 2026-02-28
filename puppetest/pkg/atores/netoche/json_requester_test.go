package netoche

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestJSONRequest_MakeRequest(t *testing.T) {
	req := newJSONRequest("POST", "/users", map[string]any{"name": "Optimus"})

	httpReq, err := req.MakeRequest("http://localhost")
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}

	if httpReq.Method != "POST" {
		t.Fatalf("unexpected method: %s", httpReq.Method)
	}
	if httpReq.URL.String() != "http://localhost/users" {
		t.Fatalf("unexpected url: %s", httpReq.URL.String())
	}
	if got := httpReq.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %s", got)
	}
}

func TestJSONRequest_EncodeBody(t *testing.T) {
	req := newJSONRequest("PUT", "/x", map[string]any{"power": 100})
	buf := new(bytes.Buffer)

	if err := req.EncodeBody(buf); err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode encoded json: %v", err)
	}
	if got := int(decoded["power"].(float64)); got != 100 {
		t.Fatalf("unexpected power value: %d", got)
	}
}
