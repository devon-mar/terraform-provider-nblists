package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	authorizationHeader = "Authorization"
	contentTypeHeader   = "Content-Type"
	contentTypeJSON     = "application/json"
)

type listsClient struct {
	url        string
	auth       string
	allowEmpty bool
	client     *http.Client
}

func newListsClient(url string, token string, allowEmpty bool) *listsClient {
	return &listsClient{
		url:        url,
		allowEmpty: allowEmpty,
		auth:       "Token " + token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *listsClient) get(ctx context.Context, endpoint string, filter map[string][]string) ([]string, error) {
	if !c.allowEmpty && len(filter) == 0 {
		return nil, errors.New("filter is nil or empty")
	}

	fullURL, err := url.JoinPath(c.url, endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(authorizationHeader, c.auth)
	req.Header.Add(contentTypeHeader, contentTypeJSON)

	if filter != nil {
		req.URL.RawQuery = url.Values(filter).Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NetBox returned status code %d", resp.StatusCode)
	} else if ct := resp.Header.Get(contentTypeHeader); ct != contentTypeJSON {
		return nil, fmt.Errorf("invalid content type %q", ct)
	}
	defer resp.Body.Close()

	ret := []string{}
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
