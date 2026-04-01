package queries

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func CreateSchedule(ctx context.Context, pool *pgxpool.Pool, s *models.Schedule) (*models.Schedule, error) {
	result := &models.Schedule{}
	err := pool.QueryRow(ctx,
		`INSERT INTO schedules
		   (project_id, template_id, name, type, cron_format, run_at, active, delete_after_run, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, project_id, template_id, name, type, cron_format, run_at,
		           active, delete_after_run, created_by, created_at, last_run_at`,
		s.ProjectID, s.TemplateID, s.Name, s.Type, s.CronFormat, s.RunAt,
		s.Active, s.DeleteAfterRun, s.CreatedBy,
	).Scan(&result.ID, &result.ProjectID, &result.TemplateID, &result.Name, &result.Type,
		&result.CronFormat, &result.RunAt, &result.Active, &result.DeleteAfterRun,
		&result.CreatedBy, &result.CreatedAt, &result.LastRunAt)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetSchedule(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Schedule, error) {
	s := &models.Schedule{}
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, template_id, name, type, cron_format, run_at,
		        active, delete_after_run, created_by, created_at, last_run_at
		 FROM schedules WHERE id = $1`, id,
	).Scan(&s.ID, &s.ProjectID, &s.TemplateID, &s.Name, &s.Type, &s.CronFormat,
		&s.RunAt, &s.Active, &s.DeleteAfterRun, &s.CreatedBy, &s.CreatedAt, &s.LastRunAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return s, nil
}

func ListSchedules(ctx context.Context, pool *pgxpool.Pool, projectID uuid.UUID) ([]*models.Schedule, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, template_id, name, type, cron_format, run_at,
		        active, delete_after_run, created_by, created_at, last_run_at
		 FROM schedules WHERE project_id = $1 ORDER BY name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []*models.Schedule
	for rows.Next() {
		s := &models.Schedule{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.TemplateID, &s.Name, &s.Type,
			&s.CronFormat, &s.RunAt, &s.Active, &s.DeleteAfterRun,
			&s.CreatedBy, &s.CreatedAt, &s.LastRunAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// GetDueSchedules returns active schedules that may need to run.
// For 'cron' type, all active cron schedules are returned — the caller
// (scheduler) is responsible for evaluating the cron expression against
// last_run_at to decide whether to fire. For 'run_at' type, only schedules
// whose run_at <= NOW() are returned.
func GetDueSchedules(ctx context.Context, pool *pgxpool.Pool) ([]*models.Schedule, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, template_id, name, type, cron_format, run_at,
		        active, delete_after_run, created_by, created_at, last_run_at
		 FROM schedules
		 WHERE active = TRUE
		   AND (
		     (type = 'run_at' AND run_at <= NOW())
		     OR type = 'cron'
		   )`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedules []*models.Schedule
	for rows.Next() {
		s := &models.Schedule{}
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.TemplateID, &s.Name, &s.Type,
			&s.CronFormat, &s.RunAt, &s.Active, &s.DeleteAfterRun,
			&s.CreatedBy, &s.CreatedAt, &s.LastRunAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func TouchSchedule(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, lastRunAt time.Time) error {
	_, err := pool.Exec(ctx,
		`UPDATE schedules SET last_run_at = $1 WHERE id = $2`, lastRunAt, id)
	return err
}

func DeleteSchedule(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) error {
	tag, err := pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
