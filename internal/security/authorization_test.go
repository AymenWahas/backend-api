package security

import "testing"

func TestCanAccessProject(t *testing.T) {
	if !CanAccessProject(RoleMember, true, true) {
		t.Fatal("member should access project when active")
	}

	if !CanAccessProject(RoleManager, true, true) {
		t.Fatal("manager should access project when active")
	}

	if !CanAccessProject(RoleAdmin, true, true) {
		t.Fatal("admin should access project when active")
	}

	if !CanAccessProject(RoleOwner, true, true) {
		t.Fatal("owner should access project when active")
	}

	if CanAccessProject(RoleMember, false, false) {
		t.Fatal("non-member should not access project")
	}

	if CanAccessProject(RoleAdmin, false, false) {
		t.Fatal("admin without membership should not access project")
	}

	if CanAccessProject(RoleAdmin, true, false) {
		t.Fatal("inactive admin should not access project")
	}

	if CanAccessProject(RoleMember, true, false) {
		t.Fatal("inactive member should not access project")
	}
}

func TestCanModifyProject(t *testing.T) {
	if !CanModifyProject(RoleOwner, true) {
		t.Fatal("owner should modify project")
	}

	if !CanModifyProject(RoleAdmin, false) {
		t.Fatal("admin should modify project")
	}

	if !CanModifyProject(RoleManager, false) {
		t.Fatal("manager should modify project")
	}

	if CanModifyProject(RoleMember, false) {
		t.Fatal("member should not modify project")
	}
}

func TestCanDeleteProject(t *testing.T) {
	if !CanDeleteProject(RoleOwner, true) {
		t.Fatal("owner should delete project")
	}

	if !CanDeleteProject(RoleAdmin, false) {
		t.Fatal("admin should delete project")
	}

	if CanDeleteProject(RoleManager, false) {
		t.Fatal("manager should not delete project")
	}

	if CanDeleteProject(RoleMember, false) {
		t.Fatal("member should not delete project")
	}
}

func TestCanModifyTask(t *testing.T) {
	if !CanModifyTask(RoleOwner, true) {
		t.Fatal("project owner should modify task")
	}

	if !CanModifyTask(RoleAdmin, false) {
		t.Fatal("admin should modify task")
	}

	if !CanModifyTask(RoleManager, false) {
		t.Fatal("manager should modify task")
	}

	if CanModifyTask(RoleMember, false) {
		t.Fatal("member should not modify task")
	}
}

func TestCanDeleteTask(t *testing.T) {
	if !CanDeleteTask(RoleOwner, true) {
		t.Fatal("project owner should delete task")
	}

	if !CanDeleteTask(RoleAdmin, false) {
		t.Fatal("admin should delete task")
	}

	if !CanDeleteTask(RoleManager, false) {
		t.Fatal("manager should delete task")
	}

	if CanDeleteTask(RoleMember, false) {
		t.Fatal("member should not delete task")
	}
}

func TestCanManageMembers(t *testing.T) {
	if !CanManageMembers(RoleOwner, true) {
		t.Fatal("owner should manage members")
	}

	if !CanManageMembers(RoleAdmin, false) {
		t.Fatal("admin should manage members")
	}

	if CanManageMembers(RoleManager, false) {
		t.Fatal("manager should not manage members")
	}

	if CanManageMembers(RoleMember, false) {
		t.Fatal("member should not manage members")
	}
}
