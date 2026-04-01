package rbac

type Role string

const (
	RoleOwner    Role = "owner"
	RoleAdmin    Role = "admin"
	RoleEngineer Role = "engineer"
	RoleViewer   Role = "viewer"
)

type Permission uint32

const (
	PermReadLogs      Permission = 1 << 0
	PermRunJobs       Permission = 1 << 1
	PermEditPlaybooks Permission = 1 << 2
	PermManageSecrets Permission = 1 << 3
	PermManageUsers   Permission = 1 << 4
	PermManageProject Permission = 1 << 5
	PermDeleteProject Permission = 1 << 6
)

var rolePermissions = map[Role]Permission{
	RoleViewer:   PermReadLogs,
	RoleEngineer: PermReadLogs | PermRunJobs | PermEditPlaybooks,
	RoleAdmin:    PermReadLogs | PermRunJobs | PermEditPlaybooks | PermManageSecrets | PermManageUsers | PermManageProject,
	RoleOwner:    ^Permission(0), // all bits set
}
