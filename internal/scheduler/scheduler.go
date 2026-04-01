package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/db/queries"
	"github.com/kgory/kirmaphor/internal/execution"
	"github.com/robfig/cron/v3"
)

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// ValidateCronFormat returns an error if the cron expression is invalid.
func ValidateCronFormat(expr string) error {
	if expr == "" {
		return fmt.Errorf("cron format is empty")
	}
	_, err := parser.Parse(expr)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}
	return nil
}

// IsCronDue returns true if the cron expression is due given the last run time.
// If lastRun is nil, it is always due.
func IsCronDue(expr string, lastRun *time.Time) (bool, error) {
	schedule, err := parser.Parse(expr)
	if err != nil {
		return false, fmt.Errorf("parse cron: %w", err)
	}
	if lastRun == nil {
		return true, nil
	}
	next := schedule.Next(*lastRun)
	return next.Before(time.Now()), nil
}

// Scheduler polls the DB every 30s and enqueues due tasks.
type Scheduler struct {
	pool     *pgxpool.Pool
	taskPool *execution.TaskPool
	deps     execution.RunnerDeps
}

func New(pool *pgxpool.Pool, taskPool *execution.TaskPool, deps execution.RunnerDeps) *Scheduler {
	return &Scheduler{pool: pool, taskPool: taskPool, deps: deps}
}

// Run starts the scheduler loop. Blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	schedules, err := queries.GetDueSchedules(ctx, s.pool)
	if err != nil {
		log.Printf("scheduler: get due schedules: %v", err)
		return
	}
	for _, sched := range schedules {
		if err := s.process(ctx, sched); err != nil {
			log.Printf("scheduler: process schedule %s: %v", sched.ID, err)
		}
	}
}

func (s *Scheduler) process(ctx context.Context, sched *models.Schedule) error {
	// For cron: check if actually due
	if sched.Type == models.ScheduleTypeCron {
		if sched.CronFormat == nil {
			return fmt.Errorf("cron schedule %s has no format", sched.ID)
		}
		due, err := IsCronDue(*sched.CronFormat, sched.LastRunAt)
		if err != nil {
			return err
		}
		if !due {
			return nil
		}
	}

	// Load template to build task
	tmpl, err := queries.GetTemplate(ctx, s.pool, sched.TemplateID)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	schedID := sched.ID
	task := &models.Task{
		ProjectID:    sched.ProjectID,
		TemplateID:   tmpl.ID,
		Playbook:     tmpl.Playbook,
		InventoryID:  tmpl.InventoryID,
		RepositoryID: tmpl.RepositoryID,
		GitBranch:    "main",
		Arguments:    tmpl.Arguments,
		Environment:  tmpl.Environment,
		CreatedBy:    tmpl.CreatedBy,
		ScheduleID:   &schedID,
	}

	created, err := queries.CreateTask(ctx, s.pool, task)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	taskCopy := *created
	deps := s.deps
	s.taskPool.Enqueue(execution.TaskRequest{
		TaskID: created.ID,
		Run: func(ctx context.Context) {
			execution.RunTask(ctx, deps, &taskCopy)
		},
	})

	// Touch schedule
	now := time.Now()
	if err := queries.TouchSchedule(ctx, s.pool, sched.ID, now); err != nil {
		log.Printf("scheduler: touch schedule %s: %v", sched.ID, err)
	}

	// Delete if one-shot
	if sched.DeleteAfterRun {
		if err := queries.DeleteSchedule(ctx, s.pool, sched.ID); err != nil {
			log.Printf("scheduler: delete one-shot schedule %s: %v", sched.ID, err)
		}
	}

	return nil
}
