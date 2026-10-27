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

// CreateTenantRequest is the payload sent to POST /tenant/
type CreateTenantRequest struct {
	Name     string `json:"name"`
	IsPublic bool   `json:"is_public"`
}

// JoinTenantRequest is the payload sent to POST /tenant/join/
type JoinTenantRequest struct {
	TenantName string `json:"tenant_name,omitempty"`
	TenantID   string `json:"tenant_id,omitempty"`
	Username   string `json:"username,omitempty"`
}

// JoinTenantResponse is returned by POST /tenant/join/
type JoinTenantResponse struct {
	Message   string `json:"message"`
	ProfileID int    `json:"profile_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// TenantInvitation represents an invitation to join a workspace.
// Returned by GET /tenant/invitations/.
type TenantInvitation struct {
	ID               string `json:"id"`
	Tenant           string `json:"tenant"`
	TenantName       string `json:"tenant_name"`
	InvitedUser      string `json:"invited_user"`
	InvitedUserEmail string `json:"invited_user_email"`
	InvitedByName    string `json:"invited_by_name"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
}

// RespondInvitationRequest is sent to POST /tenant/invitations/<id>/respond/
type RespondInvitationRequest struct {
	Action string `json:"action"` // "accept" or "decline"
}

// Member represents a member of a workspace.
// Returned by GET http://<tenant>.<domain>/members/
type Member struct {
	ID       int      `json:"id"`
	UserID   string   `json:"user_id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Username string   `json:"username"`
	Role     string   `json:"role"`
	Tags     []string `json:"tags"`
	JoinedAt string   `json:"joined_at"`
	IsOnline bool     `json:"is_online"`
}

// TenantDetail represents the full details of a workspace.
// Returned by GET http://<tenant>.<domain>/
type TenantDetail struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	IsPublic     bool     `json:"is_public"`
	Owner        string   `json:"owner"`
	CreationDate string   `json:"creation_date"`
	MemberCount  int      `json:"member_count"`
	RoomCount    int      `json:"room_count"`
	Members      []Member `json:"members"`
}

