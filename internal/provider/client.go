package provider

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"time"
)

const (
	headerAuthorization = "Authorization"
	headerAccept        = "Accept"
	contentTypeHeader   = "Content-Type"
	mediaTypeText       = "text/plain"
)

type listsClient struct {
	url        string
	auth       string
	allowEmpty bool
	client     *http.Client
}

func newListsClient(url string, token string, allowEmpty bool, timeout int) *listsClient {
	return &listsClient{
		url:        url,
		allowEmpty: allowEmpty,
		auth:       "Token " + token,
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
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
	req.Header.Add(headerAuthorization, c.auth)
	req.Header.Add(headerAccept, mediaTypeText)

	if filter != nil {
		req.URL.RawQuery = url.Values(filter).Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NetBox returned status code %d", resp.StatusCode)
	}

	ct := resp.Header.Get(contentTypeHeader)

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", contentTypeHeader, err)
	}
	if mediaType != mediaTypeText || (params["charset"] != "" && params["charset"] != "utf-8") {
		return nil, fmt.Errorf("invalid content type %q", ct)
	}

	ret := []string{}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		ret = append(ret, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return ret, nil
}
