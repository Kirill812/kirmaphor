package queries

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kgory/kirmaphor/internal/db/models"
)

func AppendLogs(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	rows := make([][]any, len(lines))
	for i, line := range lines {
		rows[i] = []any{taskID, line}
	}
	_, err := pool.CopyFrom(ctx,
		pgx.Identifier{"task_logs"},
		[]string{"task_id", "output"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func GetLogs(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID) ([]*models.TaskLog, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, task_id, output, created_at
		 FROM task_logs WHERE task_id = $1 ORDER BY id ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*models.TaskLog
	for rows.Next() {
		l := &models.TaskLog{}
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Output, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func GetLogsAfter(ctx context.Context, pool *pgxpool.Pool, taskID uuid.UUID, afterID int64) ([]*models.TaskLog, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, task_id, output, created_at
		 FROM task_logs WHERE task_id = $1 AND id > $2 ORDER BY id ASC`, taskID, afterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*models.TaskLog
	for rows.Next() {
		l := &models.TaskLog{}
		if err := rows.Scan(&l.ID, &l.TaskID, &l.Output, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
