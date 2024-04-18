package provider

import (
	"context"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewListsClient(t *testing.T) {
	token := "abcde"
	url := "https://netbox.example.com"

	c := newListsClient(url, token, true, 10)
	if c.url != url {
		t.Errorf("got url %s, want %s", c.url, url)
	}
	if want := "Token " + token; c.auth != want {
		t.Errorf("got auth %s, want %s", c.auth, want)
	}
	if !c.allowEmpty {
		t.Errorf("expected allowEmpty=true")
	}
	if c.client == nil {
		t.Errorf("client was nil")
	}
}

func TestGet(t *testing.T) {
	tests := map[string]struct {
		endpoint   string
		allowEmpty bool
		filter     map[string][]string
		want       []string
		wantError  bool
	}{
		"nofilter": {
			endpoint:   "/nofilter",
			want:       []string{"192.0.2.1", "192.0.2.2"},
			allowEmpty: true,
		},
		"nofilter,allow_empty=false": {
			endpoint:   "/nofilter",
			wantError:  true,
			allowEmpty: false,
		},
		"emptylist": {
			endpoint:   "/empty",
			want:       []string{},
			allowEmpty: true,
		},
		"wrongcontenttype": {
			endpoint:   "/empty",
			filter:     map[string][]string{"ct": {"application/json"}},
			wantError:  true,
			allowEmpty: true,
		},
		"wrongcharset": {
			endpoint:   "/empty",
			filter:     map[string][]string{"ct": {"text/plain; charset=abc"}},
			wantError:  true,
			allowEmpty: true,
		},
		"filter": {
			endpoint:   "filter",
			filter:     map[string][]string{"family": {"4"}},
			want:       []string{"192.0.2.10/32"},
			allowEmpty: false,
		},
		"nocharset": {
			endpoint: "/nocharset",
			filter:   map[string][]string{"ct": {"text/plain"}},
			want:     []string{"192.0.2.42/32"},
		},
		"not_found": {
			endpoint:   "404",
			wantError:  true,
			allowEmpty: true,
		},
	}

	token := "abcd12345"
	h := newTestListsHandler(t, token)
	h.addList("nofilter", nil, []string{"192.0.2.1", "192.0.2.2"})
	h.addList("filter", map[string][]string{"family": {"4"}}, []string{"192.0.2.10/32"})
	h.addList("empty", nil, []string{})
	h.addList("empty", map[string][]string{"ct": {"application/json"}}, []string{})
	h.addList("empty", map[string][]string{"ct": {"text/plain; charset=abc"}}, []string{})
	h.addList("nocharset", map[string][]string{"ct": {"text/plain"}}, []string{"192.0.2.42/32"})
	s := httptest.NewServer(h)
	defer s.Close()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			listsClient := newListsClient(s.URL+"/api/plugins/lists", token, tc.allowEmpty, 10)
			have, err := listsClient.get(context.Background(), tc.endpoint, tc.filter)
			if err == nil && tc.wantError {
				t.Errorf("expected an error")
			} else if err != nil && !tc.wantError {
				t.Fatalf("expected no error but got: %v", err)
			}
			if !reflect.DeepEqual(have, tc.want) {
				t.Errorf("got list %v, want %v", have, tc.want)
			}
		})
	}
}
