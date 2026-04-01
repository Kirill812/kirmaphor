package rbac_test

import (
	"testing"
	"github.com/kgory/kirmaphor/internal/rbac"
)

func TestOwnerCanDoEverything(t *testing.T) {
	for _, perm := range []rbac.Permission{
		rbac.PermRunJobs, rbac.PermEditPlaybooks, rbac.PermManageUsers, rbac.PermManageSecrets, rbac.PermDeleteProject,
	} {
		if !rbac.HasPermission(rbac.RoleOwner, perm) {
			t.Errorf("owner should have permission %d", perm)
		}
	}
}

func TestViewerCanOnlyRead(t *testing.T) {
	if !rbac.HasPermission(rbac.RoleViewer, rbac.PermReadLogs) {
		t.Error("viewer should be able to read logs")
	}
	if rbac.HasPermission(rbac.RoleViewer, rbac.PermRunJobs) {
		t.Error("viewer should NOT be able to run jobs")
	}
	if rbac.HasPermission(rbac.RoleViewer, rbac.PermManageSecrets) {
		t.Error("viewer should NOT manage secrets")
	}
}

func TestEngineerCanRunButNotManage(t *testing.T) {
	if !rbac.HasPermission(rbac.RoleEngineer, rbac.PermRunJobs) {
		t.Error("engineer should run jobs")
	}
	if rbac.HasPermission(rbac.RoleEngineer, rbac.PermManageUsers) {
		t.Error("engineer should NOT manage users")
	}
}

func TestAdminCanManageButNotDelete(t *testing.T) {
	if !rbac.HasPermission(rbac.RoleAdmin, rbac.PermManageUsers) {
		t.Error("admin should manage users")
	}
	if !rbac.HasPermission(rbac.RoleAdmin, rbac.PermManageSecrets) {
		t.Error("admin should manage secrets")
	}
	if rbac.HasPermission(rbac.RoleAdmin, rbac.PermDeleteProject) {
		t.Error("admin should NOT be able to delete the project (owner only)")
	}
}
