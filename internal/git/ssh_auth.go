package git

import (
	"fmt"
	"os"
)

// WriteKeyFile writes PEM key bytes to a temp file with 0600 permissions.
// Returns (path, cleanup func, error). Caller must call cleanup() when done.
func WriteKeyFile(keyPEM []byte) (string, func(), error) {
	f, err := os.CreateTemp("", "kirmaphore-sshkey-*.pem")
	if err != nil {
		return "", nil, fmt.Errorf("create temp key file: %w", err)
	}
	if err := f.Chmod(0600); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("chmod key file: %w", err)
	}
	if _, err := f.Write(keyPEM); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("write key file: %w", err)
	}
	f.Close()
	path := f.Name()
	return path, func() { os.Remove(path) }, nil
}
