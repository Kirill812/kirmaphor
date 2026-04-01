package crypto

import (
	"encoding/hex"
	"fmt"
)

// LoadMasterKey decodes a 64-char hex string into a 32-byte AES key.
func LoadMasterKey(hexKey string) ([]byte, error) {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid MASTER_KEY hex: %w", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("MASTER_KEY must be 32 bytes (64 hex chars), got %d", len(b))
	}
	return b, nil
}
