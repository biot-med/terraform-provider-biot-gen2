// Package portalbuilder is the API client for BioT portal builder views, served by the
// settings service.
//
// BioT creates and deletes views itself (one per portal, view type and template or entity), so
// the API offers no create or delete - only list, get and a full-replace update.
package portalbuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

const servicePrefix = "settings"

// PortalTypes are the portals that have views.
var PortalTypes = []string{"MANUFACTURER_PORTAL", "ORGANIZATION_PORTAL"}

type Client struct {
	http *transport.Client
}

func NewClient(httpClient *transport.Client) *Client {
	return &Client{http: httpClient}
}

// ListViews returns every view of a portal, without their configuration.
func (c *Client) ListViews(ctx context.Context, portalType string) ([]View, error) {
	url := fmt.Sprintf("%s/%s/v1/portal-builder/%s/views", c.http.BaseURL, servicePrefix, portalType)

	response, err := transport.Do[ListViewsResponse](ctx, c.http, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return response.Views, nil
}

func (c *Client) GetView(ctx context.Context, portalType string, viewID string) (ViewResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/portal-builder/%s/views/%s", c.http.BaseURL, servicePrefix, portalType, viewID)

	return transport.Do[ViewResponse](ctx, c.http, http.MethodGet, url, nil)
}

// UpdateView replaces the view's whole configuration: a section missing from config is cleared.
func (c *Client) UpdateView(ctx context.Context, portalType string, viewID string, config json.RawMessage) (ViewResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/portal-builder/%s/views/%s", c.http.BaseURL, servicePrefix, portalType, viewID)

	return transport.Do[ViewResponse](ctx, c.http, http.MethodPut, url, config)
}
