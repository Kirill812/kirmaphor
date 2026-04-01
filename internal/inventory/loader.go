package inventory

import (
	"fmt"
	"os"

	"github.com/kgory/kirmaphor/internal/db/models"
)

// Load materialises an inventory to a temp file.
// Returns (tmpFilePath, cleanup func, error).
// Caller must call cleanup() after ansible-playbook finishes.
func Load(inv *models.Inventory) (string, func(), error) {
	switch inv.Type {
	case models.InventoryTypeStatic, models.InventoryTypeStaticYAML:
		return loadInline(inv)
	default:
		return "", nil, fmt.Errorf("inventory type %q not yet supported", inv.Type)
	}
}

func loadInline(inv *models.Inventory) (string, func(), error) {
	if inv.InventoryData == nil || *inv.InventoryData == "" {
		return "", nil, fmt.Errorf("inventory_data is empty for static inventory %s", inv.ID)
	}
	f, err := os.CreateTemp("", "kirmaphore-inventory-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp inventory: %w", err)
	}
	if _, err := f.WriteString(*inv.InventoryData); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("write inventory: %w", err)
	}
	f.Close()
	path := f.Name()
	return path, func() { os.Remove(path) }, nil
}
