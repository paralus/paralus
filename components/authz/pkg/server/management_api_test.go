package server

import (
	"testing"

	pb "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc"
	"github.com/casbin/casbin/v2/util"
	"github.com/stretchr/testify/assert"
)

func testStringList(t *testing.T, title string, f func() []string, res []string) {
	t.Helper()
	myRes := f()
	t.Log(title+": ", myRes)

	if !util.ArrayEquals(res, myRes) {
		t.Error(title+": ", myRes, ", supposed to be ", res)
	}
}

func extractFromArrayReply(reply *pb.ArrayReply) func() []string {
	return func() []string {
		return reply.Array
	}
}

func TestGetList(t *testing.T) {
	e := newTestEngine(t, "file", "../examples/rbac_policy.csv", "../examples/rbac_model.conf")

	subjects, err := e.s.GetAllSubjects(e.ctx, &pb.EmptyRequest{Handler: e.h})
	if err != nil {
		t.Fatal(err)
	}
	testStringList(t, "Subjects", extractFromArrayReply(subjects), []string{"alice", "bob", "data2_admin", "data3_admin", "data4_admin"})

	objects, err := e.s.GetAllObjects(e.ctx, &pb.EmptyRequest{Handler: e.h})
	if err != nil {
		t.Fatal(err)
	}
	testStringList(t, "Objects", extractFromArrayReply(objects), []string{"data1", "data2", "data3", "data4"})

	actions, err := e.s.GetAllActions(e.ctx, &pb.EmptyRequest{Handler: e.h})
	if err != nil {
		t.Fatal(err)
	}
	testStringList(t, "Actions", extractFromArrayReply(actions), []string{"read", "write", "admin"})

	roles, err := e.s.GetAllRoles(e.ctx, &pb.EmptyRequest{Handler: e.h})
	if err != nil {
		t.Fatal(err)
	}
	testStringList(t, "Roles", extractFromArrayReply(roles), []string{"data2_admin", "data3_admin", "data4_admin"})
}

func extractFromArray2DReply(reply *pb.Array2DReply) [][]string {
	array2d := make([][]string, len(reply.D2))
	for i := 0; i < len(reply.D2); i++ {
		array2d[i] = reply.D2[i].D1
	}

	return array2d
}

func testGetPolicy(t *testing.T, e *testEngine, res [][]string) {
	t.Helper()
	req := &pb.EmptyRequest{Handler: e.h}
	reply, err := e.s.GetPolicy(e.ctx, req)
	assert.NoError(t, err)

	myRes := extractFromArray2DReply(reply)
	t.Log("Policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredPolicy(t *testing.T, e *testEngine, fieldIndex int, res [][]string, fieldValues ...string) {
	t.Helper()
	req := &pb.FilteredPolicyRequest{
		EnforcerHandler: e.h, FieldIndex: int32(fieldIndex), FieldValues: fieldValues}
	reply, err := e.s.GetFilteredPolicy(e.ctx, req)
	assert.NoError(t, err)

	myRes := extractFromArray2DReply(reply)
	t.Log("Policy for ", util.ParamsToString(req.FieldValues...), ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Policy for ", util.ParamsToString(req.FieldValues...), ": ", myRes, ", supposed to be ", res)
	}
}

func testGetGroupingPolicy(t *testing.T, e *testEngine, res [][]string) {
	t.Helper()
	req := &pb.EmptyRequest{Handler: e.h}
	reply, err := e.s.GetGroupingPolicy(e.ctx, req)
	assert.NoError(t, err)

	myRes := extractFromArray2DReply(reply)
	t.Log("Grouping policy: ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy: ", myRes, ", supposed to be ", res)
	}
}

