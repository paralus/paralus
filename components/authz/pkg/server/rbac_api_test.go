package server

import (
	"testing"

	pb "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc"
	"github.com/casbin/casbin/v2/util"
	"github.com/stretchr/testify/assert"
)

func testGetRoles(t *testing.T, e *testEngine, name string, res []string) {
	t.Helper()
	reply, err := e.s.GetRolesForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: name})
	assert.NoError(t, err)

	t.Log("Roles for ", name, ": ", reply.Array)

	if !util.SetEquals(res, reply.Array) {
		t.Error("Roles for ", name, ": ", reply.Array, ", supposed to be ", res)
	}
}

func testGetImplicitRoles(t *testing.T, e *testEngine, name string, res []string) {
	t.Helper()
	reply, err := e.s.GetImplicitRolesForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: name})
	assert.NoError(t, err)

	t.Log("Implicit Roles for ", name, ": ", reply.Array)

	if !util.SetEquals(res, reply.Array) {
		t.Error("Implicit Roles for ", name, ": ", reply.Array, ", supposed to be ", res)
	}
}

func testGetUsers(t *testing.T, e *testEngine, name string, res []string) {
	t.Helper()
	reply, err := e.s.GetUsersForRole(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: name})
	assert.NoError(t, err)

	t.Log("Users for ", name, ": ", reply.Array)

	if !util.SetEquals(res, reply.Array) {
		t.Error("Users for ", name, ": ", reply.Array, ", supposed to be ", res)
	}
}

func testHasRole(t *testing.T, e *testEngine, name string, role string, res bool) {
	t.Helper()
	reply, err := e.s.HasRoleForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: name, Role: role})
	assert.NoError(t, err)

	t.Log(name, " has role ", role, ": ", reply.Res)

	if res != reply.Res {
		t.Error(name, " has role ", role, ": ", reply.Res, ", supposed to be ", res)
	}
}

func TestRoleAPI(t *testing.T) {
	e := newTestEngine(t, "file", "../examples/rbac_policy.csv", "../examples/rbac_model.conf")

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})
	testGetRoles(t, e, "non_exist", []string{})

	testHasRole(t, e, "alice", "data1_admin", false)
	testHasRole(t, e, "alice", "data2_admin", true)

	_, err := e.s.AddRoleForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: "alice", Role: "data1_admin"})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{"data1_admin", "data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	_, err = e.s.DeleteRoleForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: "alice", Role: "data1_admin"})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	testGetImplicitRoles(t, e, "alice", []string{"data2_admin"})
	testGetImplicitRoles(t, e, "bob", []string{})
	testGetImplicitRoles(t, e, "george", []string{"data3_admin", "data4_admin"})

	_, err = e.s.DeleteRolesForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: "alice"})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	_, err = e.s.AddRoleForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: "alice", Role: "data1_admin"})
	assert.NoError(t, err)

	_, err = e.s.DeleteUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: "alice"})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "data2_admin", []string{})

	_, err = e.s.AddRoleForUser(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, User: "alice", Role: "data2_admin"})
	assert.NoError(t, err)

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", true)
	testEnforce(t, e, "alice", "data2", "write", true)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)
	testEnforce(t, e, "bob", "data4", "read", false)
	testEnforce(t, e, "george", "data4", "write", false)
	testEnforce(t, e, "george", "data4", "read", true)

	_, err = e.s.DeleteRole(e.ctx, &pb.UserRoleRequest{EnforcerHandler: e.h, Role: "data2_admin"})
	assert.NoError(t, err)

	testEnforce(t, e, "alice", "data1", "read", true)
	testEnforce(t, e, "alice", "data1", "write", false)
	testEnforce(t, e, "alice", "data2", "read", false)
	testEnforce(t, e, "alice", "data2", "write", false)
	testEnforce(t, e, "bob", "data1", "read", false)
	testEnforce(t, e, "bob", "data1", "write", false)
	testEnforce(t, e, "bob", "data2", "read", false)
	testEnforce(t, e, "bob", "data2", "write", true)

	testGetPermissions(t, e, "alice", [][]string{{"alice", "data1", "read"}}) //Added these in this class as it's part of RBAC.
	testGetPermissions(t, e, "bob", [][]string{{"bob", "data2", "write"}})
	testGetPermissions(t, e, "george", [][]string{})
	testGetPermissions(t, e, "data3_admin", [][]string{{"data3_admin", "data3", "admin"}})

	testGetImplicitPermissions(t, e, "bob", [][]string{{"bob", "data2", "write"}})
	testGetImplicitPermissions(t, e, "data3_admin", [][]string{{"data3_admin", "data3", "admin"}, {"data4_admin", "data4", "read"}})
}

func testGetPermissions(t *testing.T, e *testEngine, name string, res [][]string) {
	t.Helper()
	reply, err := e.s.GetPermissionsForUser(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, User: name})
	assert.NoError(t, err)

	myRes := extractFromArray2DReply(reply)
	t.Log("Permissions for ", name, ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Permissions for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testGetImplicitPermissions(t *testing.T, e *testEngine, name string, res [][]string) {
	t.Helper()
	reply, err := e.s.GetImplicitPermissionsForUser(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, User: name})
	assert.NoError(t, err)

	myRes := extractFromArray2DReply(reply)
	t.Log("Implicit Permissions for ", name, ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Implicit Permissions for ", name, ": ", myRes, ", supposed to be ", res)
	}
}

func testHasPermission(t *testing.T, e *testEngine, name string, permission []string, res bool) {
	t.Helper()
	reply, err := e.s.HasPermissionForUser(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, User: name, Permissions: permission})
	assert.NoError(t, err)

	t.Log(name, " has permission ", util.ArrayToString(permission), ": ", reply.Res)

	if res != reply.Res {
		t.Error(name, " has permission ", util.ArrayToString(permission), ": ", reply.Res, ", supposed to be ", res)
	}
}

func TestPermissionAPI(t *testing.T) {
	e := newTestEngine(t, "file", "../examples/basic_without_resources_policy.csv",
		"../examples/basic_without_resources_model.conf")

	testEnforceWithoutUsers(t, e, "alice", "read", true)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	testGetPermissions(t, e, "alice", [][]string{{"alice", "read"}})
	testGetPermissions(t, e, "bob", [][]string{{"bob", "write"}})

	testHasPermission(t, e, "alice", []string{"read"}, true)
	testHasPermission(t, e, "alice", []string{"write"}, false)
	testHasPermission(t, e, "bob", []string{"read"}, false)
	testHasPermission(t, e, "bob", []string{"write"}, true)

	_, err := e.s.DeletePermission(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, Permissions: []string{"read"}})
	assert.NoError(t, err)

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	_, err = e.s.AddPermissionForUser(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, User: "bob", Permissions: []string{"read"}})
	assert.NoError(t, err)

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", true)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	_, err = e.s.DeletePermissionForUser(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, User: "bob", Permissions: []string{"read"}})
	assert.NoError(t, err)

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", true)

	_, err = e.s.DeletePermissionsForUser(e.ctx, &pb.PermissionRequest{EnforcerHandler: e.h, User: "bob"})
	assert.NoError(t, err)

	testEnforceWithoutUsers(t, e, "alice", "read", false)
	testEnforceWithoutUsers(t, e, "alice", "write", false)
	testEnforceWithoutUsers(t, e, "bob", "read", false)
	testEnforceWithoutUsers(t, e, "bob", "write", false)
}
