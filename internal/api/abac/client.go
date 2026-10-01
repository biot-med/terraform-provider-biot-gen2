// Package abac is the API client for BioT access-control rules, actions and conditions.
package abac

import (
	"context"
	"fmt"
	"net/http"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

const servicePrefix = "access-control"

type Client struct {
	http *transport.Client
}

func NewClient(httpClient *transport.Client) *Client {
	return &Client{http: httpClient}
}

/* Conditions */

func (c *Client) CreateCondition(ctx context.Context, request CreateConditionRequest) (ConditionResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/conditions", c.http.BaseURL, servicePrefix)

	return transport.Do[ConditionResponse](ctx, c.http, http.MethodPost, url, request)
}

func (c *Client) GetCondition(ctx context.Context, id string) (ConditionResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/conditions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[ConditionResponse](ctx, c.http, http.MethodGet, url, nil)
}

// UpdateCondition issues a PATCH. The service applies partial updates, but the provider
// always sends every mutable field - see the note in models.go.
func (c *Client) UpdateCondition(ctx context.Context, id string, request UpdateConditionRequest) (ConditionResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/conditions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[ConditionResponse](ctx, c.http, http.MethodPatch, url, request)
}

func (c *Client) DeleteCondition(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/v1/conditions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.DoNoContent(ctx, c.http, http.MethodDelete, url)
}
