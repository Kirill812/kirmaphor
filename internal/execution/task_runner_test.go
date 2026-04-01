package execution_test

import (
	"testing"

	"github.com/kgory/kirmaphor/internal/execution"
)

func TestTaskRunnerBuildArgs(t *testing.T) {
	args := execution.BuildAnsibleArgs("site.yml", "/tmp/inv", "--check --diff")
	if len(args) < 3 {
		t.Fatalf("expected at least 3 args, got %v", args)
	}
	found := false
	for _, a := range args {
		if a == "--check" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected --check in args, got %v", args)
	}
}

func TestTaskRunnerBuildEnv(t *testing.T) {
	env := execution.BuildEnv(map[string]string{
		"ANSIBLE_NOCOWS": "1",
		"MY_VAR":         "hello",
	})
	found := false
	for _, e := range env {
		if e == "ANSIBLE_NOCOWS=1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected ANSIBLE_NOCOWS=1 in env, got %v", env)
	}
}
