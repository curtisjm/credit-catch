package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "my-secure-password-123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword() returned empty string")
	}
	if hash == password {
		t.Fatal("HashPassword() returned plaintext password")
	}

	if !CheckPassword(hash, password) {
		t.Error("CheckPassword() returned false for correct password")
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if CheckPassword(hash, "wrong-password") {
		t.Error("CheckPassword() returned true for wrong password")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	password := "same-password"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error: %v", err)
	}

	if hash1 == hash2 {
		t.Error("two hashes of the same password should differ (different salts)")
	}

	// Both should still validate.
	if !CheckPassword(hash1, password) {
		t.Error("CheckPassword() failed for hash1")
	}
	if !CheckPassword(hash2, password) {
		t.Error("CheckPassword() failed for hash2")
	}
}

func TestCheckPasswordEmptyHash(t *testing.T) {
	if CheckPassword("", "any-password") {
		t.Error("CheckPassword() should return false for empty hash")
	}
}
