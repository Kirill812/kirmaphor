package scheduler_test

import (
	"testing"
	"time"

	"github.com/kgory/kirmaphor/internal/scheduler"
)

func TestValidateCronFormat(t *testing.T) {
	valid := []string{
		"0 9 * * 1-5",
		"*/5 * * * *",
		"0 0 1 * *",
	}
	invalid := []string{
		"not a cron",
		"99 * * * *",
		"",
	}
	for _, c := range valid {
		if err := scheduler.ValidateCronFormat(c); err != nil {
			t.Errorf("expected %q to be valid, got: %v", c, err)
		}
	}
	for _, c := range invalid {
		if err := scheduler.ValidateCronFormat(c); err == nil {
			t.Errorf("expected %q to be invalid", c)
		}
	}
}

func TestIsDuePastLastRun(t *testing.T) {
	// Every-minute cron, last run 2 minutes ago -> should be due
	twoMinutesAgo := time.Now().Add(-2 * time.Minute)
	isDue, err := scheduler.IsCronDue("* * * * *", &twoMinutesAgo)
	if err != nil {
		t.Fatalf("IsCronDue: %v", err)
	}
	if !isDue {
		t.Fatal("expected cron to be due when last run 2 minutes ago")
	}
}

func TestIsDue(t *testing.T) {
	isDue, err := scheduler.IsCronDue("* * * * *", nil)
	if err != nil {
		t.Fatalf("IsCronDue: %v", err)
	}
	if !isDue {
		t.Fatal("expected every-minute cron to be due when never run")
	}
}
