package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// JSONRequest creates an HTTP request with a JSON body and Content-Type header set.
func JSONRequest(method, path string, body any) *http.Request {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}
