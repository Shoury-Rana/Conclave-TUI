package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthLoginSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode login body: %v", err)
		}
		if req.Email != "alex@conclave.local" || req.Password != "password123" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password."})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AuthResponse{
			User: User{
				ID:    "user-uuid-999",
				Email: req.Email,
				Name:  "Alex Member",
			},
			AccessToken:  "mock-access-jwt",
			RefreshToken: "mock-refresh-jwt",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Login(context.Background(), "alex@conclave.local", "password123")
	if err != nil {
		t.Fatalf("expected login to succeed, got error: %v", err)
	}

	if resp.AccessToken != "mock-access-jwt" {
		t.Errorf("expected access token 'mock-access-jwt', got %s", resp.AccessToken)
	}
	if resp.User.Name != "Alex Member" {
		t.Errorf("expected user name 'Alex Member', got %s", resp.User.Name)
	}

	// Verify client tokens were cached
	access, refresh := client.GetTokens()
	if access != "mock-access-jwt" || refresh != "mock-refresh-jwt" {
		t.Errorf("client tokens not set properly: %s / %s", access, refresh)
	}
}

func TestAuthLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password."})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Login(context.Background(), "wrong@email.com", "wrongpwd")
	if err == nil {
		t.Fatalf("expected error on invalid credentials, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError type, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Invalid email or password." {
		t.Errorf("expected message 'Invalid email or password.', got %q", apiErr.Message)
	}
}

func TestAuthSignupSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/signup/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode signup body: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(AuthResponse{
			User: User{
				ID:    "new-user-uuid",
				Email: req.Email,
				Name:  req.Name,
			},
			AccessToken:  "signup-access-jwt",
			RefreshToken: "signup-refresh-jwt",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Signup(context.Background(), "Sam Test", "sam@conclave.local", "securepass123")
	if err != nil {
		t.Fatalf("expected signup to succeed, got error: %v", err)
	}
	if resp.User.Email != "sam@conclave.local" {
		t.Errorf("expected email sam@conclave.local, got %s", resp.User.Email)
	}
}

func TestGetMeWithAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer my-valid-token" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"detail": "Authentication credentials were not provided."})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(MeResponse{
			ID:    "my-uuid-1",
			Email: "alex@conclave.local",
			Name:  "Alex Member",
			TenantsJoined: []Tenant{
				{ID: "t-1", Name: "engineering", IsPublic: true},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetTokens("my-valid-token", "my-refresh-token")

	me, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("expected GetMe to succeed, got: %v", err)
	}
	if me.Name != "Alex Member" {
		t.Errorf("expected Alex Member, got %s", me.Name)
	}
	if len(me.TenantsJoined) != 1 || me.TenantsJoined[0].Name != "engineering" {
		t.Errorf("unexpected tenants joined: %+v", me.TenantsJoined)
	}
}
