package crypto_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key { key[i] = byte(i) }

	plaintext := []byte("my secret ansible vault password")
	ciphertext, nonce, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if len(ciphertext) == 0 || len(nonce) == 0 {
		t.Fatal("expected non-empty ciphertext and nonce")
	}

	result, err := crypto.Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if string(result) != string(plaintext) {
		t.Fatalf("got %q want %q", result, plaintext)
	}
}

func TestDecryptFailsWithWrongKey(t *testing.T) {
	key := make([]byte, 32)
	wrongKey := make([]byte, 32)
	for i := range wrongKey { wrongKey[i] = 0xFF }

	ciphertext, nonce, _ := crypto.Encrypt(key, []byte("secret"))
	_, err := crypto.Decrypt(wrongKey, ciphertext, nonce)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
}
