package server

import (
	"context"

	"github.com/RafaySystems/rcloud-base/pkg/service"
	systemrpc "github.com/RafaySystems/rcloud-base/proto/rpc/system"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	systempbv3 "github.com/RafaySystems/rcloud-base/proto/types/systempb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type projectServer struct {
	service.ProjectService
}

// NewProjectServer returns new project server implementation
func NewProjectServer(ps service.ProjectService) systemrpc.ProjectServer {
	return &projectServer{ps}
}

func updateProjectStatus(req *systempbv3.Project, resp *systempbv3.Project, err error) *systempbv3.Project {
	if err != nil {
		req.Status = &v3.Status{
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return req
	}
	resp.Status = &v3.Status{ConditionStatus: v3.ConditionStatus_StatusOK}
	return resp
}

func (s *projectServer) CreateProject(ctx context.Context, req *systempbv3.Project) (*systempbv3.Project, error) {
	resp, err := s.Create(ctx, req)
	return updateProjectStatus(req, resp, err), err
}

func (s *projectServer) GetProjects(ctx context.Context, req *systempbv3.Project) (*systempbv3.ProjectList, error) {
	return s.List(ctx, req)
}

func (s *projectServer) GetProject(ctx context.Context, req *systempbv3.Project) (*systempbv3.Project, error) {

	resp, err := s.GetByName(ctx, req.Metadata.Name)
	return updateProjectStatus(req, resp, err), err
}

func (s *projectServer) DeleteProject(ctx context.Context, req *systempbv3.Project) (*systempbv3.Project, error) {
	resp, err := s.Delete(ctx, req)
	return updateProjectStatus(req, resp, err), err
}

func (s *projectServer) UpdateProject(ctx context.Context, req *systempbv3.Project) (*systempbv3.Project, error) {
	resp, err := s.Update(ctx, req)
	return updateProjectStatus(req, resp, err), err
}
