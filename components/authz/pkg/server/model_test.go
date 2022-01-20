package server

import (
	"context"
	"io/ioutil"
	"testing"

	pb "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc"
	"github.com/stretchr/testify/assert"
)

func testEnforce(t *testing.T, e *testEngine, sub string, obj string, act string, res bool) {
	t.Helper()
	reply, err := e.s.Enforce(e.ctx, &pb.EnforceRequest{EnforcerHandler: e.h, Params: []string{sub, obj, act}})
	assert.NoError(t, err)

	if reply.Res != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, !res, res)
	} else {
		t.Logf("Enforce for %s, %s, %s : %v", sub, obj, act, reply.Res)
	}
}

func testEnforceWithoutUsers(t *testing.T, e *testEngine, obj string, act string, res bool) {
	t.Helper()
	reply, err := e.s.Enforce(e.ctx, &pb.EnforceRequest{EnforcerHandler: e.h, Params: []string{obj, act}})
	assert.NoError(t, err)

	if reply.Res != res {
		t.Errorf("%s, %s: %t, supposed to be %t", obj, act, !res, res)
	}
}

func TestRBACModel(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	_, err := s.NewAdapter(ctx, &pb.NewAdapterRequest{DriverName: "file", ConnectString: "../examples/rbac_policy.csv"})
	if err != nil {
		t.Error(err)
	}

	modelText, err := ioutil.ReadFile("../examples/rbac_model.conf")
	if err != nil {
		t.Error(err)
	}

	resp, err := s.NewEnforcer(ctx, &pb.NewEnforcerRequest{ModelText: string(modelText), AdapterHandle: 0})
	if err != nil {
		t.Error(err)
	}
	e := resp.Handler

	sub := "alice"
	obj := "data1"
	act := "read"
	res := true

	resp2, err := s.Enforce(ctx, &pb.EnforceRequest{EnforcerHandler: e, Params: []string{sub, obj, act}})
	if err != nil {
		t.Error(err)
	}
	myRes := resp2.Res

	if myRes != res {
		t.Errorf("%s, %s, %s: %t, supposed to be %t", sub, obj, act, myRes, res)
	}
}

func TestABACModel(t *testing.T) {
	s := NewServer()
	ctx := context.Background()

	modelText, err := ioutil.ReadFile("../examples/abac_model.conf")
	if err != nil {
		t.Error(err)
	}

	resp, err := s.NewEnforcer(ctx, &pb.NewEnforcerRequest{ModelText: string(modelText), AdapterHandle: -1})
	if err != nil {
		t.Error(err)
	}
	type ABACModel struct {
		Name  string
		Owner string
	}
	e := resp.Handler

	data1, _ := MakeABAC(ABACModel{Name: "data1", Owner: "alice"})
	data2, _ := MakeABAC(ABACModel{Name: "data2", Owner: "bob"})

	testModel(t, s, e, "alice", data1, "read", true)
	testModel(t, s, e, "alice", data1, "write", true)
	testModel(t, s, e, "alice", data2, "read", false)
	testModel(t, s, e, "alice", data2, "write", false)
	testModel(t, s, e, "bob", data1, "read", false)
	testModel(t, s, e, "bob", data1, "write", false)
	testModel(t, s, e, "bob", data2, "read", true)
	testModel(t, s, e, "bob", data2, "write", true)

}

func testModel(t *testing.T, s *Server, enforcerHandler int32, sub string, obj string, act string, res bool) {
	t.Helper()

	reply, err := s.Enforce(nil, &pb.EnforceRequest{EnforcerHandler: enforcerHandler, Params: []string{sub, obj, act}})
	assert.NoError(t, err)

	if reply.Res != res {
		t.Errorf("%s, %v, %s: %t, supposed to be %t", sub, obj, act, !res, res)
	}
}
