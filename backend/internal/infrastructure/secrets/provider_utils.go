package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// httpClient abstracts the Do method so we can mock HTTP interactions in tests.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// accessToken represents a bearer token with expiry metadata.
type accessToken struct {
	Value  string
	Expiry time.Time
}

// tokenProvider fetches access tokens for a given scope.
type tokenProvider interface {
	Token(ctx context.Context, scope string) (accessToken, error)
}

// defaultHTTPClient provides a safe default HTTP client with sane timeouts.
var defaultHTTPClient = func() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

var jsonMarshal = json.Marshal

// readResponseBody safely reads response bodies for error reporting.
func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// tokenCache caches a single accessToken for a cloud secrets backend, refreshing
// it via the supplied provider once the cached token has less than one minute of
// validity remaining. It is safe for concurrent use.
type tokenCache struct {
	mu    sync.Mutex
	token accessToken
}

// get returns a cached accessToken when it still has at least one minute of
// validity remaining, otherwise fetches a new one via provider. providerName is
// used as a prefix when wrapping provider errors (e.g. "Azure", "Google").
func (c *tokenCache) get(ctx context.Context, provider tokenProvider, providerName, scope string) (accessToken, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token.Value != "" && time.Until(c.token.Expiry) > time.Minute {
		return c.token, nil
	}

	token, err := provider.Token(ctx, scope)
	if err != nil {
		return accessToken{}, fmt.Errorf("failed to fetch %s access token: %w", providerName, err)
	}
	c.token = token
	return token, nil
}

// doCloudRequest executes an authenticated HTTP request against a cloud secrets
// backend (Azure Key Vault, Google Secret Manager, etc.), returning the response
// body on a 2xx status. The request body is sent as application/json when non-nil.
// 404 responses are surfaced as "not found: <body>" so callers can detect them
// with isNotFoundError. providerName is used in error messages (e.g. "azure",
// "gcp").
func doCloudRequest(
	ctx context.Context,
	client httpClient,
	cache *tokenCache,
	provider tokenProvider,
	scope string,
	providerName string,
	method string,
	endpoint string,
	body io.Reader,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	token, err := cache.get(ctx, provider, providerName, scope)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.Value)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", providerName, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		respBody, _ := readResponseBody(resp)
		return nil, fmt.Errorf("not found: %s", string(respBody))
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, readErr := readResponseBody(resp)
		if readErr != nil {
			return nil, fmt.Errorf("%s request failed with status %d and unreadable body: %w", providerName, resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("%s request failed with status %d: %s", providerName, resp.StatusCode, string(respBody))
	}

	return readResponseBody(resp)
}

// decodeOAuthTokenResponse parses the standard OAuth2 access_token response
// shape and returns a freshly-expired accessToken. providerName is used as a
// prefix in error messages (e.g. "Azure token", "service account token").
func decodeOAuthTokenResponse(body []byte, providerName string) (accessToken, error) {
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return accessToken{}, fmt.Errorf("failed to decode %s response: %w", providerName, err)
	}
	if tokenResp.AccessToken == "" {
		return accessToken{}, fmt.Errorf("%s response missing access_token", providerName)
	}
	return accessToken{
		Value:  tokenResp.AccessToken,
		Expiry: time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}
