package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GetMyTenants fetches all workspaces the authenticated user belongs to or owns.
// Calls GET /tenant/mine/.
func (c *Client) GetMyTenants(ctx context.Context) ([]Tenant, error) {
	var tenants []Tenant
	err := c.do(ctx, http.MethodGet, "/tenant/mine/", nil, &tenants, true)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

// SearchTenants searches public workspaces by name.
// Calls GET /tenant/search/?q=<query>.
func (c *Client) SearchTenants(ctx context.Context, query string) ([]Tenant, error) {
	path := "/tenant/search/"
	query = strings.TrimSpace(query)
	if query != "" {
		path = fmt.Sprintf("/tenant/search/?q=%s", url.QueryEscape(query))
	}

	var results []Tenant
	err := c.do(ctx, http.MethodGet, path, nil, &results, true)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// CreateTenant provisions a new workspace organization.
// Calls POST /tenant/.
func (c *Client) CreateTenant(ctx context.Context, name string, isPublic bool) (*Tenant, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return nil, errors.New("workspace name cannot be empty")
	}

	payload := CreateTenantRequest{
		Name:     name,
		IsPublic: isPublic,
	}

	var tenant Tenant
	err := c.do(ctx, http.MethodPost, "/tenant/", payload, &tenant, true)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// JoinTenant requests membership into a workspace.
// Calls POST /tenant/join/.
// If the workspace is public, membership is created immediately.
// If the workspace is private, a pending join request is submitted to the owner.
func (c *Client) JoinTenant(ctx context.Context, tenantName string) (*JoinTenantResponse, error) {
	tenantName = strings.TrimSpace(strings.ToLower(tenantName))
	if tenantName == "" {
		return nil, errors.New("workspace name is required")
	}

	payload := JoinTenantRequest{
		TenantName: tenantName,
	}

	var resp JoinTenantResponse
	err := c.do(ctx, http.MethodPost, "/tenant/join/", payload, &resp, true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMyInvitations retrieves pending workspace invitations sent to the user.
// Calls GET /tenant/invitations/.
func (c *Client) GetMyInvitations(ctx context.Context) ([]TenantInvitation, error) {
	var invitations []TenantInvitation
	err := c.do(ctx, http.MethodGet, "/tenant/invitations/", nil, &invitations, true)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// RespondInvitation accepts or declines a pending workspace invitation.
// Calls POST /tenant/invitations/<id>/respond/.
func (c *Client) RespondInvitation(ctx context.Context, invitationID string, accept bool) error {
	action := "accept"
	if !accept {
		action = "decline"
	}

	path := fmt.Sprintf("/tenant/invitations/%s/respond/", invitationID)
	payload := RespondInvitationRequest{Action: action}

	return c.do(ctx, http.MethodPost, path, payload, nil, true)
}

// GetTenantDetail retrieves the full workspace overview and member stats.
// Subdomain-scoped call: GET http://<tenant_name>.<host>/
func (c *Client) GetTenantDetail(ctx context.Context, tenantName string) (*TenantDetail, error) {
	var detail TenantDetail
	err := c.doTenant(ctx, tenantName, http.MethodGet, "/", nil, &detail, true)
	if err != nil {
		return nil, err
	}
	return &detail, nil
}

// GetTenantMembers fetches the workspace member roster with realtime online status.
// Subdomain-scoped call: GET http://<tenant_name>.<host>/members/
func (c *Client) GetTenantMembers(ctx context.Context, tenantName string) ([]Member, error) {
	var members []Member
	err := c.doTenant(ctx, tenantName, http.MethodGet, "/members/", nil, &members, true)
	if err != nil {
		return nil, err
	}
	return members, nil
}
