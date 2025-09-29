package api

import (
	"context"
	"fmt"
)

type APIClient struct {
	BiotSdk       BiotSdk
	authenticator *AuthenticatorService
}

func NewAPIClient(biotSdk BiotSdk, authenticator *AuthenticatorService) *APIClient {
	return &APIClient{
		BiotSdk:       biotSdk,
		authenticator: authenticator,
	}
}

func (apiClient *APIClient) CreateTemplate(ctx context.Context, req CreateTemplateRequest) (TemplateResponse, error) {
	token, err := apiClient.authenticator.GetAccessToken(ctx)

	if err != nil {
		return TemplateResponse{}, err
	}

	response, err := apiClient.BiotSdk.CreateTemplate(ctx, token, req)
	if err != nil {
		return TemplateResponse{}, err
	}

	return response, nil
}

func (apiClient *APIClient) GetTemplate(ctx context.Context, id string) (TemplateResponse, error) {
	token, err := apiClient.authenticator.GetAccessToken(ctx)

	if err != nil {
		return TemplateResponse{}, err
	}

	response, err := apiClient.BiotSdk.GetTemplate(ctx, token, id)
	if err != nil {
		return TemplateResponse{}, err
	}

	return response, nil
}

func (apiClient APIClient) GetTemplateByTypeAndName(ctx context.Context, entityType string, templateName string) (TemplateResponse, error) {
	token, err := apiClient.authenticator.GetAccessToken(ctx)

	if err != nil {
		return TemplateResponse{}, err
	}

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

	response, err := apiClient.BiotSdk.SearchTemplates(ctx, token, searchRequest)
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

func (apiClient *APIClient) UpdateTemplate(ctx context.Context, id string, req UpdateTemplateRequest, force bool) (TemplateResponse, error) {
	token, err := apiClient.authenticator.GetAccessToken(ctx)

	if err != nil {
		return TemplateResponse{}, err
	}

	response, err := apiClient.BiotSdk.UpdateTemplate(ctx, token, id, req, force)
	if err != nil {
		return TemplateResponse{}, err
	}

	return response, nil
}

func (apiClient *APIClient) DeleteTemplate(ctx context.Context, id string) error {
	token, err := apiClient.authenticator.GetAccessToken(ctx)

	if err != nil {
		return err
	}

	return apiClient.BiotSdk.DeleteTemplate(ctx, token, id)
}
