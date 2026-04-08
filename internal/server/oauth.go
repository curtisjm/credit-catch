package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"creditcatch/backend/internal/auth"
)

type oauthRequest struct {
	Provider string `json:"provider"`
	IDToken  string `json:"id_token"`
}

// handleOAuth verifies an OAuth provider's ID token and either logs in an
// existing user or creates a new one. Links the provider account in oauth_accounts.
func (s *Server) handleOAuth(w http.ResponseWriter, r *http.Request) {
	var req oauthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	// Verify the ID token with the provider.
	var info *auth.OAuthUserInfo
	var err error

	switch req.Provider {
	case "google":
		info, err = auth.VerifyGoogleIDToken(r.Context(), req.IDToken)
	// case "apple":
	//     info, err = auth.VerifyAppleIDToken(r.Context(), req.IDToken)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unsupported provider: %s", req.Provider)})
		return
	}

	if err != nil {
		s.logger.Error("oauth: verify id token", "provider", req.Provider, "error", err)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid id token"})
		return
	}

	ctx := r.Context()

	// Check if this OAuth account is already linked.
	var userID, email, name string
	err = s.db.QueryRow(ctx,
		`SELECT u.id, u.email, u.name
		 FROM oauth_accounts oa
		 JOIN users u ON u.id = oa.user_id
		 WHERE oa.provider = $1 AND oa.provider_uid = $2`,
		info.Provider, info.ProviderUID,
	).Scan(&userID, &email, &name)

	if err == nil {
		// Existing OAuth link — issue tokens and done.
		s.issueTokenPair(w, r, userID, email, name, http.StatusOK)
		return
	}

	// No existing link. Check if a user with this email exists.
	err = s.db.QueryRow(ctx,
		`SELECT id, email, name FROM users WHERE email = $1`, info.Email,
	).Scan(&userID, &email, &name)

	if err != nil {
		// No existing user — create one (no password, OAuth-only).
		err = s.db.QueryRow(ctx,
			`INSERT INTO users (email, name) VALUES ($1, $2)
			 RETURNING id, email, name`,
			info.Email, info.Name,
		).Scan(&userID, &email, &name)
		if err != nil {
			s.logger.Error("oauth: create user", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
			return
		}
	}

	// Link the OAuth account to the user.
	_, err = s.db.Exec(ctx,
		`INSERT INTO oauth_accounts (user_id, provider, provider_uid, email)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, provider_uid) DO NOTHING`,
		userID, info.Provider, info.ProviderUID, info.Email,
	)
	if err != nil {
		s.logger.Error("oauth: link account", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	s.issueTokenPair(w, r, userID, email, name, http.StatusOK)
}
