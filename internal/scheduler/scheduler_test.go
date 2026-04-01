package scheduler_test

import (
	"testing"
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

func TestIsDue(t *testing.T) {
	isDue, err := scheduler.IsCronDue("* * * * *", nil)
	if err != nil {
		t.Fatalf("IsCronDue: %v", err)
	}
	if !isDue {
		t.Fatal("expected every-minute cron to be due when never run")
	}
}
