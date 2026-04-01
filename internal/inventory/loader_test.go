package inventory_test

import (
	"os"
	"strings"
	"testing"

	"github.com/kgory/kirmaphor/internal/db/models"
	"github.com/kgory/kirmaphor/internal/inventory"
)

func TestLoadStaticInventory(t *testing.T) {
	content := "[web]\n192.168.1.10\n192.168.1.11\n"
	inv := &models.Inventory{
		Type:          models.InventoryTypeStatic,
		InventoryData: &content,
	}

	path, cleanup, err := inventory.Load(inv)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "192.168.1.10") {
		t.Fatalf("expected inventory content in file, got: %s", data)
	}
}

func TestLoadStaticInventoryNilData(t *testing.T) {
	inv := &models.Inventory{
		Type:          models.InventoryTypeStatic,
		InventoryData: nil,
	}
	_, _, err := inventory.Load(inv)
	if err == nil {
		t.Fatal("expected error for nil inventory data")
	}
}

func TestLoadUnsupportedType(t *testing.T) {
	inv := &models.Inventory{
		Type: models.InventoryTypeAWSEC2,
	}
	_, _, err := inventory.Load(inv)
	if err == nil {
		t.Fatal("expected error for unsupported inventory type")
	}
}

func TestLoadStaticInventoryEmptyData(t *testing.T) {
	empty := ""
	inv := &models.Inventory{
		Type:          models.InventoryTypeStatic,
		InventoryData: &empty,
	}
	_, _, err := inventory.Load(inv)
	if err == nil {
		t.Fatal("expected error for empty inventory data")
	}
}
