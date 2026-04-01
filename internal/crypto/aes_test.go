package crypto_test

import (
	"errors"
	"testing"

	"github.com/kgory/kirmaphor/internal/crypto"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

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
	for i := range wrongKey {
		wrongKey[i] = 0xFF
	}

	ciphertext, nonce, err := crypto.Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	_, err = crypto.Decrypt(wrongKey, ciphertext, nonce)
	if err == nil {
		t.Fatal("expected error with wrong key")
	}
	if !errors.Is(err, crypto.ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got: %v", err)
	}
}

func TestDecryptFailsWithTamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ciphertext, nonce, err := crypto.Encrypt(key, []byte("ansible-vault-secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	// Flip a bit in the ciphertext
	ciphertext[0] ^= 0xFF
	_, err = crypto.Decrypt(key, ciphertext, nonce)
	if err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
	if !errors.Is(err, crypto.ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got: %v", err)
	}
}

func TestDecryptFailsWithWrongNonce(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	ciphertext, nonce, err := crypto.Encrypt(key, []byte("my-secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	// Use a different nonce
	wrongNonce := make([]byte, len(nonce))
	wrongNonce[0] = 0xFF
	_, err = crypto.Decrypt(key, ciphertext, wrongNonce)
	if err == nil {
		t.Fatal("expected error for wrong nonce")
	}
	if !errors.Is(err, crypto.ErrDecryptFailed) {
		t.Fatalf("expected ErrDecryptFailed, got: %v", err)
	}
}

func TestEncryptEmptyPlaintext(t *testing.T) {
	key := make([]byte, 32)
	ciphertext, nonce, err := crypto.Encrypt(key, []byte{})
	if err != nil {
		t.Fatalf("encrypt empty: %v", err)
	}
	result, err := crypto.Decrypt(key, ciphertext, nonce)
	if err != nil {
		t.Fatalf("decrypt empty: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d bytes", len(result))
	}
}

func TestLoadMasterKeyValid(t *testing.T) {
	hexKey := "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"
	key, err := crypto.LoadMasterKey(hexKey)
	if err != nil {
		t.Fatalf("expected valid key, got: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(key))
	}
}

func TestLoadMasterKeyInvalidHex(t *testing.T) {
	_, err := crypto.LoadMasterKey("ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ")
	if err == nil {
		t.Fatal("expected error for invalid hex")
	}
}

func TestLoadMasterKeyWrongLength(t *testing.T) {
	// 63 chars (should fail — 31.5 bytes)
	_, err := crypto.LoadMasterKey("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestLoadMasterKeyEmpty(t *testing.T) {
	_, err := crypto.LoadMasterKey("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}
