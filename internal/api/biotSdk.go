package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Jwt struct {
	Token      string `json:"accessToken"`
	Expiration string `json:"accessTokenExpiration"`
}

// TODO: Move to seperate file - Login Model ?
type LoginResponse struct {
	UserId              string `json:"userId"`
	OwnerOrganizationId string `json:"ownerOrganizationId"`
	AccessJwt           Jwt    `json:"accessJwt"`
	RefreshJwt          Jwt    `json:"refreshJwt"`
}

/* Errors: */
type errorCodesStruct struct {
	NotFound       error
	Unauthorized   error
	InvalidRequest error
	// Add more as needed
}

// Exported instance of the struct
var SpecificErrorCodes = errorCodesStruct{
	NotFound: errors.New("resource not found"),
}

type BiotSdk interface {
	LoginAsService(ctx context.Context, seviceId string, serviceSecretKey string) (Jwt, error)
	CreateTemplate(ctx context.Context, accessToken string, request CreateTemplateRequest) (TemplateResponse, error)
	UpdateTemplate(ctx context.Context, accessToken string, id string, request UpdateTemplateRequest) (TemplateResponse, error)
	GetTemplate(ctx context.Context, token string, id string) (TemplateResponse, error)
	DeleteTemplate(ctx context.Context, accessToken string, id string) error
	SearchTemplates(ctx context.Context, token string, searchrequest map[string]interface{}) (SearchTemplatesResponse, error)
	ValidateVersions(ctx context.Context, accessToken string, terraformProviderVersion string, minimumBiotVersion string) (TerraformVersionValidationResponse, error)
}

const (
	umsPrefix      = "ums"
	settingsPrefix = "settings"

	authorizationHeaderKey = "Authorization"
)

var httpClient = &http.Client{}

type biotSdkImpl struct {
	baseUrl string
}

func (biotSdkImpl biotSdkImpl) LoginAsService(ctx context.Context, serviceId string, serviceSecretKey string) (Jwt, error) {
	var url = fmt.Sprintf("%s/%s/v2/services/accessToken", biotSdkImpl.baseUrl, umsPrefix)

	requestBody, err := json.Marshal(map[string]string{
		"id":        serviceId,
		"secretKey": serviceSecretKey,
	})

	if err != nil {
		return Jwt{}, err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		tflog.Warn(ctx, "Failed to create login request", map[string]interface{}{
			"url":   url,
			"error": err,
		})
		return Jwt{}, err
	}

	request.Header.Set("Content-Type", "application/json")
	response, requestError := httpClient.Do(request)

	if requestError != nil {
		tflog.Warn(ctx, "Failed to call login API", map[string]interface{}{
			"url":   url,
			"error": requestError,
		})
		return Jwt{}, requestError
	}

	if !isResponseOk(response) {
		tflog.Warn(ctx, "Login API returned non-200 status", map[string]interface{}{
			"url":         url,
			"status_code": response.StatusCode,
		})
		return Jwt{}, getErrorMessage(response)
	}

	defer response.Body.Close()

	var JwtResponse Jwt
	if err := json.NewDecoder(response.Body).Decode(&JwtResponse); err != nil {
		return Jwt{}, fmt.Errorf("failed to decode JWT response: %w", err)
	}

	return JwtResponse, nil
}

func (biotSdkImpl biotSdkImpl) CreateTemplate(ctx context.Context, accessToken string, request CreateTemplateRequest) (TemplateResponse, error) {

	var url = fmt.Sprintf("%s/%s/v1/templates", biotSdkImpl.baseUrl, settingsPrefix)

	jsonBody, _ := json.Marshal(request)

	httpResponse, err := biotSdkImpl.crudTemplateHelper(ctx, accessToken, url, http.MethodPost, bytes.NewBuffer(jsonBody))
	if err != nil {
		return TemplateResponse{}, err
	}
	defer httpResponse.Body.Close()

	return getTemplateResponseBody(httpResponse)
}

func (biotSdkImpl biotSdkImpl) UpdateTemplate(ctx context.Context, accessToken string, id string, request UpdateTemplateRequest) (TemplateResponse, error) {
	// TODO: Add ?force=true ????
	var url = fmt.Sprintf("%s/%s/v1/templates/%s", biotSdkImpl.baseUrl, settingsPrefix, id)

	jsonBody, _ := json.Marshal(request)
	httpResponse, err := biotSdkImpl.crudTemplateHelper(ctx, accessToken, url, http.MethodPut, bytes.NewBuffer(jsonBody))
	if err != nil {
		return TemplateResponse{}, err
	}
	defer httpResponse.Body.Close()

	return getTemplateResponseBody(httpResponse)
}

func (biotSdkImpl biotSdkImpl) GetTemplate(ctx context.Context, accessToken string, id string) (TemplateResponse, error) {
	var url = fmt.Sprintf("%s/%s/v1/templates/%s", biotSdkImpl.baseUrl, settingsPrefix, id)

	httpResponse, err := biotSdkImpl.crudTemplateHelper(ctx, accessToken, url, http.MethodGet, nil)
	if err != nil {
		return TemplateResponse{}, err
	}
	defer httpResponse.Body.Close()

	return getTemplateResponseBody(httpResponse)
}

