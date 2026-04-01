package models

import (
	"time"
	"github.com/google/uuid"
)

type TaskLog struct {
	ID        int64
	TaskID    uuid.UUID
	Output    string
	CreatedAt time.Time
}
