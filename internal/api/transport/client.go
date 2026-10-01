// Package transport holds the HTTP plumbing shared by every BioT API package: request
// execution, authentication and BioT's error envelope. It knows nothing about templates,
// access control or any other domain.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// TokenSource supplies the bearer token for authenticated requests. It is an interface so
// that transport does not depend on the auth package, which depends on transport.
type TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

const authorizationHeaderKey = "Authorization"

type Client struct {
	// BaseURL is the root of the BioT deployment, e.g. https://api.example.biot-med.com.
	// API packages build their paths from it.
	BaseURL string

	tokens     TokenSource
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// WithTokens returns a copy of the client that authenticates every request. It is separate
// from New because the authenticator needs an unauthenticated client to log in with.
func (c *Client) WithTokens(tokens TokenSource) *Client {
	authenticated := *c
	authenticated.tokens = tokens

	return &authenticated
}

// do issues the request and converts a non-2xx response into an error. On success the caller
// owns the response body and must close it; on every error path the body is closed here.
func (c *Client) do(ctx context.Context, method string, requestURL string, body io.Reader, authenticate bool) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		tflog.Error(ctx, "Failed to create request", map[string]interface{}{
			"method": method,
			"url":    requestURL,
			"error":  err,
		})
		return nil, err
	}

	if authenticate {
		if c.tokens == nil {
			return nil, fmt.Errorf("no token source configured for authenticated request to %s", requestURL)
		}

		token, err := c.tokens.AccessToken(ctx)
		if err != nil {
			return nil, err
		}

		request.Header.Set(authorizationHeaderKey, fmt.Sprintf("Bearer %s", token))
	}

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		tflog.Error(ctx, "Failed to call API", map[string]interface{}{
			"method": method,
			"url":    requestURL,
			"error":  err,
		})
		return nil, err
	}

	if response.StatusCode == http.StatusNotFound {
		defer response.Body.Close()
		return nil, fmt.Errorf("%w: %w", ErrNotFound, ParseAPIError(response))
	}

	if !isResponseOk(response) {
		tflog.Error(ctx, "API returned error status", map[string]interface{}{
			"method":      method,
			"url":         requestURL,
			"status_code": response.StatusCode,
		})
		defer response.Body.Close()
		return nil, ParseAPIError(response)
	}

	return response, nil
}

// Do performs an authenticated request and decodes the JSON response into TResp. A nil body
// sends no request payload.
func Do[TResp any](ctx context.Context, c *Client, method string, url string, body any) (TResp, error) {
	return decodeInto[TResp](ctx, c, method, url, body, true)
}

// DoUnauthenticated is Do without a bearer token, for the login call that produces one.
func DoUnauthenticated[TResp any](ctx context.Context, c *Client, method string, url string, body any) (TResp, error) {
	return decodeInto[TResp](ctx, c, method, url, body, false)
}

// DoNoContent performs an authenticated request and discards the response body, for calls
// like DELETE that answer 204.
func DoNoContent(ctx context.Context, c *Client, method string, url string) error {
	response, err := c.do(ctx, method, url, nil, true)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func decodeInto[TResp any](ctx context.Context, c *Client, method string, url string, body any, authenticate bool) (TResp, error) {
	var zero TResp

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return zero, err
		}
		requestBody = bytes.NewBuffer(encoded)
	}

	response, err := c.do(ctx, method, url, requestBody, authenticate)
	if err != nil {
		return zero, err
	}
	defer response.Body.Close()

	var decoded TResp
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return zero, fmt.Errorf("failed to decode response from %s: %w", url, err)
	}

	return decoded, nil
}

func isResponseOk(response *http.Response) bool {
	return response.StatusCode >= 200 && response.StatusCode < 300
}

// EncodeSearchRequest renders a search request for use as a query-string parameter.
func EncodeSearchRequest(searchRequest map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(searchRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search request: %w", err)
	}

	return url.QueryEscape(string(jsonBytes)), nil
}
