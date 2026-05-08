package googlemail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

func TestEnsureTokenClearsCachedServiceWhenRefreshTokenChanges(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}

		if got := r.Form.Get("refresh_token"); got != "refresh-token-b" {
			t.Fatalf("refresh_token = %q, want %q", got, "refresh-token-b")
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access-token-b",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	}))
	defer tokenServer.Close()

	cachedService := &gmail.Service{}
	g := &googleService{
		config: &oauth2.Config{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				TokenURL: tokenServer.URL,
			},
		},
		token: &oauth2.Token{
			AccessToken:  "access-token-a",
			RefreshToken: "refresh-token-a",
			Expiry:       time.Now().Add(time.Hour),
		},
		activeRefreshToken: "refresh-token-a",
		service:            cachedService,
	}

	if err := g.ensureToken("refresh-token-b"); err != nil {
		t.Fatalf("ensureToken() error = %v", err)
	}

	if g.activeRefreshToken != "refresh-token-b" {
		t.Fatalf("activeRefreshToken = %q, want %q", g.activeRefreshToken, "refresh-token-b")
	}

	if g.token == nil || g.token.AccessToken != "access-token-b" {
		t.Fatalf("token access token = %v, want %q", g.token, "access-token-b")
	}

	if g.service != nil {
		t.Fatal("service cache not cleared after refresh token switch")
	}
}

func TestEnsureTokenKeepsCachedServiceForSameRefreshToken(t *testing.T) {
	cachedService := &gmail.Service{}
	g := &googleService{
		token: &oauth2.Token{
			AccessToken:  "access-token-a",
			RefreshToken: "refresh-token-a",
			Expiry:       time.Now().Add(time.Hour),
		},
		activeRefreshToken: "refresh-token-a",
		service:            cachedService,
	}

	if err := g.ensureToken("refresh-token-a"); err != nil {
		t.Fatalf("ensureToken() error = %v", err)
	}

	if g.service != cachedService {
		t.Fatal("service cache changed for same refresh token")
	}
}
