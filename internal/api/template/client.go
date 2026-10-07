// Package template is the API client for BioT templates, served by the settings service.
package template

import (
	"context"
	"fmt"
	"net/http"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

const servicePrefix = "settings"

type Client struct {
	http *transport.Client
}

func NewClient(httpClient *transport.Client) *Client {
	return &Client{http: httpClient}
}

func (c *Client) Create(ctx context.Context, request CreateTemplateRequest) (TemplateResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/templates", c.http.BaseURL, servicePrefix)

	return transport.Do[TemplateResponse](ctx, c.http, http.MethodPost, url, request)
}

func (c *Client) Get(ctx context.Context, id string) (TemplateResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/templates/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[TemplateResponse](ctx, c.http, http.MethodGet, url, nil)
}

func (c *Client) Update(ctx context.Context, id string, request UpdateTemplateRequest, force bool) (TemplateResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/templates/%s", c.http.BaseURL, servicePrefix, id)
	if force {
		url += "?force=true"
	}

	return transport.Do[TemplateResponse](ctx, c.http, http.MethodPut, url, request)
}

func (c *Client) Delete(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/v1/templates/%s", c.http.BaseURL, servicePrefix, id)

	return transport.DoNoContent(ctx, c.http, http.MethodDelete, url)
}

func (c *Client) Search(ctx context.Context, searchRequest map[string]interface{}) (SearchTemplatesResponse, error) {
	encoded, err := transport.EncodeSearchRequest(searchRequest)
	if err != nil {
		return SearchTemplatesResponse{}, err
	}

	url := fmt.Sprintf("%s/%s/v1/templates?searchRequest=%s", c.http.BaseURL, servicePrefix, encoded)

	return transport.Do[SearchTemplatesResponse](ctx, c.http, http.MethodGet, url, nil)
}

func (c *Client) GetByTypeAndName(ctx context.Context, entityType string, templateName string) (TemplateResponse, error) {
	searchRequest := map[string]interface{}{
		"filter": map[string]interface{}{
			"entityTypeName": map[string]interface{}{
				"in": []string{entityType},
			},
			"name": map[string]interface{}{
				"in": []string{templateName},
			},
		},
	}

	response, err := c.Search(ctx, searchRequest)
	if err != nil {
		return TemplateResponse{}, err
	}

	if response.Metadata.Page.TotalResults != 1 {
		return TemplateResponse{}, fmt.Errorf(
			"unexpected number of results for template with name=%q and type=%q: expected 1, got %d",
			templateName, entityType, response.Metadata.Page.TotalResults,
		)
	}

	return response.Data[0], nil
}
