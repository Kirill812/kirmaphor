package execution

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	gitpkg "github.com/kgory/kirmaphor/internal/git"
	"github.com/kgory/kirmaphor/internal/inventory"
)

// BuildAnsibleArgs builds the argument list for ansible-playbook.
func BuildAnsibleArgs(playbookPath, inventoryPath, extraArgs string) []string {
	args := []string{"-i", inventoryPath, playbookPath}
	if extraArgs != "" {
		for _, a := range strings.Fields(extraArgs) {
			args = append(args, a)
		}
	}
	return args
}

// BuildEnv converts a map of env vars to KEY=VALUE slice for exec.Cmd.Env.
func BuildEnv(env map[string]string) []string {
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// RunnerDeps holds dependencies injected into RunTask.
type RunnerDeps struct {
	Pool    *pgxpool.Pool
	Decrypt func(encrypted, nonce []byte) ([]byte, error)
}

// RunTask executes a task end-to-end:
// 1. Loads repository credentials and clones the repo
// 2. Loads inventory to temp file
// 3. Runs ansible-playbook via LocalJob
// 4. Streams logs to DB via LogWriter (500ms batch)
// 5. Updates task status on completion
func RunTask(ctx context.Context, deps RunnerDeps, task *models.Task) {
	pool := deps.Pool
	taskID := task.ID

	setStatus := func(status models.TaskStatus, msg string) {
		if err := queries.UpdateTaskStatus(ctx, pool, taskID, status, msg); err != nil {
			log.Printf("task %s: update status: %v", taskID, err)
		}
	}

	setStatus(models.TaskStatusRunning, "")

	// 1. Clone repository
	repo, err := queries.GetRepository(ctx, pool, task.RepositoryID)
	if err != nil {
		setStatus(models.TaskStatusError, fmt.Sprintf("get repository: %v", err))
		return
	}

	var keyPEM []byte
	if repo.SSHKeyID != nil {
		secret, err := queries.GetSecret(ctx, pool, *repo.SSHKeyID)
		if err != nil {
			setStatus(models.TaskStatusError, fmt.Sprintf("get ssh key: %v", err))
			return
		}
		keyPEM, err = deps.Decrypt(secret.EncryptedValue, secret.Nonce)
		if err != nil {
			setStatus(models.TaskStatusError, "decrypt ssh key failed")
			return
		}
	}

	workDir, repoCleanup, err := gitpkg.CloneOrPull(repo.GitURL, task.GitBranch, keyPEM)
	if err != nil {
		setStatus(models.TaskStatusError, fmt.Sprintf("git clone: %v", err))
		return
	}
	defer repoCleanup()

	// Store commit hash
	if hash, err := gitpkg.CommitHash(workDir); err == nil {
		queries.UpdateTaskCommit(ctx, pool, taskID, hash)
	}

	// 2. Load inventory
	var invFilePath string
	var invCleanup func()

	if task.InventoryID != nil {
		inv, err := queries.GetInventory(ctx, pool, *task.InventoryID)
		if err != nil {
			setStatus(models.TaskStatusError, fmt.Sprintf("get inventory: %v", err))
			return
		}
		invFilePath, invCleanup, err = inventory.Load(inv)
		if err != nil {
			setStatus(models.TaskStatusError, fmt.Sprintf("load inventory: %v", err))
			return
		}
		defer invCleanup()
	} else {
		invFilePath = "localhost,"
	}

	// 3. Set up log writer (flush every 500ms — Semaphore pattern)
	logWriter := NewLogWriter(func(lines []string) error {
		return queries.AppendLogs(ctx, pool, taskID, lines)
	}, 500*time.Millisecond)
	defer logWriter.Close()

	// 4. Run job
	extraArgs := []string{}
	if task.Arguments != "" {
		extraArgs = strings.Fields(task.Arguments)
	}
	extraEnv := BuildEnv(task.Environment)

	job := NewLocalJob(task.Playbook, invFilePath, workDir, extraArgs, extraEnv)
	output := make(chan string, 100)

	go func() {
		for line := range output {
			logWriter.Write(line)
		}
	}()

	runErr := job.Run(ctx, output)
	close(output)

	// 5. Final status
	if job.IsKilled() {
		setStatus(models.TaskStatusStopped, "killed by user")
		return
	}
	if runErr != nil {
		setStatus(models.TaskStatusError, runErr.Error())
		return
	}
	setStatus(models.TaskStatusSuccess, "")
}
