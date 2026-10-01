package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

// VersionValidator checks that this provider is compatible with the BioT deployment. The
// endpoint lives on the settings service but the concern is provider-wide, so it sits here
// rather than in the template package.
type VersionValidator struct {
	http *transport.Client
}

func newVersionValidator(httpClient *transport.Client) *VersionValidator {
	return &VersionValidator{http: httpClient}
}

// ValidationUnsupportedError is returned when the API returned 200 OK but status==UNSUPPORTED
type ValidationUnsupportedError struct {
	Response TerraformVersionValidationResponse
}

func (e ValidationUnsupportedError) Error() string {
	b, _ := json.MarshalIndent(e.Response, "", "  ")
	return fmt.Sprintf("versions are not compatible. %s", string(b))
}

// ValidationAPIError is returned when the call to the API failed (non-2xx or transport error)
type ValidationAPIError struct {
	Err error
}

func (e ValidationAPIError) Error() string {
	return fmt.Sprintf("version validation API error: %v", e.Err)
}
func (e ValidationAPIError) Unwrap() error { return e.Err }

// ValidateVersions validates that the provider version is compatible with the Biot version
func (v *VersionValidator) ValidateVersions(ctx context.Context, providerVersion string, minimumBiotVersion string) (*TerraformVersionValidationResponse, error) {
	params := url.Values{}
	params.Add("terraform-provider", providerVersion)
	params.Add("minimum-biot", minimumBiotVersion)

	requestURL := fmt.Sprintf("%s/settings/v1/terraform/versions/validate?%s", v.http.BaseURL, params.Encode())

	response, err := transport.Do[TerraformVersionValidationResponse](ctx, v.http, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// Validate performs complete version validation including compatibility check and logging
func (v *VersionValidator) Validate(ctx context.Context, providerVersion string, minimumBiotVersion string) error {
	tflog.Debug(ctx, "Starting version validation", map[string]interface{}{
		"provider_version": providerVersion,
		"minimum_biot":     minimumBiotVersion,
	})

	validationResponse, err := v.ValidateVersions(ctx, providerVersion, minimumBiotVersion)
	if err != nil {
		tflog.Error(ctx, "Version validation API error", map[string]interface{}{"error": err})
		return ValidationAPIError{Err: err}
	}

	b, _ := json.MarshalIndent(validationResponse, "", "  ")

	if validationResponse.Status != StatusSupported {
		tflog.Error(ctx, "Incompatible versions detected", map[string]interface{}{"response": string(b)})
		return ValidationUnsupportedError{Response: *validationResponse}
	}

	tflog.Info(ctx, "Version validation successful", map[string]interface{}{"response": string(b)})
	return nil
}

// IsVersionSupported checks if the current provider version is supported by the Biot service
func (v *VersionValidator) IsVersionSupported(ctx context.Context, providerVersion string, minimumBiotVersion string) (bool, error) {
	response, err := v.ValidateVersions(ctx, providerVersion, minimumBiotVersion)
	if err != nil {
		return false, err
	}

	return response.Status == StatusSupported, nil
}
