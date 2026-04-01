package models

import (
	"time"
	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskStatusWaiting TaskStatus = "waiting"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusError   TaskStatus = "error"
	TaskStatusStopped TaskStatus = "stopped"
)

type Task struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	TemplateID   uuid.UUID
	Status       TaskStatus
	Message      string
	Playbook     string
	InventoryID  *uuid.UUID
	RepositoryID uuid.UUID
	GitBranch    string
	CommitHash   *string
	Arguments    string
	Environment  map[string]string
	CreatedBy    uuid.UUID
	ScheduleID   *uuid.UUID
	CreatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}
