package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"biot.com/terraform-provider-biot/internal/version"
)

// VersionValidator handles version validation for the Terraform provider
type VersionValidator struct {
	biotSdk         BiotSdk
	authenticator   *AuthenticatorService
	providerVersion string
}

// NewVersionValidator creates a new version validator
func NewVersionValidator(biotSdk BiotSdk, authenticator *AuthenticatorService, providerVersion string) *VersionValidator {
	return &VersionValidator{
		biotSdk:         biotSdk,
		authenticator:   authenticator,
		providerVersion: providerVersion,
	}
}

// getMinimumBiotVersion returns the minimum Biot version required by this provider
func (v *VersionValidator) getMinimumBiotVersion() string {
	return version.MinimumBiotVersion
}

// ValidateVersions validates that the provider version is compatible with the Biot version
func (v *VersionValidator) ValidateVersions(ctx context.Context) (*TerraformVersionValidationResponse, error) {
	// Get access token
	token, err := v.authenticator.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token for version validation: %w", err)
	}

	// Call the version validation endpoint
	response, err := v.biotSdk.ValidateVersions(ctx, token, v.providerVersion, v.getMinimumBiotVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to validate versions: %w", err)
	}

	return &response, nil
}

// Validate performs complete version validation including compatibility check and logging
func (v *VersionValidator) Validate(ctx context.Context) error {
	log.Printf("[DEBUG] VersionValidator - Starting version validation for provider version: %s", v.providerVersion)

	validationResponse, err := v.ValidateVersions(ctx)
	if err != nil {
		log.Printf("[ERROR] VersionValidator - Version validation failed: %v", err)
		return fmt.Errorf("version validation failed: %w", err)
	}

	if validationResponse.Status != StatusSupported {
		log.Printf("[ERROR] VersionValidator - Incompatible versions. Provider: %s, Biot: %s, Status: %s",
			v.providerVersion, validationResponse.BiotVersion.Version, validationResponse.Status)

		// Convert the full response to JSON for detailed error information
		jsonResponse, _ := json.MarshalIndent(validationResponse, "", "  ")
		return fmt.Errorf("versions are not compatible. %s", string(jsonResponse))
	}

	log.Printf("[INFO] VersionValidator - Version validation successful. Provider: %s, Biot: %s",
		v.providerVersion, validationResponse.BiotVersion.Version)
	return nil
}

// IsVersionSupported checks if the current provider version is supported by the Biot service
func (v *VersionValidator) IsVersionSupported(ctx context.Context) (bool, error) {
	response, err := v.ValidateVersions(ctx)
	if err != nil {
		return false, err
	}

	return response.Status == StatusSupported, nil
}
