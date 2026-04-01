package rbac

func HasPermission(role Role, perm Permission) bool {
	return rolePermissions[role]&perm != 0
}
