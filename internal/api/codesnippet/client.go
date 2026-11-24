// Package codesnippet is the API client for BioT code snippets, served by the settings service.
//
// A snippet cannot be changed once created, and BioT deletes snippets nothing uses once they are a
// day old, so they behave as values owned by whatever uses them rather than as objects of their own.
package codesnippet

import (
	"context"
	"fmt"
	"net/http"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

const servicePrefix = "settings"

// LanguageJavaScript is the only language BioT supports for code snippets.
const LanguageJavaScript = "JAVASCRIPT"

type Client struct {
	http *transport.Client
}

func NewClient(httpClient *transport.Client) *Client {
	return &Client{http: httpClient}
}

type CreateRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Language    string  `json:"language"`
	Code        string  `json:"code"`
}

type Response struct {
	ID          string  `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Language    string  `json:"language"`
	Code        string  `json:"code"`
}

func (c *Client) Create(ctx context.Context, request CreateRequest) (Response, error) {
	url := fmt.Sprintf("%s/%s/v1/code-snippets", c.http.BaseURL, servicePrefix)

	return transport.Do[Response](ctx, c.http, http.MethodPost, url, request)
}

func (c *Client) Get(ctx context.Context, id string) (Response, error) {
	url := fmt.Sprintf("%s/%s/v1/code-snippets/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[Response](ctx, c.http, http.MethodGet, url, nil)
}
