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
	Token      string `json:"token"`
	Expiration string `json:"expiration"`
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
	LoginWithCredentials(ctx context.Context, username string, password string) (LoginResponse, error)
	CreateTemplate(ctx context.Context, accessToken string, request CreateTemplateRequest) (TemplateResponse, error)
	UpdateTemplate(ctx context.Context, accessToken string, id string, request UpdateTemplateRequest) (TemplateResponse, error)
	GetTemplate(ctx context.Context, token string, id string) (TemplateResponse, error)
	DeleteTemplate(ctx context.Context, accessToken string, id string) error
	SearchTemplates(ctx context.Context, token string, searchrequest map[string]interface{}) (SearchTemplatesResponse, error)
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

// TODO: Change to login as service.
func (biotSdkImpl biotSdkImpl) LoginWithCredentials(ctx context.Context, username string, password string) (LoginResponse, error) {
	var url = fmt.Sprintf("%s/%s/v2/users/login", biotSdkImpl.baseUrl, umsPrefix)

	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	if err != nil {
		return LoginResponse{}, err
	}

	return login(ctx, url, requestBody)
}

func login(ctx context.Context, url string, requestBody []byte) (LoginResponse, error) {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("biotSdk: login - failed to create request with url: [%s]", url))
		return LoginResponse{}, err
	}

	request.Header.Set("Content-Type", "application/json")
	response, requestError := httpClient.Do(request)

	if requestError != nil {
		tflog.Warn(ctx, fmt.Sprintf("biotSdk: login - failed to call api with url: [%s]", url))
		return LoginResponse{}, requestError
	}

	if !isResponseOk(response) {
		tflog.Warn(ctx, fmt.Sprintf("biotSdk: login - api result is not 200 with url: [%s]", url))
		return LoginResponse{}, getErrorMessage(response)
	}

	defer response.Body.Close()

	var body []byte
	var loginResponse LoginResponse
	json.NewDecoder(response.Body).Decode(&loginResponse)
	json.Unmarshal(body, &loginResponse)

	return loginResponse, nil
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
		tflog.Error(ctx, fmt.Sprintf("biotSdk: [%s] Template - failed to create request with url: [%s]", method, url))
		return nil, requestErr
	}

	req.Header.Set(authorizationHeaderKey, fmt.Sprintf("Bearer %s", accessToken))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	var httpResponse, err = httpClient.Do(req)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("biotSdk: [%s] Template - failed to call API with url: [%s]", method, url))
		return nil, err
	}

	if httpResponse.StatusCode == 404 { // NOT_FOUND
		tflog.Error(ctx, fmt.Sprintf("biotSdk: [%s] Template - got 404 not found for url: [%s]", method, url))
		defer httpResponse.Body.Close()
		return nil, SpecificErrorCodes.NotFound
	}

	if !isResponseOk(httpResponse) {
		tflog.Error(ctx, fmt.Sprintf("biotSdk: [%s] Template  - api result with status code: [%d] with url: [%s]", method, httpResponse.StatusCode, url))
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

	if msg, _ = biotErrorObject["message"].(string); msg == "" {
		msg = "unknown error message"
	}

	if code, _ = biotErrorObject["code"].(string); code == "" {
		code = "unknown error code"
	}

	return fmt.Errorf("server error (status: [%d], code: [%s]): [%s]", response.StatusCode, code, msg)
}

func encodeSearchRequest(searchRequest map[string]interface{}) (string, error) {
	jsonBytes, err := json.Marshal(searchRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search request: %w", err)
	}

	encoded := url.QueryEscape(string(jsonBytes))
	return encoded, nil
}

func NewBiotSdkImpl(baseUrl string) *biotSdkImpl {
	return &biotSdkImpl{
		baseUrl: baseUrl,
	}
}
