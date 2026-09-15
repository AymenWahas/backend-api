package security

type Role string

const (
	RoleMember  Role = "member"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
	RoleOwner   Role = "owner"
)

// CanAccessProject determines whether a user can access a project.
func CanAccessProject(role Role, isMember, isActive bool) bool {
	if !isMember || !isActive {
		return false
	}

	switch role {
	case RoleMember, RoleManager, RoleAdmin, RoleOwner:
		return true
	default:
		return false
	}
}

// CanModifyProject determines whether a user can modify a project.
func CanModifyProject(role Role, isOwner bool) bool {
	if isOwner {
		return true
	}

	switch role {
	case RoleAdmin, RoleOwner, RoleManager:
		return true
	default:
		return false
	}
}

// CanDeleteProject determines whether a user can delete a project.
func CanDeleteProject(role Role, isOwner bool) bool {
	if isOwner {
		return true
	}

	return role == RoleAdmin
}

// CanModifyTask determines whether a user can modify a task.
func CanModifyTask(role Role, isOwner bool) bool {
	if isOwner {
		return true
	}

	switch role {
	case RoleAdmin, RoleOwner, RoleManager:
		return true
	default:
		return false
	}
}

// CanDeleteTask determines whether a user can delete a task.
func CanDeleteTask(role Role, isOwner bool) bool {
	if isOwner {
		return true
	}

	switch role {
	case RoleAdmin, RoleOwner, RoleManager:
		return true
	default:
		return false
	}
}

// CanManageMembers determines whether a user can manage project members.
func CanManageMembers(role Role, isOwner bool) bool {
	if isOwner {
		return true
	}

	return role == RoleAdmin
}
