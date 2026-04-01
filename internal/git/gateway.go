package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CloneOrPull clones a git repository to a temp directory.
// keyPEM is optional (nil for public repos / HTTPS).
// Returns (workDir, cleanup, error). Caller must call cleanup() when done.
func CloneOrPull(gitURL, branch string, keyPEM []byte) (string, func(), error) {
	workDir, err := os.MkdirTemp("", "kirmaphore-repo-*")
	if err != nil {
		return "", nil, fmt.Errorf("create work dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(workDir) }

	env := os.Environ()
	var keyCleanup func()

	if len(keyPEM) > 0 {
		keyPath, kc, err := WriteKeyFile(keyPEM)
		if err != nil {
			cleanup()
			return "", nil, err
		}
		keyCleanup = kc
		env = append(env,
			fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", keyPath))
	}

	args := []string{"clone", "--depth=1", "--branch", branch, gitURL, workDir}
	cmd := exec.Command("git", args...)
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if keyCleanup != nil {
		keyCleanup()
	}
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("git clone: %w\noutput: %s", err, out)
	}
	return workDir, cleanup, nil
}

// CommitHash returns the HEAD commit hash for a cloned repo at workDir.
func CommitHash(workDir string) (string, error) {
	cmd := exec.Command("git", "-C", workDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}
