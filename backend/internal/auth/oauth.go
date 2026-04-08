package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OAuthUserInfo is the verified identity from an OAuth provider.
type OAuthUserInfo struct {
	Provider    string // "google", "apple"
	ProviderUID string // provider's unique user ID
	Email       string
	Name        string
}

// VerifyGoogleIDToken verifies a Google ID token using Google's tokeninfo endpoint.
func VerifyGoogleIDToken(ctx context.Context, idToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://oauth2.googleapis.com/tokeninfo?id_token="+idToken, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google tokeninfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google tokeninfo returned %d", resp.StatusCode)
	}

	var info struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google tokeninfo: %w", err)
	}

	if info.Sub == "" || info.Email == "" {
		return nil, fmt.Errorf("google token missing sub or email")
	}

	return &OAuthUserInfo{
		Provider:    "google",
		ProviderUID: info.Sub,
		Email:       info.Email,
		Name:        info.Name,
	}, nil
}
