package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTokenRevoked  = errors.New("refresh token revoked")
	ErrTokenExpired  = errors.New("refresh token expired")
	ErrTokenNotFound = errors.New("refresh token not found")
)

const refreshTokenBytes = 32
const refreshTokenExpiry = 30 * 24 * time.Hour // 30 days

// GenerateRefreshToken creates a cryptographically random token and stores
// its SHA-256 hash in the database. Returns the raw token to send to the client.
func GenerateRefreshToken(ctx context.Context, db *pgxpool.Pool, userID string, family string) (string, error) {
	raw := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	hash := hashToken(token)

	_, err := db.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, family, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		userID, hash, family, time.Now().Add(refreshTokenExpiry),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

// RotateRefreshToken validates the presented token, revokes it, and issues
// a new token in the same family. If the presented token is already revoked,
// the entire family is revoked (theft detection).
func RotateRefreshToken(ctx context.Context, db *pgxpool.Pool, rawToken string) (newToken string, userID string, err error) {
	hash := hashToken(rawToken)

	tx, err := db.Begin(ctx)
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback(ctx)

	var tokenID, family string
	var revoked bool
	var expiresAt time.Time
	err = tx.QueryRow(ctx,
		`SELECT id, user_id, family, revoked, expires_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		hash,
	).Scan(&tokenID, &userID, &family, &revoked, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", ErrTokenNotFound
		}
		return "", "", err
	}

	// Theft detection: if this token was already revoked, someone is replaying
	// a stolen token. Revoke the entire family.
	if revoked {
		tx.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE family = $1`, family)
		tx.Commit(ctx)
		return "", "", ErrTokenRevoked
	}

	if time.Now().After(expiresAt) {
		return "", "", ErrTokenExpired
	}

	// Revoke the current token.
	_, err = tx.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE id = $1`, tokenID)
	if err != nil {
		return "", "", err
	}

	// Issue new token in the same family.
	raw := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	newToken = hex.EncodeToString(raw)
	newHash := hashToken(newToken)

	_, err = tx.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, family, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		userID, newHash, family, time.Now().Add(refreshTokenExpiry),
	)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", err
	}

	return newToken, userID, nil
}

// RevokeFamily revokes all refresh tokens in a family (used on logout).
func RevokeFamily(ctx context.Context, db *pgxpool.Pool, rawToken string) error {
	hash := hashToken(rawToken)

	var family string
	err := db.QueryRow(ctx,
		`SELECT family FROM refresh_tokens WHERE token_hash = $1`, hash,
	).Scan(&family)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // already gone, nothing to revoke
		}
		return err
	}

	_, err = db.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE family = $1`, family)
	return err
}

// RevokeAllForUser revokes every refresh token for a user (password change, account compromise).
func RevokeAllForUser(ctx context.Context, db *pgxpool.Pool, userID string) error {
	_, err := db.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID)
	return err
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
