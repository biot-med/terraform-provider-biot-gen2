// Package api wires the BioT API packages together. Each domain lives in its own package -
// template, abac - over the shared plumbing in transport and auth.
package api

import (
	"biot.com/terraform-provider-biot-gen2/internal/api/abac"
	"biot.com/terraform-provider-biot-gen2/internal/api/auth"
	"biot.com/terraform-provider-biot-gen2/internal/api/template"
	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

// APIClient is what the provider hands to every resource. Reach the domain you need through
// its field, e.g. client.Template.Get(...) or client.Abac.CreateCondition(...).
type APIClient struct {
	Template *template.Client
	Abac     *abac.Client
	Versions *VersionValidator
}

func New(baseURL string, serviceID string, serviceSecretKey string) *APIClient {
	// The authenticator logs in over an unauthenticated client; everything else goes out on
	// a copy of it that attaches a bearer token to every request.
	anonymous := transport.New(baseURL)
	authenticator := auth.New(anonymous, serviceID, serviceSecretKey)
	authenticated := anonymous.WithTokens(authenticator)

	return &APIClient{
		Template: template.NewClient(authenticated),
		Abac:     abac.NewClient(authenticated),
		Versions: newVersionValidator(authenticated),
	}
}
