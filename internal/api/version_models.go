package api

// TerraformVersionValidationResponse represents the response from the version validation endpoint
type TerraformVersionValidationResponse struct {
	Status          TerraformVersionValidationStatusEnum `json:"status"`
	ProviderVersion VersionInfo                          `json:"providerVersion"`
	BiotVersion     VersionInfo                          `json:"biotVersion"`
}

// TerraformVersionValidationStatusEnum represents the validation status
type TerraformVersionValidationStatusEnum string

const (
	StatusSupported   TerraformVersionValidationStatusEnum = "SUPPORTED"
	StatusUnsupported TerraformVersionValidationStatusEnum = "UNSUPPORTED"
)

// VersionInfo represents version information
type VersionInfo struct {
	Version     string `json:"version"`
	MinRequired string `json:"minRequired"`
}