func testGetFilteredGroupingPolicy(t *testing.T, e *testEngine, fieldIndex int, res [][]string, fieldValues ...string) {
	t.Helper()
	req := &pb.FilteredPolicyRequest{
		EnforcerHandler: e.h, FieldIndex: int32(fieldIndex), FieldValues: fieldValues}
	reply, err := e.s.GetFilteredGroupingPolicy(e.ctx, req)
	assert.NoError(t, err)

	myRes := extractFromArray2DReply(reply)
	t.Log("Grouping policy for ", util.ParamsToString(fieldValues...), ": ", myRes)

	if !util.Array2DEquals(res, myRes) {
		t.Error("Grouping policy for ", util.ParamsToString(fieldValues...), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasPolicy(t *testing.T, e *testEngine, policy []string, res bool) {
	t.Helper()
	req := &pb.PolicyRequest{EnforcerHandler: e.h, PType: "p", Params: policy}
	reply, err := e.s.HasPolicy(e.ctx, req)
	assert.NoError(t, err)

	myRes := reply.Res
	t.Log("Has policy ", util.ArrayToString(policy), ": ", myRes)

	if res != myRes {
		t.Error("Has policy ", util.ArrayToString(policy), ": ", myRes, ", supposed to be ", res)
	}
}

func testHasGroupingPolicy(t *testing.T, e *testEngine, policy []string, res bool) {
	t.Helper()
	req := &pb.PolicyRequest{EnforcerHandler: e.h, PType: "g", Params: policy}
	reply, err := e.s.HasNamedGroupingPolicy(e.ctx, req)
	assert.NoError(t, err)

	myRes := reply.Res
	t.Log("Has grouping policy ", util.ArrayToString(policy), ": ", myRes)

	if res != myRes {
		t.Error("Has grouping policy ", util.ArrayToString(policy), ": ", myRes, ", supposed to be ", res)
	}
}

func TestGetPolicyAPI(t *testing.T) {
	e := newTestEngine(t, "file", "../examples/rbac_policy.csv", "../examples/rbac_model.conf")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"data3_admin", "data3", "admin"},
		{"data4_admin", "data4", "read"}})

	testGetFilteredPolicy(t, e, 0, [][]string{{"alice", "data1", "read"}}, "alice")
	testGetFilteredPolicy(t, e, 0, [][]string{{"bob", "data2", "write"}}, "bob")
	testGetFilteredPolicy(t, e, 0, [][]string{{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}}, "data2_admin")
	testGetFilteredPolicy(t, e, 1, [][]string{{"alice", "data1", "read"}}, "data1")
	testGetFilteredPolicy(t, e, 1, [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}}, "data2")
	testGetFilteredPolicy(t, e, 2, [][]string{{"alice", "data1", "read"}, {"data2_admin", "data2", "read"}, {"data4_admin", "data4", "read"}}, "read")
	testGetFilteredPolicy(t, e, 2, [][]string{{"bob", "data2", "write"}, {"data2_admin", "data2", "write"}}, "write")

	testGetFilteredPolicy(t, e, 0, [][]string{{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"}}, "data2_admin", "data2")
	// Note: "" (empty string) in fieldValues means matching all values.
	testGetFilteredPolicy(t, e, 0, [][]string{{"data2_admin", "data2", "read"}}, "data2_admin", "", "read")
	testGetFilteredPolicy(t, e, 1, [][]string{{"bob", "data2", "write"},
		{"data2_admin", "data2", "write"}}, "data2", "write")

	testHasPolicy(t, e, []string{"alice", "data1", "read"}, true)
	testHasPolicy(t, e, []string{"bob", "data2", "write"}, true)
	testHasPolicy(t, e, []string{"alice", "data2", "read"}, false)
	testHasPolicy(t, e, []string{"bob", "data3", "write"}, false)

	testGetGroupingPolicy(t, e, [][]string{{"alice", "data2_admin"}, {"george", "data3_admin"}, {"data3_admin", "data4_admin"}})

	testGetFilteredGroupingPolicy(t, e, 0, [][]string{{"alice", "data2_admin"}}, "alice")
	testGetFilteredGroupingPolicy(t, e, 0, [][]string{}, "bob")
	testGetFilteredGroupingPolicy(t, e, 1, [][]string{}, "data1_admin")
	testGetFilteredGroupingPolicy(t, e, 1, [][]string{{"alice", "data2_admin"}}, "data2_admin")
	// Note: "" (empty string) in fieldValues means matching all values.
	testGetFilteredGroupingPolicy(t, e, 0, [][]string{{"alice", "data2_admin"}}, "", "data2_admin")

	testHasGroupingPolicy(t, e, []string{"alice", "data2_admin"}, true)
	testHasGroupingPolicy(t, e, []string{"bob", "data2_admin"}, false)
}

func TestModifyPolicyAPI(t *testing.T) {
	e := newTestEngine(t, "file", "../examples/rbac_policy.csv", "../examples/rbac_model.conf")

	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"data3_admin", "data3", "admin"},
		{"data4_admin", "data4", "read"}})

	_, err := e.s.RemovePolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"alice", "data1", "read"}})
	assert.NoError(t, err)
	_, err = e.s.RemovePolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"bob", "data2", "write"}})
	assert.NoError(t, err)
	_, err = e.s.RemovePolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"alice", "data1", "read"}})
	assert.NoError(t, err)

	_, err = e.s.AddPolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"eve", "data3", "read"}})
	assert.NoError(t, err)
	_, err = e.s.AddPolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"eve", "data3", "read"}})
	assert.NoError(t, err)

	namedPolicy := []string{"eve", "data3", "read"}
	_, err = e.s.RemovePolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, PType: "p", Params: namedPolicy})
	assert.NoError(t, err)
	_, err = e.s.AddNamedPolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, PType: "p", Params: namedPolicy})
	assert.NoError(t, err)

	testGetPolicy(t, e, [][]string{
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
		{"data3_admin", "data3", "admin"},
		{"data4_admin", "data4", "read"},
		{"eve", "data3", "read"}})

	_, err = e.s.RemoveFilteredPolicy(e.ctx, &pb.FilteredPolicyRequest{EnforcerHandler: e.h, FieldIndex: 1, FieldValues: []string{"data2"}})
	assert.NoError(t, err)

	_, err = e.s.RemoveFilteredPolicy(e.ctx, &pb.FilteredPolicyRequest{EnforcerHandler: e.h, FieldIndex: 1, FieldValues: []string{"data4"}})
	assert.NoError(t, err)

	testGetPolicy(t, e, [][]string{{"data3_admin", "data3", "admin"}, {"eve", "data3", "read"}})
}

