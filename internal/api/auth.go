package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// Login authenticates an existing user with their email and password.
// Calls POST /auth/login/ on the root domain.
func (c *Client) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return nil, errors.New("email and password cannot be empty")
	}

	payload := LoginRequest{
		Email:    email,
		Password: password,
	}

	var authResp AuthResponse
	err := c.do(ctx, http.MethodPost, "/auth/login/", payload, &authResp, false)
	if err != nil {
		return nil, err
	}

	// Update stored tokens in the client
	c.SetTokens(authResp.AccessToken, authResp.RefreshToken)
	return &authResp, nil
}

// Signup registers a new user with full name, email, and password.
// Calls POST /auth/signup/ on the root domain.
func (c *Client) Signup(ctx context.Context, name, email, password string) (*AuthResponse, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	if name == "" {
		return nil, errors.New("name is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	payload := SignupRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}

	var authResp AuthResponse
	err := c.do(ctx, http.MethodPost, "/auth/signup/", payload, &authResp, false)
	if err != nil {
		return nil, err
	}

	// Update stored tokens in the client
	c.SetTokens(authResp.AccessToken, authResp.RefreshToken)
	return &authResp, nil
}

// RefreshToken exchanges an existing refresh token for a fresh access token.
// Calls POST /auth/token/refresh/.
func (c *Client) RefreshToken(ctx context.Context) (*RefreshTokenResponse, error) {
	c.mu.RLock()
	refresh := c.refreshToken
	c.mu.RUnlock()

	if refresh == "" {
		return nil, errors.New("no refresh token available")
	}

	payload := RefreshTokenRequest{
		Refresh: refresh,
	}

	var refreshResp RefreshTokenResponse
	err := c.do(ctx, http.MethodPost, "/auth/token/refresh/", payload, &refreshResp, false)
	if err != nil {
		return nil, err
	}

	// When ROTATE_REFRESH_TOKENS is enabled in Django settings,
	// a new refresh token is also issued alongside the new access token.
	c.SetTokens(refreshResp.Access, refreshResp.Refresh)
	return &refreshResp, nil
}

// GetMe retrieves the authenticated user's global profile and workspace list.
// Calls GET /auth/me/ with the Bearer token.
func (c *Client) GetMe(ctx context.Context) (*MeResponse, error) {
	var me MeResponse
	err := c.do(ctx, http.MethodGet, "/auth/me/", nil, &me, true)
	if err != nil {
		return nil, err
	}
	return &me, nil
}