func (biotSdkImpl biotSdkImpl) DeleteTemplate(ctx context.Context, accessToken string, id string) error {
	var url = fmt.Sprintf("%s/%s/v1/templates/%s", biotSdkImpl.baseUrl, settingsPrefix, id)

	httpResponse, err := biotSdkImpl.crudTemplateHelper(ctx, accessToken, url, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()

	return err
}

// Using this function requires the user to close the httpResponse body (httpResponse.Body.Close()
// Only in the cases where the response returned with status OK (200 / 201 / 2xx...)
// In case of errors, the body will be closed within this funciton.
func (biotSdkImpl biotSdkImpl) crudTemplateHelper(ctx context.Context, accessToken string, url string, method string, body io.Reader) (*http.Response, error) {
	req, requestErr := http.NewRequest(method, url, body)
	if requestErr != nil {
		tflog.Error(ctx, "Failed to create template request", map[string]interface{}{
			"method": method,
			"url":    url,
			"error":  requestErr,
		})
		return nil, requestErr
	}

	req.Header.Set(authorizationHeaderKey, fmt.Sprintf("Bearer %s", accessToken))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	var httpResponse, err = httpClient.Do(req)

	if err != nil {
		tflog.Error(ctx, "Failed to call template API", map[string]interface{}{
			"method": method,
			"url":    url,
			"error":  err,
		})
		return nil, err
	}

	if httpResponse.StatusCode == 404 { // NOT_FOUND
		tflog.Error(ctx, "Template not found", map[string]interface{}{
			"method":      method,
			"url":         url,
			"status_code": httpResponse.StatusCode,
		})
		defer httpResponse.Body.Close()
		return nil, SpecificErrorCodes.NotFound
	}

	if !isResponseOk(httpResponse) {
		tflog.Error(ctx, "Template API returned error status", map[string]interface{}{
			"method":      method,
			"url":         url,
			"status_code": httpResponse.StatusCode,
		})
		defer httpResponse.Body.Close()
		return nil, getErrorMessage(httpResponse)
	}

	return httpResponse, nil
}

// This function does NOT close the httpResponse body.
func getTemplateResponseBody(httpResponse *http.Response) (TemplateResponse, error) {
	var templateResponse TemplateResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&templateResponse); err != nil {
		return TemplateResponse{}, err
	}

	return templateResponse, nil
}

func (biotSdkImpl biotSdkImpl) SearchTemplates(ctx context.Context, accessToken string, searchRequest map[string]interface{}) (SearchTemplatesResponse, error) {
	encodedSearchRequest, err := encodeSearchRequest(searchRequest)
	if err != nil {
		return SearchTemplatesResponse{}, err
	}

	var url = fmt.Sprintf("%s/%s/v1/templates?searchRequest=%s", biotSdkImpl.baseUrl, settingsPrefix, encodedSearchRequest)

	req, requestErr := http.NewRequest(http.MethodGet, url, nil)
	if requestErr != nil {
		return SearchTemplatesResponse{}, requestErr
	}

	req.Header.Set(authorizationHeaderKey, fmt.Sprintf("Bearer %s", accessToken))

	var httpResponse, responseErr = httpClient.Do(req)
	if responseErr != nil {
		return SearchTemplatesResponse{}, responseErr
	}
	defer httpResponse.Body.Close()

	if !isResponseOk(httpResponse) {
		return SearchTemplatesResponse{}, getErrorMessage(httpResponse)
	}

	// Parse response body into TemplateResponse
	var searchRemplatesResponse SearchTemplatesResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&searchRemplatesResponse); err != nil {
		return SearchTemplatesResponse{}, err
	}

	return searchRemplatesResponse, nil
}

func isResponseOk(response *http.Response) bool {
	return response.StatusCode >= 200 && response.StatusCode < 300
}

func getErrorMessage(response *http.Response) error {
	var biotErrorObject map[string]interface{}
	json.NewDecoder(response.Body).Decode(&biotErrorObject)

	var msg string
	var code string
	var traceId string

	if msg, _ = biotErrorObject["message"].(string); msg == "" {
		msg = "unknown error message"
	}

	if code, _ = biotErrorObject["code"].(string); code == "" {
		code = "unknown error code"
	}

	if traceId, _ = biotErrorObject["traceId"].(string); traceId == "" {
		traceId = "unknown trade-id"
	}

	return fmt.Errorf("server error (status: [%d], code: [%s], traceId: [%s]): [%s]", response.StatusCode, code, traceId, msg)
}

func encodeSearchRequest(searchRequest map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(searchRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search request: %w", err)
	}

	encoded := url.QueryEscape(string(jsonBytes))
	return encoded, nil
}

func (biotSdkImpl biotSdkImpl) ValidateVersions(ctx context.Context, accessToken string, terraformProviderVersion string, minimumBiotVersion string) (TerraformVersionValidationResponse, error) {
	baseURL := fmt.Sprintf("%s/%s/v1/terraform/versions/validate", biotSdkImpl.baseUrl, settingsPrefix)

	// Add query parameters
	params := url.Values{}
	params.Add("terraform-provider", terraformProviderVersion)
	params.Add("minimum-biot", minimumBiotVersion)
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return TerraformVersionValidationResponse{}, err
	}

	req.Header.Set(authorizationHeaderKey, fmt.Sprintf("Bearer %s", accessToken))

	httpResponse, err := httpClient.Do(req)
	if err != nil {
		return TerraformVersionValidationResponse{}, err
	}
	defer httpResponse.Body.Close()

	if !isResponseOk(httpResponse) {
		return TerraformVersionValidationResponse{}, getErrorMessage(httpResponse)
	}

	var validationResponse TerraformVersionValidationResponse
	if err := json.NewDecoder(httpResponse.Body).Decode(&validationResponse); err != nil {
		return TerraformVersionValidationResponse{}, err
	}

	return validationResponse, nil
}

func NewBiotSdkImpl(baseUrl string) *biotSdkImpl {
	return &biotSdkImpl{
		baseUrl: baseUrl,
	}
}
