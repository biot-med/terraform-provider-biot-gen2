// Package abac is the API client for BioT access-control rules, actions and conditions.
package abac

import (
	"context"
	"fmt"
	"iter"
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

// SearchConditions yields every condition matching filter, paging through the search endpoint.
// The service can filter conditions on id, type and value.
func (c *Client) SearchConditions(ctx context.Context, filter map[string]transport.SearchFilter) iter.Seq2[ConditionResponse, error] {
	url := fmt.Sprintf("%s/%s/v1/conditions", c.http.BaseURL, servicePrefix)

	return transport.SearchAll[ConditionResponse](ctx, c.http, url, filter)
}

func (c *Client) DeleteCondition(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/v1/conditions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.DoNoContent(ctx, c.http, http.MethodDelete, url)
}

/* Actions */

func (c *Client) CreateAction(ctx context.Context, request CreateActionRequest) (ActionResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/actions", c.http.BaseURL, servicePrefix)

	return transport.Do[ActionResponse](ctx, c.http, http.MethodPost, url, request)
}

func (c *Client) GetAction(ctx context.Context, id string) (ActionResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/actions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[ActionResponse](ctx, c.http, http.MethodGet, url, nil)
}

// UpdateAction issues a PATCH. The service applies partial updates, but the provider
// always sends every mutable field - see the note in models.go.
func (c *Client) UpdateAction(ctx context.Context, id string, request UpdateActionRequest) (ActionResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/actions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[ActionResponse](ctx, c.http, http.MethodPatch, url, request)
}

// SearchActions yields every action matching filter, paging through the search endpoint.
// The service can filter actions on id, type and value.
func (c *Client) SearchActions(ctx context.Context, filter map[string]transport.SearchFilter) iter.Seq2[ActionResponse, error] {
	url := fmt.Sprintf("%s/%s/v1/actions", c.http.BaseURL, servicePrefix)

	return transport.SearchAll[ActionResponse](ctx, c.http, url, filter)
}

func (c *Client) DeleteAction(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/v1/actions/%s", c.http.BaseURL, servicePrefix, id)

	return transport.DoNoContent(ctx, c.http, http.MethodDelete, url)
}

/* Rules */

func (c *Client) CreateRule(ctx context.Context, request CreateRuleRequest) (RuleResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/rules", c.http.BaseURL, servicePrefix)

	return transport.Do[RuleResponse](ctx, c.http, http.MethodPost, url, request)
}

func (c *Client) GetRule(ctx context.Context, id string) (RuleResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/rules/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[RuleResponse](ctx, c.http, http.MethodGet, url, nil)
}

// UpdateRule issues a PATCH. The service applies partial updates, but the provider
// always sends every mutable field - see the note in models.go.
func (c *Client) UpdateRule(ctx context.Context, id string, request UpdateRuleRequest) (RuleResponse, error) {
	url := fmt.Sprintf("%s/%s/v1/rules/%s", c.http.BaseURL, servicePrefix, id)

	return transport.Do[RuleResponse](ctx, c.http, http.MethodPatch, url, request)
}

// SearchRules yields every rule matching filter, paging through the search endpoint.
// The service can filter rules on id, tags, apiId, actionId and conditionId - the last three by
// looking up which rules reference them.
func (c *Client) SearchRules(ctx context.Context, filter map[string]transport.SearchFilter) iter.Seq2[RuleResponse, error] {
	url := fmt.Sprintf("%s/%s/v1/rules", c.http.BaseURL, servicePrefix)

	return transport.SearchAll[RuleResponse](ctx, c.http, url, filter)
}

func (c *Client) DeleteRule(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/%s/v1/rules/%s", c.http.BaseURL, servicePrefix, id)

	return transport.DoNoContent(ctx, c.http, http.MethodDelete, url)
}
