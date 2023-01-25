package provider

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
)

type testListsHandler struct {
	token string
	lists map[string][]string
	t     *testing.T
}

func newTestListsHandler(t *testing.T, token string) *testListsHandler {
	return &testListsHandler{
		token: token,
		lists: map[string][]string{},
		t:     t,
	}
}

func (h *testListsHandler) addList(endpoint string, params map[string][]string, list []string) {
	uri, err := url.JoinPath("/api/plugins/lists", endpoint)
	if err != nil {
		panic(err)
	}
	if len(params) > 0 {
		uri += "?" + url.Values(params).Encode()
	}
	h.t.Logf("testServer: adding list for uri %s", uri)
	h.lists[uri] = list
}

// ServeHTTP implements http.Handler
func (h *testListsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.t.Logf("testServer: %s %s", r.Method, r.RequestURI)

	if r.Method != http.MethodGet {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
	}
	if r.Header.Get(contentTypeHeader) != contentTypeJSON {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}
	if r.Header.Get(authorizationHeader) != "Token "+h.token {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	list, ok := h.lists[r.RequestURI]
	if !ok {
		http.Error(w, "invalid request URI", http.StatusNotFound)
		return
	}

	// allow overriding the content type
	if ct := r.URL.Query().Get("ct"); ct != "" {
		w.Header().Add(contentTypeHeader, ct)
	} else {
		w.Header().Add(contentTypeHeader, contentTypeJSON)
	}
	json.NewEncoder(w).Encode(list)
}
