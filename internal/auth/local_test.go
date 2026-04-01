package auth_test

import (
	"testing"

	"github.com/kgory/kirmaphor/internal/auth"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-horse-battery")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := auth.VerifyPassword(hash, "correct-horse-battery"); err != nil {
		t.Fatal("expected valid password to pass verification")
	}
	if err := auth.VerifyPassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail verification")
	}
}
