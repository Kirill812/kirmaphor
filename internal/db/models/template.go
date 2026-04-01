package models

import (
	"time"
	"github.com/google/uuid"
)

type JobTemplate struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	Name         string
	Description  string
	Playbook     string
	InventoryID  *uuid.UUID
	RepositoryID uuid.UUID
	Environment  map[string]string
	Arguments    string
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
}
