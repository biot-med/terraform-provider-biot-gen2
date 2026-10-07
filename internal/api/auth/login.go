// Package auth logs in as a BioT service user and hands out access tokens. It satisfies
// transport.TokenSource, so the authenticated API clients never deal with tokens directly.
package auth

import (
	"context"
	"fmt"
	"net/http"

	"biot.com/terraform-provider-biot-gen2/internal/api/transport"
)

type Jwt struct {
	Token      string `json:"accessToken"`
	Expiration string `json:"accessTokenExpiration"`
}

// Login exchanges service credentials for an access token. It goes out unauthenticated, for
// the obvious reason.
func Login(ctx context.Context, client *transport.Client, serviceID string, serviceSecretKey string) (Jwt, error) {
	url := fmt.Sprintf("%s/ums/v2/services/accessToken", client.BaseURL)

	body := map[string]string{
		"id":        serviceID,
		"secretKey": serviceSecretKey,
	}

	return transport.DoUnauthenticated[Jwt](ctx, client, http.MethodPost, url, body)
}
