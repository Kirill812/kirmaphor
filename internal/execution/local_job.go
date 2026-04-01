package execution

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sync/atomic"
)

// LocalJob runs ansible-playbook as a subprocess.
type LocalJob struct {
	playbookPath  string
	inventoryPath string
	workDir       string
	extraArgs     []string
	extraEnv      []string
	cmd           *exec.Cmd
	killed        atomic.Bool
}

func NewLocalJob(playbookPath, inventoryPath, workDir string, extraArgs, extraEnv []string) *LocalJob {
	return &LocalJob{
		playbookPath:  playbookPath,
		inventoryPath: inventoryPath,
		workDir:       workDir,
		extraArgs:     extraArgs,
		extraEnv:      extraEnv,
	}
}

func (j *LocalJob) Run(ctx context.Context, output chan<- string) error {
	args := append([]string{"-i", j.inventoryPath, j.playbookPath}, j.extraArgs...)
	j.cmd = exec.CommandContext(ctx, "ansible-playbook", args...)
	j.cmd.Dir = j.workDir
	j.cmd.Env = append(j.cmd.Environ(), j.extraEnv...)

	stdout, err := j.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := j.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := j.cmd.Start(); err != nil {
		return fmt.Errorf("start ansible-playbook: %w", err)
	}

	done := make(chan struct{}, 2)
	scanAndSend := func(r interface{ Read([]byte) (int, error) }) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			select {
			case output <- scanner.Text():
			case <-ctx.Done():
				done <- struct{}{}
				return
			}
		}
		done <- struct{}{}
	}
	go scanAndSend(stdout)
	go scanAndSend(stderr)

	<-done
	<-done

	if err := j.cmd.Wait(); err != nil {
		if j.killed.Load() {
			return nil // killed intentionally
		}
		return fmt.Errorf("ansible-playbook: %w", err)
	}
	return nil
}

func (j *LocalJob) Kill() {
	j.killed.Store(true)
	if j.cmd != nil && j.cmd.Process != nil {
		j.cmd.Process.Kill()
	}
}

func (j *LocalJob) IsKilled() bool {
	return j.killed.Load()
}
