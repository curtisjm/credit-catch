package auth

import (
	"testing"
	"time"
)

func TestJWTIssueAndValidate(t *testing.T) {
	issuer := NewJWTIssuer("test-secret-key-1234", 1*time.Hour)

	token, err := issuer.Issue("user-abc-123")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}
	if token == "" {
		t.Fatal("Issue() returned empty token")
	}

	claims, err := issuer.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}
	if claims.UserID != "user-abc-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-abc-123")
	}
}

func TestJWTExpiredToken(t *testing.T) {
	// Issue with a negative expiry so the token is immediately expired.
	issuer := NewJWTIssuer("test-secret", -1*time.Second)

	token, err := issuer.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}

	_, err = issuer.Validate(token)
	if err == nil {
		t.Fatal("Validate() should fail for expired token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	issuer1 := NewJWTIssuer("secret-one", 1*time.Hour)
	issuer2 := NewJWTIssuer("secret-two", 1*time.Hour)

	token, err := issuer1.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}

	_, err = issuer2.Validate(token)
	if err == nil {
		t.Fatal("Validate() should fail with wrong secret")
	}
}

func TestJWTInvalidToken(t *testing.T) {
	issuer := NewJWTIssuer("test-secret", 1*time.Hour)

	_, err := issuer.Validate("not-a-real-token")
	if err == nil {
		t.Fatal("Validate() should fail for garbage input")
	}

	_, err = issuer.Validate("")
	if err == nil {
		t.Fatal("Validate() should fail for empty string")
	}
}

func TestJWTClaimsTimestamps(t *testing.T) {
	expiry := 2 * time.Hour
	issuer := NewJWTIssuer("test-secret", expiry)
	before := time.Now()

	token, err := issuer.Issue("user-123")
	if err != nil {
		t.Fatalf("Issue() error: %v", err)
	}

	claims, err := issuer.Validate(token)
	if err != nil {
		t.Fatalf("Validate() error: %v", err)
	}

	if claims.IssuedAt == nil {
		t.Fatal("IssuedAt should be set")
	}
	if claims.ExpiresAt == nil {
		t.Fatal("ExpiresAt should be set")
	}

	issuedAt := claims.IssuedAt.Time
	expiresAt := claims.ExpiresAt.Time

	if issuedAt.Before(before.Add(-1 * time.Second)) {
		t.Errorf("IssuedAt %v is before test start %v", issuedAt, before)
	}

	diff := expiresAt.Sub(issuedAt)
	if diff < expiry-time.Second || diff > expiry+time.Second {
		t.Errorf("ExpiresAt - IssuedAt = %v, want ~%v", diff, expiry)
	}
}
