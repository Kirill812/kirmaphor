package models

import (
	"time"
	"github.com/google/uuid"
)

type ScheduleType string

const (
	ScheduleTypeCron  ScheduleType = "cron"
	ScheduleTypeRunAt ScheduleType = "run_at"
)

type Schedule struct {
	ID             uuid.UUID
	ProjectID      uuid.UUID
	TemplateID     uuid.UUID
	Name           string
	Type           ScheduleType
	CronFormat     *string
	RunAt          *time.Time
	Active         bool
	DeleteAfterRun bool
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
	LastRunAt      *time.Time
}
