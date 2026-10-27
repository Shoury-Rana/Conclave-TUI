package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client is the primary HTTP client for communicating with the Conclave backend.
// It manages base URLs, bearer token injection, and thread-safe token updates.
type Client struct {
	baseURL    string
	httpClient *http.Client

	mu           sync.RWMutex
	accessToken  string
	refreshToken string
}

// NewClient initializes a new Conclave API client with reasonable network timeouts.
func NewClient(baseURL string) *Client {
	// Trim trailing slash to ensure clean URL concatenation.
	baseURL = strings.TrimRight(baseURL, "/")

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SetTokens updates the cached JWT access and refresh tokens in a thread-safe manner.
func (c *Client) SetTokens(access, refresh string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = access
	if refresh != "" {
		c.refreshToken = refresh
	}
}

// GetTokens returns the currently stored access and refresh tokens.
func (c *Client) GetTokens() (access, refresh string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.accessToken, c.refreshToken
}

// BaseURL returns the configured backend root URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// do performs an HTTP request against the root domain.
func (c *Client) do(ctx context.Context, method, path string, bodyData any, result any, requiresAuth bool) error {
	return c.doRequest(ctx, "", method, path, bodyData, result, requiresAuth)
}

// doTenant performs an HTTP request targeted at a specific workspace subdomain
// (e.g. acme.lvh.me:8000) using the Host header.
func (c *Client) doTenant(ctx context.Context, tenantName, method, path string, bodyData any, result any, requiresAuth bool) error {
	return c.doRequest(ctx, tenantName, method, path, bodyData, result, requiresAuth)
}

// doRequest performs an HTTP request against either root domain or a workspace subdomain.
func (c *Client) doRequest(ctx context.Context, tenantName, method, path string, bodyData any, result any, requiresAuth bool) error {
	reqURL := fmt.Sprintf("%s%s", c.baseURL, path)

	var reqBody io.Reader
	if bodyData != nil {
		jsonBytes, err := json.Marshal(bodyData)
		if err != nil {
			return fmt.Errorf("failed to encode request JSON: %w", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if bodyData != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Subdomain routing: If a tenant is specified, set the Host header so Django's SubDomainMiddleware
	// resolves the workspace context authoritatively (e.g. acme.lvh.me:8000).
	if tenantName != "" {
		host := req.URL.Hostname()
		port := req.URL.Port()
		if host == "127.0.0.1" || host == "localhost" {
			if port != "" {
				req.Host = fmt.Sprintf("%s.lvh.me:%s", tenantName, port)
			} else {
				req.Host = fmt.Sprintf("%s.lvh.me", tenantName)
			}
		} else {
			if port != "" {
				req.Host = fmt.Sprintf("%s.%s:%s", tenantName, host, port)
			} else {
				req.Host = fmt.Sprintf("%s.%s", tenantName, host)
			}
		}
	}

	if requiresAuth {
		c.mu.RLock()
		token := c.accessToken
		c.mu.RUnlock()

		if token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("network request failed (%s %s): %w", method, path, err)
	}
	defer resp.Body.Close()

	// If 401 Unauthorized and we have a refresh token (and this is not the refresh endpoint itself)
	if resp.StatusCode == http.StatusUnauthorized && requiresAuth && path != "/auth/token/refresh/" {
		c.mu.RLock()
		hasRefresh := c.refreshToken != ""
		c.mu.RUnlock()

		if hasRefresh {
			refreshResp, err := c.RefreshToken(ctx)
			if err == nil && refreshResp != nil && refreshResp.Access != "" {
				// Retry the original request once with new token
				return c.doRequest(ctx, tenantName, method, path, bodyData, result, requiresAuth)
			}
		}
	}

	// If response is not in the 2xx success range, parse and return the error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseErrorResponse(resp)
	}

	// If the caller requested decoding the response body into a struct
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response JSON: %w", err)
		}
	}

	return nil
}
