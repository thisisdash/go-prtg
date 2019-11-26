package prtgapi

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	srv := httptest.NewServer(mux)
	u, _ := url.Parse(srv.URL)
	client = NewClient(*u, "testsuite", "987654321", "prtgapi testsuite", srv.Client())

	return client, mux, srv.URL, srv.Close
}

func testMethod(t *testing.T, request *http.Request, method string) {
	if request.Method != method {
		t.Errorf("Expected method %s, got method %s", method, request.Method)
	}
}

func testParams(t *testing.T, request *http.Request, params map[string]string) {
	for key, value := range params {
		if realValue := request.URL.Query().Get(key); realValue != value {
			t.Errorf("Expected parameter %s to have value %s, instead it has value %s", key, value, realValue)
		}
	}
}

func testAuthentication(t *testing.T, request *http.Request) {
	username := request.URL.Query().Get("username")
	passhash := request.URL.Query().Get("passhash")
	if username != "testsuite" || passhash != "987654321" {
		t.Errorf("Expected authentication with username testsuite and passhash 987654321, got username %s and passhash %s", username, passhash)
	}
}
