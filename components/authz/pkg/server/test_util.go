package server

import (
	"context"
	"io/ioutil"
	"testing"

	pb "github.com/RafaySystems/rcloud-base/components/authz/proto/rpc"
)

type testEngine struct {
	s   *Server
	ctx context.Context
	h   int32
}

func newTestEngine(t *testing.T, from, connectStr string, modelLoc string) *testEngine {
	s := NewServer()
	ctx := context.Background()

	_, err := s.NewAdapter(ctx, &pb.NewAdapterRequest{DriverName: from, ConnectString: connectStr})
	if err != nil {
		t.Fatal(err)
	}

	modelText, err := ioutil.ReadFile(modelLoc)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := s.NewEnforcer(ctx, &pb.NewEnforcerRequest{ModelText: string(modelText), AdapterHandle: 0})
	if err != nil {
		t.Fatal(err)
	}

	return &testEngine{s: s, ctx: ctx, h: resp.Handler}
}
