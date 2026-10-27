package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMyTenants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/mine/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]Tenant{
			{ID: "t-1", Name: "engineering", IsPublic: true, MemberCount: 12},
			{ID: "t-2", Name: "design", IsPublic: false, MemberCount: 4},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetTokens("test-token", "")

	tenants, err := client.GetMyTenants(context.Background())
	if err != nil {
		t.Fatalf("expected GetMyTenants to succeed, got: %v", err)
	}

	if len(tenants) != 2 {
		t.Fatalf("expected 2 tenants, got %d", len(tenants))
	}
	if tenants[0].Name != "engineering" || tenants[1].Name != "design" {
		t.Errorf("unexpected tenant names: %+v", tenants)
	}
}

func TestSearchTenants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/search/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		q := r.URL.Query().Get("q")
		if q != "conclave" {
			t.Errorf("expected query 'conclave', got %q", q)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]Tenant{
			{ID: "t-3", Name: "conclave-devs", IsPublic: true, MemberCount: 42},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetTokens("test-token", "")

	results, err := client.SearchTenants(context.Background(), "conclave")
	if err != nil {
		t.Fatalf("expected SearchTenants to succeed, got: %v", err)
	}
	if len(results) != 1 || results[0].Name != "conclave-devs" {
		t.Errorf("unexpected search results: %+v", results)
	}
}

func TestCreateTenant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/" || r.Method != http.MethodPost {
			t.Errorf("unexpected path or method: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req CreateTenantRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "devops" || !req.IsPublic {
			t.Errorf("unexpected payload: %+v", req)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Tenant{
			ID:       "t-new",
			Name:     req.Name,
			IsPublic: req.IsPublic,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetTokens("test-token", "")

	tenant, err := client.CreateTenant(context.Background(), "devops", true)
	if err != nil {
		t.Fatalf("expected CreateTenant to succeed, got: %v", err)
	}
	if tenant.Name != "devops" {
		t.Errorf("expected devops, got %s", tenant.Name)
	}
}

func TestJoinTenant(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tenant/join/" || r.Method != http.MethodPost {
			t.Errorf("unexpected path or method: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var req JoinTenantRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.TenantName != "acme" {
			t.Errorf("unexpected join target: %s", req.TenantName)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(JoinTenantResponse{
			Message:   "Successfully joined acme.",
			ProfileID: 101,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetTokens("test-token", "")

	resp, err := client.JoinTenant(context.Background(), "acme")
	if err != nil {
		t.Fatalf("expected JoinTenant to succeed, got: %v", err)
	}
	if resp.ProfileID != 101 {
		t.Errorf("expected profile ID 101, got %d", resp.ProfileID)
	}
}