func TestModifyGroupingPolicyAPI(t *testing.T) {
	e := newTestEngine(t, "file", "../examples/rbac_policy.csv", "../examples/rbac_model.conf")

	testGetRoles(t, e, "alice", []string{"data2_admin"})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "eve", []string{})
	testGetRoles(t, e, "non_exist", []string{})

	_, err := e.s.RemoveGroupingPolicy(e.ctx,
		&pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"alice", "data2_admin"}})
	assert.NoError(t, err)

	_, err = e.s.AddGroupingPolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"bob", "data1_admin"}})
	assert.NoError(t, err)

	_, err = e.s.AddGroupingPolicy(e.ctx, &pb.PolicyRequest{EnforcerHandler: e.h, Params: []string{"eve", "data3_admin"}})
	assert.NoError(t, err)

	namedGroupingPolicy := []string{"alice", "data2_admin"}
	testGetRoles(t, e, "alice", []string{})

	_, err = e.s.AddNamedGroupingPolicy(e.ctx,
		&pb.PolicyRequest{EnforcerHandler: e.h, PType: "g", Params: namedGroupingPolicy})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{"data2_admin"})

	_, err = e.s.RemoveNamedGroupingPolicy(e.ctx,
		&pb.PolicyRequest{EnforcerHandler: e.h, PType: "g", Params: namedGroupingPolicy})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{"data1_admin"})
	testGetRoles(t, e, "eve", []string{"data3_admin"})
	testGetRoles(t, e, "non_exist", []string{})

	testGetUsers(t, e, "data1_admin", []string{"bob"})
	testGetUsers(t, e, "data2_admin", []string{})
	testGetUsers(t, e, "data3_admin", []string{"eve", "george"})
	testGetUsers(t, e, "data4_admin", []string{"data3_admin"})

	_, err = e.s.RemoveFilteredGroupingPolicy(e.ctx,
		&pb.FilteredPolicyRequest{EnforcerHandler: e.h, FieldIndex: 0, FieldValues: []string{"bob"}})
	assert.NoError(t, err)

	testGetRoles(t, e, "alice", []string{})
	testGetRoles(t, e, "bob", []string{})
	testGetRoles(t, e, "eve", []string{"data3_admin"})
	testGetRoles(t, e, "non_exist", []string{})

	testGetUsers(t, e, "data1_admin", []string{})
	testGetUsers(t, e, "data2_admin", []string{})
	testGetUsers(t, e, "data3_admin", []string{"george", "eve"})
}
