package api

// User represents a global user account in Conclave.
// Corresponds to the UserSerializer / core.models.User in Django.
type User struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	DateJoined string `json:"date_joined,omitempty"`
}

// Tenant represents an isolated workspace organization.
// Corresponds to TenantSerializer / core.models.Tenant in Django.
type Tenant struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	IsPublic     bool   `json:"is_public"`
	CreationDate string `json:"creation_date,omitempty"`
	MemberCount  int    `json:"member_count,omitempty"`
}

// SignupRequest is the payload sent to POST /auth/signup/
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest is the payload sent to POST /auth/login/
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the payload returned by both POST /auth/signup/ and POST /auth/login/
type AuthResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenRequest is sent to POST /auth/token/refresh/
type RefreshTokenRequest struct {
	Refresh string `json:"refresh"`
}

// RefreshTokenResponse is returned by POST /auth/token/refresh/
type RefreshTokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh,omitempty"`
}

// MeResponse represents the full profile returned by GET /auth/me/
type MeResponse struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	DateJoined    string   `json:"date_joined"`
	TenantsJoined []Tenant `json:"tenants_joined"`
	TenantsOwned  []Tenant `json:"tenants_owned"`
}
