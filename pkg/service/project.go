package service

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/internal/models"
	"github.com/RafaySystems/rcloud-base/internal/persistence/provider/pg"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	systemv3 "github.com/RafaySystems/rcloud-base/proto/types/systempb/v3"
	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	projectKind     = "Project"
	projectListKind = "ProjectList"
)

// ProjectService is the interface for project operations
type ProjectService interface {
	// create project
	Create(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error)
	// get project by id
	GetByID(ctx context.Context, id string) (*systemv3.Project, error)
	// get project by name
	GetByName(ctx context.Context, name string) (*systemv3.Project, error)
	// create or update project
	Update(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error)
	// delete project
	Delete(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error)
	// list projects
	List(ctx context.Context, project *systemv3.Project) (*systemv3.ProjectList, error)
	//TODO Associate project with groups, user, roles
}

// projectService implements ProjectService
type projectService struct {
	db *bun.DB
}

// NewProjectService return new project service
func NewProjectService(db *bun.DB) ProjectService {
	return &projectService{db}
}

func (s *projectService) Create(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {

	if project.Metadata.Organization == "" {
		return nil, fmt.Errorf("missing organization in metadata")
	}

	var org models.Organization
	_, err := pg.GetByName(ctx, s.db, project.Metadata.Organization, &org)
	if err != nil {
		return nil, err
	}

	//convert v3 spec to internal models
	proj := models.Project{
		Name:           project.GetMetadata().GetName(),
		Description:    project.GetMetadata().GetDescription(),
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		OrganizationId: org.ID,
		PartnerId:      org.PartnerId,
		Default:        project.GetSpec().GetDefault(),
	}
	entity, err := pg.Create(ctx, s.db, &proj)
	if err != nil {
		return &systemv3.Project{}, err
	}

	//update v3 spec
	if createdProject, ok := entity.(*models.Project); ok {
		project.Metadata.Id = createdProject.ID.String()
		project.Spec = &systemv3.ProjectSpec{
			Default: createdProject.Default,
		}
	}

	return project, nil

}

func (s *projectService) GetByID(ctx context.Context, id string) (*systemv3.Project, error) {

	project := &systemv3.Project{
		ApiVersion: apiVersion,
		Kind:       projectKind,
		Metadata: &v3.Metadata{
			Id: id,
		},
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return &systemv3.Project{}, err
	}
	entity, err := pg.GetByID(ctx, s.db, uid, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	if proj, ok := entity.(*models.Project); ok {

		project.Metadata = &v3.Metadata{
			Name:         proj.Name,
			Description:  proj.Description,
			Id:           proj.ID.String(),
			Organization: proj.OrganizationId.String(),
			Partner:      proj.PartnerId.String(),
			ModifiedAt:   timestamppb.New(proj.ModifiedAt),
		}
		project.Spec = &systemv3.ProjectSpec{
			Default: proj.Default,
		}

		return project, nil
	}
	return project, nil

}

func (s *projectService) GetByName(ctx context.Context, name string) (*systemv3.Project, error) {

	project := &systemv3.Project{
		ApiVersion: apiVersion,
		Kind:       projectKind,
		Metadata: &v3.Metadata{
			Name: name,
		},
	}

	entity, err := pg.GetByName(ctx, s.db, name, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	if proj, ok := entity.(*models.Project); ok {

		var org models.Organization
		_, err := pg.GetByID(ctx, s.db, proj.OrganizationId, &org)
		if err != nil {
			return nil, err
		}

		var partner models.Partner
		_, err = pg.GetByID(ctx, s.db, proj.PartnerId, &partner)
		if err != nil {
			return nil, err
		}

		project.Metadata = &v3.Metadata{
			Name:         proj.Name,
			Description:  proj.Description,
			Organization: org.Name,
			Partner:      partner.Name,
			ModifiedAt:   timestamppb.New(proj.ModifiedAt),
		}
		project.Spec = &systemv3.ProjectSpec{
			Default: proj.Default,
		}

		return project, nil
	}
	return project, nil

}

func (s *projectService) Update(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {

	entity, err := pg.GetByName(ctx, s.db, project.Metadata.Name, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	if proj, ok := entity.(*models.Project); ok {
		//update project details
		proj.Description = project.Metadata.Description
		proj.Default = project.Spec.Default
		proj.ModifiedAt = time.Now()

		_, err = pg.Update(ctx, s.db, proj.ID, proj)
		if err != nil {
			return &systemv3.Project{}, err
		}

		//update spec and status
		project.Spec = &systemv3.ProjectSpec{
			Default: proj.Default,
		}
	}

	return project, nil
}

func (s *projectService) Delete(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {
	entity, err := pg.GetByName(ctx, s.db, project.Metadata.Name, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}
	if proj, ok := entity.(*models.Project); ok {
		err := pg.Delete(ctx, s.db, proj.ID, proj)
		if err != nil {
			return &systemv3.Project{}, err
		}
		//update v3 spec
		project.Metadata.Id = proj.ID.String()
		project.Metadata.Name = proj.Name
	}

	return project, nil
}

func (s *projectService) List(ctx context.Context, project *systemv3.Project) (*systemv3.ProjectList, error) {

	var projects []*systemv3.Project
	projectList := &systemv3.ProjectList{
		ApiVersion: apiVersion,
		Kind:       projectListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}
	if len(project.Metadata.Organization) > 0 {
		var org models.Organization
		_, err := pg.GetByName(ctx, s.db, project.Metadata.Organization, &org)
		if err != nil {
			return &systemv3.ProjectList{}, err
		}
		var part models.Partner
		_, err = pg.GetByName(ctx, s.db, project.Metadata.Partner, &part)
		if err != nil {
			return &systemv3.ProjectList{}, err
		}
		var projs []models.Project
		entities, err := pg.List(ctx, s.db, uuid.NullUUID{UUID: part.ID, Valid: true}, uuid.NullUUID{UUID: org.ID, Valid: true}, &projs)
		if err != nil {
			return &systemv3.ProjectList{}, err
		}
		if projs, ok := entities.(*[]models.Project); ok {
			for _, proj := range *projs {
				labels := make(map[string]string)
				labels["organization"] = proj.OrganizationId.String()
				labels["partner"] = proj.PartnerId.String()

				project.Metadata = &v3.Metadata{
					Name:         proj.Name,
					Description:  proj.Description,
					Id:           proj.ID.String(),
					Organization: proj.OrganizationId.String(),
					Partner:      proj.PartnerId.String(),
					Labels:       labels,
					ModifiedAt:   timestamppb.New(proj.ModifiedAt),
				}
				project.Spec = &systemv3.ProjectSpec{
					Default: proj.Default,
				}
				projects = append(projects, project)
			}

			//update the list metadata and items response
			projectList.Metadata = &v3.ListMetadata{
				Count: int64(len(projects)),
			}
			projectList.Items = projects
		}

	} else {
		return projectList, fmt.Errorf("missing organization id in metadata")
	}
	return projectList, nil
}
