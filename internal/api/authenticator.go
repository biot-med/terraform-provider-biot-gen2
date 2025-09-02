package api

import (
	"context"
	"fmt"
)

// AuthenticatorService handles authentication and token management
type AuthenticatorService struct {
	biotSdk          BiotSdk
	serviceId        string
	serviceSecretKey string
}

// NewAuthenticatorService creates a new authenticator service
func NewAuthenticatorService(biotSdk BiotSdk, serviceId string, serviceSecretKey string) *AuthenticatorService {
	return &AuthenticatorService{
		biotSdk:          biotSdk,
		serviceId:        serviceId,
		serviceSecretKey: serviceSecretKey,
	}
}

// GetAccessToken retrieves a fresh access token from the Biot service
func (auth *AuthenticatorService) GetAccessToken(ctx context.Context) (string, error) {
	response, err := auth.biotSdk.LoginAsService(ctx, auth.serviceId, auth.serviceSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to login as service using service ID [%s]: %w", auth.serviceId, err)
	}

	return response.Token, nil
}
