package common

import (
	"Conclave-TUI/internal/api"
)

// LoginSuccessMsg is emitted when a user successfully logs in.
type LoginSuccessMsg struct {
	Response *api.AuthResponse
}

// SignupSuccessMsg is emitted when a user successfully signs up.
type SignupSuccessMsg struct {
	Response *api.AuthResponse
}

// AuthErrorMsg is emitted when login or signup fails.
type AuthErrorMsg struct {
	Err error
}

// TokenValidatedMsg is emitted on startup when saved tokens are successfully verified via GET /auth/me/.
type TokenValidatedMsg struct {
	Me *api.MeResponse
}

// TokenInvalidMsg is emitted when saved tokens are expired or rejected.
type TokenInvalidMsg struct {
	Err error
}

// LogoutMsg is emitted when the user requests to log out and clear saved credentials.
type LogoutMsg struct{}

// WorkspaceListLoadedMsg is emitted when GET /tenant/mine/ completes.
type WorkspaceListLoadedMsg struct {
	Tenants []api.Tenant
}

// SearchResultsLoadedMsg is emitted when GET /tenant/search/?q= completes.
type SearchResultsLoadedMsg struct {
	Query   string
	Results []api.Tenant
}

// WorkspaceCreatedMsg is emitted when POST /tenant/ succeeds.
type WorkspaceCreatedMsg struct {
	Tenant *api.Tenant
}

// WorkspaceJoinedMsg is emitted when POST /tenant/join/ succeeds.
type WorkspaceJoinedMsg struct {
	Response   *api.JoinTenantResponse
	TenantName string
}

// InvitationsLoadedMsg is emitted when GET /tenant/invitations/ completes.
type InvitationsLoadedMsg struct {
	Invitations []api.TenantInvitation
}

// InvitationRespondedMsg is emitted when responding to an invitation completes.
type InvitationRespondedMsg struct {
	InvitationID string
	Action       string
}

// WorkspaceSelectedMsg is emitted when the user selects a workspace to enter.
type WorkspaceSelectedMsg struct {
	Tenant api.Tenant
}

// WorkspaceDetailLoadedMsg is emitted when GET / on a tenant subdomain finishes.
type WorkspaceDetailLoadedMsg struct {
	Detail *api.TenantDetail
}

// WorkspaceErrorMsg is emitted when an error occurs during workspace operations.
type WorkspaceErrorMsg struct {
	Err error
}

