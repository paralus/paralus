package service

import (
	"context"
	"fmt"
	"time"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/models"
	systemv3 "github.com/RafaySystems/rcloud-base/components/adminsrv/proto/types/systempb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
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
	Close() error
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
	dao pg.EntityDAO
}

// NewProjectService return new project service
func NewProjectService(db *bun.DB) ProjectService {
	return &projectService{
		dao: pg.NewEntityDAO(db),
	}
}

func (s *projectService) Create(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {

	partnerId, _ := uuid.Parse(project.GetMetadata().GetPartner())
	organizationId, _ := uuid.Parse(project.GetMetadata().GetOrganization())
	//convert v3 spec to internal models
	proj := models.Project{
		Name:           project.GetMetadata().GetName(),
		Description:    project.GetMetadata().GetDescription(),
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		OrganizationId: organizationId,
		PartnerId:      partnerId,
		Default:        project.GetSpec().GetDefault(),
	}
	entity, err := s.dao.Create(ctx, &proj)
	if err != nil {
		project.Status = &v3.Status{
			ConditionType:   "Create",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
		}
		return project, err
	}

	//update v3 spec
	if createdProject, ok := entity.(*models.Project); ok {
		project.Metadata.Id = createdProject.ID.String()
		project.Spec = &systemv3.ProjectSpec{
			Default: createdProject.Default,
		}
		if project.Status != nil {
			project.Status = &v3.Status{
				ConditionType:   "Create",
				ConditionStatus: v3.ConditionStatus_StatusOK,
				LastUpdated:     timestamppb.Now(),
			}
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
		project.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return project, err
	}
	entity, err := s.dao.GetByID(ctx, uid, &models.Project{})
	if err != nil {
		project.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return project, err
	}

	if proj, ok := entity.(*models.Project); ok {
		labels := make(map[string]string)
		labels["organization"] = proj.OrganizationId.String()

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
		project.Status = &v3.Status{
			LastUpdated:     timestamppb.Now(),
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
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

	entity, err := s.dao.GetByName(ctx, name, &models.Project{})
	if err != nil {
		project.Status = &v3.Status{
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return project, err
	}

	if proj, ok := entity.(*models.Project); ok {
		labels := make(map[string]string)
		labels["organization"] = proj.OrganizationId.String()

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
		project.Status = &v3.Status{
			LastUpdated:     timestamppb.Now(),
			ConditionType:   "Describe",
			ConditionStatus: v3.ConditionStatus_StatusOK,
		}

		return project, nil
	}
	return project, nil

}

func (s *projectService) Update(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {

	entity, err := s.dao.GetByName(ctx, project.Metadata.Name, &models.Project{})
	if err != nil {
		project.Status = &v3.Status{
			ConditionType:   "Update",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return project, err
	}

	if proj, ok := entity.(*models.Project); ok {
		//update project details
		proj.Description = project.Metadata.Description
		proj.Default = project.Spec.Default
		proj.ModifiedAt = time.Now()

		_, err = s.dao.Update(ctx, proj.ID, proj)
		if err != nil {
			project.Status = &v3.Status{
				ConditionType:   "Update",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				LastUpdated:     timestamppb.Now(),
				Reason:          err.Error(),
			}
			return project, err
		}

		//update spec and status
		project.Spec = &systemv3.ProjectSpec{
			Default: proj.Default,
		}
		project.Status = &v3.Status{
			ConditionType:   "Update",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
		}
	}

	return project, nil
}

func (s *projectService) Delete(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {
	entity, err := s.dao.GetByName(ctx, project.Metadata.Name, &models.Project{})
	if err != nil {
		project.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusFailed,
			LastUpdated:     timestamppb.Now(),
			Reason:          err.Error(),
		}
		return project, err
	}
	if proj, ok := entity.(*models.Project); ok {
		err = s.dao.Delete(ctx, proj.ID, proj)
		if err != nil {
			project.Status = &v3.Status{
				ConditionType:   "Delete",
				ConditionStatus: v3.ConditionStatus_StatusFailed,
				LastUpdated:     timestamppb.Now(),
				Reason:          err.Error(),
			}
			return project, err
		}
		//update v3 spec
		project.Metadata.Id = proj.ID.String()
		project.Metadata.Name = proj.Name
		project.Status = &v3.Status{
			ConditionType:   "Delete",
			ConditionStatus: v3.ConditionStatus_StatusOK,
			LastUpdated:     timestamppb.Now(),
		}
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
		_, err := s.dao.GetByName(ctx, project.Metadata.Organization, &org)
		if err != nil {
			return projectList, err
		}
		var part models.Partner
		_, err = s.dao.GetByName(ctx, project.Metadata.Partner, &part)
		if err != nil {
			return projectList, err
		}
		var projs []models.Project
		entities, err := s.dao.List(ctx, uuid.NullUUID{UUID: part.ID, Valid: true}, uuid.NullUUID{UUID: org.ID, Valid: true}, &projs)
		if err != nil {
			return projectList, err
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

func (s *projectService) Close() error {
	return s.dao.Close()
}
