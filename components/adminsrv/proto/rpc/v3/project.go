package rpcv3

import (
	"context"

	"github.com/RafaySystems/rcloud-base/components/adminsrv/pkg/service"
	systempbv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
)

type projectServer struct {
	service.ProjectService
}

// NewProjectServer returns new project server implementation
func NewProjectServer(ps service.ProjectService) ProjectServer {
	return &projectServer{ps}
}

func (s *projectServer) CreateProject(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
	project, err := s.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectServer) GetProjects(ctx context.Context, p *systempbv3.Project) (*systempbv3.ProjectList, error) {
	projects, err := s.List(ctx, p)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *projectServer) GetProject(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {

	project, err := s.GetByName(ctx, p.Metadata.Name)
	if err != nil {
		project, err = s.GetByID(ctx, p.Metadata.Id)
		if err != nil {
			return nil, err
		}
	}

	return project, nil
}

func (s *projectServer) DeleteProject(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
	project, err := s.Delete(ctx, p)
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (s *projectServer) UpdateProject(ctx context.Context, p *systempbv3.Project) (*systempbv3.Project, error) {
	project, err := s.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return project, nil
}
