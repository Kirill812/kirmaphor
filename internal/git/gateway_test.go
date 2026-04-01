package git_test

import (
	"os"
	"testing"
	"github.com/kgory/kirmaphor/internal/git"
)

func TestClonePublicRepo(t *testing.T) {
	// Clone a small public repo to a temp dir
	dir, cleanup, err := git.CloneOrPull("https://github.com/nicholaswilde/hello-world-ansible.git", "main", nil)
	if err != nil {
		t.Skipf("skipping: network unavailable or repo changed: %v", err)
	}
	defer cleanup()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("expected clone dir to exist at %s", dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one file in cloned repo")
	}
}

func TestWriteKeyFileCleansUp(t *testing.T) {
	keyPEM := []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nfakekey\n-----END OPENSSH PRIVATE KEY-----")
	path, cleanup, err := git.WriteKeyFile(keyPEM)
	if err != nil {
		t.Fatalf("WriteKeyFile: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("key file should exist at %s", path)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", info.Mode().Perm())
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("key file should be removed after cleanup")
	}
}
