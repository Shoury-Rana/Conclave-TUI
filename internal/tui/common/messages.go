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
