package auth

import "testing"

func TestHashTokenDeterministic(t *testing.T) {
	token := "abc123def456"
	h1 := hashToken(token)
	h2 := hashToken(token)

	if h1 != h2 {
		t.Error("hashToken should be deterministic")
	}
}

func TestHashTokenDifferentInputs(t *testing.T) {
	h1 := hashToken("token-a")
	h2 := hashToken("token-b")

	if h1 == h2 {
		t.Error("different tokens should produce different hashes")
	}
}

func TestHashTokenNotPlaintext(t *testing.T) {
	token := "my-secret-token"
	h := hashToken(token)

	if h == token {
		t.Error("hash should not equal the plaintext token")
	}
}
