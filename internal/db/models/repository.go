package models

import (
	"time"
	"github.com/google/uuid"
)

type Repository struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	Name      string
	GitURL    string
	GitBranch string
	SSHKeyID  *uuid.UUID
	CreatedBy uuid.UUID
	CreatedAt time.Time
}
