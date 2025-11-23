package api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// tokenCache represents the structure of the cached token file
type tokenCache struct {
	Token      string    `json:"token"`
	Expiration time.Time `json:"expiration"`
}

// AuthenticatorService handles authentication and token management
type AuthenticatorService struct {
	biotSdk          BiotSdk
	serviceId        string
	serviceSecretKey string

	// Token caching fields
	cachedToken     string
	tokenExpiration time.Time
	tokenMutex      sync.RWMutex
	cacheFilePath   string
}

// NewAuthenticatorService creates a new authenticator service
func NewAuthenticatorService(biotSdk BiotSdk, serviceId string, serviceSecretKey string) *AuthenticatorService {
	auth := &AuthenticatorService{
		biotSdk:          biotSdk,
		serviceId:        serviceId,
		serviceSecretKey: serviceSecretKey,
	}

	// Set up cache file path based on service ID (hashed for security)
	auth.cacheFilePath = auth.getCacheFilePath()

	// Load cached token from disk if available
	auth.loadCachedToken()

	return auth
}

// getCacheFilePath returns the path to the cache file for this service
func (auth *AuthenticatorService) getCacheFilePath() string {
	// Create a hash of the service ID to use as filename (for security)
	hash := sha256.Sum256([]byte(auth.serviceId))
	hashStr := fmt.Sprintf("%x", hash)[:16] // Use first 16 chars

	// Use Terraform's plugin cache directory or fallback to temp dir
	cacheDir := os.Getenv("TF_PLUGIN_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = os.TempDir()
	}

	// Create a subdirectory for our cache
	cacheSubDir := filepath.Join(cacheDir, "terraform-provider-biot-gen2")
	os.MkdirAll(cacheSubDir, 0700) // Create directory with read/write permissions for owner only

	return filepath.Join(cacheSubDir, fmt.Sprintf("token_%s.json", hashStr))
}

// loadCachedToken loads a cached token from disk if it exists and is still valid
func (auth *AuthenticatorService) loadCachedToken() {
	auth.tokenMutex.Lock()
	defer auth.tokenMutex.Unlock()

	data, err := os.ReadFile(auth.cacheFilePath)
	if err != nil {
		// Cache file doesn't exist or can't be read - that's okay
		return
	}

	var cache tokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		// Invalid cache file - ignore it
		return
	}

	// Check if token is still valid (with a small buffer)
	if time.Now().Before(cache.Expiration) {
		auth.cachedToken = cache.Token
		auth.tokenExpiration = cache.Expiration
	}
}

// saveCachedToken saves the token to disk
func (auth *AuthenticatorService) saveCachedToken(token string, expiration time.Time) error {
	cache := tokenCache{
		Token:      token,
		Expiration: expiration,
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal token cache: %w", err)
	}

	// Write atomically using a temp file and rename
	tmpFile := auth.cacheFilePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token cache: %w", err)
	}

	// On Windows, os.Rename fails if the target file exists, so remove it first
	// This is safe because we're replacing it with the temp file
	if _, err := os.Stat(auth.cacheFilePath); err == nil {
		if err := os.Remove(auth.cacheFilePath); err != nil {
			os.Remove(tmpFile) // Clean up temp file on error
			return fmt.Errorf("failed to remove existing cache file: %w", err)
		}
	}

	if err := os.Rename(tmpFile, auth.cacheFilePath); err != nil {
		os.Remove(tmpFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename token cache file: %w", err)
	}

	return nil
}

// GetAccessToken retrieves an access token from the Biot service, reusing cached token if still valid
func (auth *AuthenticatorService) GetAccessToken(ctx context.Context) (string, error) {
	// Check if we have a valid cached token
	auth.tokenMutex.RLock()
	if auth.cachedToken != "" && time.Now().Before(auth.tokenExpiration) {
		token := auth.cachedToken
		auth.tokenMutex.RUnlock()
		return token, nil
	}
	auth.tokenMutex.RUnlock()

	// Need to fetch a new token
	auth.tokenMutex.Lock()
	defer auth.tokenMutex.Unlock()

	// Double-check after acquiring write lock (another goroutine might have updated it)
	if auth.cachedToken != "" && time.Now().Before(auth.tokenExpiration) {
		return auth.cachedToken, nil
	}

	// Fetch new token
	response, err := auth.biotSdk.LoginAsService(ctx, auth.serviceId, auth.serviceSecretKey)
	if err != nil {
		return "", fmt.Errorf("failed to login as service using service ID [%s]: %w", auth.serviceId, err)
	}

	// Parse expiration time
	expiration, err := time.Parse(time.RFC3339, response.Expiration)
	if err != nil {
		// If we can't parse expiration, assume it's valid for 5 minutes as a fallback
		expiration = time.Now().Add(5 * time.Minute)
	}

	// Cache the token with a small buffer (refresh 5 minutes before expiration)
	bufferTime := 5 * time.Minute
	auth.cachedToken = response.Token
	auth.tokenExpiration = expiration.Add(-bufferTime)

	// Save to disk for persistence across runs
	if err := auth.saveCachedToken(response.Token, auth.tokenExpiration); err != nil {
		// Log error but don't fail - in-memory cache still works
		// We could use tflog here, but to avoid circular dependencies, we'll just continue
	}

	return response.Token, nil
}
