package models

import (
	"time"
	"github.com/google/uuid"
)

type InventoryType string

const (
	InventoryTypeStatic     InventoryType = "static"
	InventoryTypeStaticYAML InventoryType = "static-yaml"
	InventoryTypeFile       InventoryType = "file"
	InventoryTypeAWSEC2     InventoryType = "aws-ec2"
	InventoryTypeAzureVMSS  InventoryType = "azure-vmss"
	InventoryTypeGCPGCE     InventoryType = "gcp-gce"
)

type Inventory struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	Name          string
	Type          InventoryType
	InventoryData *string
	RepositoryID  *uuid.UUID
	InventoryPath *string
	SSHKeyID      *uuid.UUID
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
}
