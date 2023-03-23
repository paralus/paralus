package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	cdao "github.com/paralus/paralus/internal/cluster/dao"
	"github.com/paralus/paralus/internal/dao"
	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/common"
	authzv1 "github.com/paralus/paralus/proto/types/authz"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	systemv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	bun "github.com/uptrace/bun"
	"go.uber.org/zap"
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
	db  *bun.DB
	azc AuthzService
	al  *zap.Logger
	dev bool
}

// NewProjectService return new project service
func NewProjectService(db *bun.DB, azc AuthzService, al *zap.Logger, dev bool) ProjectService {
	return &projectService{db: db, azc: azc, al: al, dev: dev}
}

func (s *projectService) Create(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {

	if project.Metadata.Organization == "" {
		return nil, fmt.Errorf("missing organization in metadata")
	}

	matched := common.PrjNameRX.MatchString(project.Metadata.GetName())
	if !matched {
		return nil, errors.New("project name contains invalid characters. Valid characters are alphanumeric and hyphen, except at the beginning or the end")
	}

	var org models.Organization
	_, err := dao.GetByName(ctx, s.db, project.Metadata.Organization, &org)
	if err != nil {
		return nil, err
	}

	p, _ := dao.GetIdByNamePartnerOrg(ctx, s.db, project.GetMetadata().GetName(), uuid.NullUUID{}, uuid.NullUUID{}, &models.Project{})
	if p != nil {
		return nil, fmt.Errorf("project '%v' already exists", project.GetMetadata().GetName())
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

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	entity, err := dao.Create(ctx, tx, &proj)
	if err != nil {
		tx.Rollback()
		return &systemv3.Project{}, err
	}

	//update v3 spec
	if createdProject, ok := entity.(*models.Project); ok {

		project, err = s.createGroupRoleRelations(ctx, tx, project, parsedIds{Id: createdProject.ID, Partner: createdProject.PartnerId, Organization: createdProject.OrganizationId})
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		project, err = s.createProjectAccountRelations(ctx, tx, createdProject.ID, project)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		project.Metadata.Id = createdProject.ID.String()
		project.Spec = &systemv3.ProjectSpec{
			Default: createdProject.Default,
		}

		CreateProjectAuditEvent(ctx, s.al, AuditActionCreate, project.GetMetadata().GetName(), createdProject.ID)
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		_log.Warn("unable to commit changes", err)
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
	entity, err := dao.GetByID(ctx, s.db, uid, &models.Project{})
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

	entity, err := dao.GetByName(ctx, s.db, name, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	if proj, ok := entity.(*models.Project); ok {

		var org models.Organization
		_, err := dao.GetByID(ctx, s.db, proj.OrganizationId, &org)
		if err != nil {
			return nil, err
		}

		var partner models.Partner
		_, err = dao.GetByID(ctx, s.db, proj.PartnerId, &partner)
		if err != nil {
			return nil, err
		}

		pnr, err := dao.GetProjectGroupRoles(ctx, s.db, proj.ID)
		if err != nil {
			return nil, err
		}

		ur, err := dao.GetProjectUserRoles(ctx, s.db, proj.ID)
		if err != nil {
			return nil, err
		}

		project.Metadata = &v3.Metadata{
			Name:         proj.Name,
			Description:  proj.Description,
			Id:           proj.ID.String(),
			Organization: org.Name,
			Partner:      partner.Name,
			ModifiedAt:   timestamppb.New(proj.ModifiedAt),
		}
		project.Spec = &systemv3.ProjectSpec{
			Default:               proj.Default,
			ProjectNamespaceRoles: pnr,
			UserRoles:             ur,
		}

		return project, nil
	}
	return project, nil

}

func (s *projectService) Update(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {

	entity, err := dao.GetByName(ctx, s.db, project.Metadata.Name, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	if proj, ok := entity.(*models.Project); ok {

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &systemv3.Project{}, err
		}

		//update project details
		proj.Description = project.Metadata.Description
		proj.Default = project.Spec.Default
		proj.ModifiedAt = time.Now()

		project, err = s.deleteGroupRoleRelations(ctx, tx, proj.ID, project)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}
		project, err = s.createGroupRoleRelations(ctx, tx, project, parsedIds{Id: proj.ID, Partner: proj.PartnerId, Organization: proj.OrganizationId})
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		project, err = s.deleteProjectAccountRelations(ctx, tx, proj.ID, project)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}
		project, err = s.createProjectAccountRelations(ctx, tx, proj.ID, project)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		_, err = dao.Update(ctx, tx, proj.ID, proj)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		pnr, err := dao.GetProjectGroupRoles(ctx, tx, proj.ID)
		if err != nil {
			return nil, err
		}

		ur, err := dao.GetProjectUserRoles(ctx, tx, proj.ID)
		if err != nil {
			return nil, err
		}

		//update spec and status
		project.Spec = &systemv3.ProjectSpec{
			Default:               proj.Default,
			ProjectNamespaceRoles: pnr,
			UserRoles:             ur,
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}

		CreateProjectAuditEvent(ctx, s.al, AuditActionUpdate, project.GetMetadata().GetName(), proj.ID)
	}

	return project, nil
}

func (s *projectService) Delete(ctx context.Context, project *systemv3.Project) (*systemv3.Project, error) {
	entity, err := dao.GetByName(ctx, s.db, project.Metadata.Name, &models.Project{})
	if err != nil {
		return &systemv3.Project{}, err
	}
	if proj, ok := entity.(*models.Project); ok {

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &systemv3.Project{}, err
		}

		clusters, err := cdao.ListClusters(ctx, s.db, commonv3.QueryOptions{
			Project:      proj.ID.String(),
			Organization: proj.OrganizationId.String(),
			Partner:      proj.PartnerId.String(),
		})
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}
		if len(clusters) > 0 {
			tx.Rollback()
			return &systemv3.Project{}, fmt.Errorf("there is(are) active cluster(s) %d in the project %s", len(clusters), proj.Name)
		}

		project, err = s.deleteGroupRoleRelations(ctx, tx, proj.ID, project)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		project, err = s.deleteProjectAccountRelations(ctx, tx, proj.ID, project)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}

		err = dao.Delete(ctx, tx, proj.ID, proj)
		if err != nil {
			tx.Rollback()
			return &systemv3.Project{}, err
		}
		//update v3 spec
		project.Metadata.Id = proj.ID.String()
		project.Metadata.Name = proj.Name

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
			return &systemv3.Project{}, err
		}

		CreateProjectAuditEvent(ctx, s.al, AuditActionDelete, project.GetMetadata().GetName(), proj.ID)
	}

	return project, nil
}

func (s *projectService) List(ctx context.Context, project *systemv3.Project) (*systemv3.ProjectList, error) {

	username := ""
	if !s.dev {
		sd, ok := GetSessionDataFromContext(ctx)
		if !ok {
			return &systemv3.ProjectList{}, fmt.Errorf("cannot perform project listing without auth")
		}
		username = sd.Username
	}

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
		_, err := dao.GetByName(ctx, s.db, project.Metadata.Organization, &org)
		if err != nil {
			return &systemv3.ProjectList{}, err
		}
		var part models.Partner
		_, err = dao.GetByName(ctx, s.db, project.Metadata.Partner, &part)
		if err != nil {
			return &systemv3.ProjectList{}, err
		}

		var projs []models.Project
		if !s.dev {
			entity, err := dao.GetUserByEmail(ctx, s.db, username, &models.KratosIdentities{})
			if err != nil {
				return &systemv3.ProjectList{}, err
			}

			if usr, ok := entity.(*models.KratosIdentities); ok {
				projs, err = dao.GetFileteredProjects(ctx, s.db, usr.ID, part.ID, org.ID)
				if err != nil {
					return &systemv3.ProjectList{}, err
				}
			}

		} else {
			_, err = dao.List(ctx, s.db, uuid.NullUUID{UUID: part.ID, Valid: true}, uuid.NullUUID{UUID: org.ID, Valid: true}, &projs)
			if err != nil {
				return &systemv3.ProjectList{}, err
			}
		}

		for _, proj := range projs {
			labels := make(map[string]string)
			labels["organization"] = proj.OrganizationId.String()
			labels["partner"] = proj.PartnerId.String()

			pnr, err := dao.GetProjectGroupRoles(ctx, s.db, proj.ID)
			if err != nil {
				return nil, err
			}
			ur, err := dao.GetProjectUserRoles(ctx, s.db, proj.ID)
			if err != nil {
				return nil, err
			}
			project := &systemv3.Project{
				Metadata: &v3.Metadata{
					Name:         proj.Name,
					Description:  proj.Description,
					Id:           proj.ID.String(),
					Organization: proj.OrganizationId.String(),
					Partner:      proj.PartnerId.String(),
					Labels:       labels,
					CreatedAt:    timestamppb.New(proj.CreatedAt),
					ModifiedAt:   timestamppb.New(proj.ModifiedAt),
				},
				Spec: &systemv3.ProjectSpec{
					Default:               proj.Default,
					ProjectNamespaceRoles: pnr,
					UserRoles:             ur,
				},
			}
			projects = append(projects, project)
		}

		//update the list metadata and items response
		projectList.Metadata = &v3.ListMetadata{
			Count: int64(len(projects)),
		}
		projectList.Items = projects
		return projectList, nil

	}
	return projectList, fmt.Errorf("missing organization id in metadata")
}

// Map roles to groups
func (s *projectService) createGroupRoleRelations(ctx context.Context, db bun.IDB, project *systemv3.Project, ids parsedIds) (*systemv3.Project, error) {
	projectNamespaceRoles := project.GetSpec().GetProjectNamespaceRoles()

	var pgrs []models.ProjectGroupRole
	var pgnr []models.ProjectGroupNamespaceRole
	var ps []*authzv1.Policy
	for _, pnr := range projectNamespaceRoles {
		role := pnr.GetRole()
		entity, err := dao.GetByName(ctx, db, role, &models.Role{})
		if err != nil {
			return &systemv3.Project{}, fmt.Errorf("unable to find role '%v'", role)
		}
		var roleId uuid.UUID
		var scope string
		var roleName string
		if rle, ok := entity.(*models.Role); ok {
			roleId = rle.ID
			scope = rle.Scope
			roleName = rle.Name
		} else {
			return &systemv3.Project{}, fmt.Errorf("unable to find role '%v'", role)
		}

		grp := pnr.Group
		entity, err = dao.GetIdByName(ctx, s.db, *grp, &models.Group{})
		if err != nil {
			return &systemv3.Project{}, fmt.Errorf("unable to find group '%v'", grp)
		}
		var grpId uuid.UUID
		var grpName string
		if g, ok := entity.(*models.Group); ok {
			grpId = g.ID
			grpName = g.Name
		} else {
			return &systemv3.Project{}, fmt.Errorf("unable to find group '%v'", grp)
		}

		org := project.Metadata.Organization
		switch scope {
		case "project":
			if org == "" {
				return &systemv3.Project{}, fmt.Errorf("no org name provided for role '%v'", roleName)
			}

			pgr := models.ProjectGroupRole{
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				GroupId:        grpId,
				ProjectId:      ids.Id,
				Active:         true,
			}
			pgrs = append(pgrs, pgr)

			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + grpName,
				Ns:   "*",
				Proj: project.Metadata.Name,
				Org:  org,
				Obj:  role,
			})
		case "namespace":
			if org == "" {
				return &systemv3.Project{}, fmt.Errorf("no org name provided for role '%v'", roleName)
			}

			namespace := pnr.GetNamespace()
			pgnrObj := models.ProjectGroupNamespaceRole{
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				RoleId:         roleId,
				GroupId:        grpId,
				ProjectId:      ids.Id,
				Namespace:      namespace,
				Active:         true,
			}
			pgnr = append(pgnr, pgnrObj)

			ps = append(ps, &authzv1.Policy{
				Sub:  "g:" + grpName,
				Ns:   namespace,
				Proj: project.Metadata.Name,
				Org:  org,
				Obj:  role,
			})
		default:
			if err != nil {
				return project, fmt.Errorf("other scoped roles are not handled")
			}
		}

	}
	if len(pgrs) > 0 {
		_, err := dao.Create(ctx, db, &pgrs)
		if err != nil {
			return &systemv3.Project{}, err
		}
	}
	if len(pgnr) > 0 {
		_, err := dao.Create(ctx, db, &pgnr)
		if err != nil {
			return &systemv3.Project{}, err
		}
	}

	if len(ps) > 0 {
		success, err := s.azc.CreatePolicies(ctx, &authzv1.Policies{Policies: ps})
		if err != nil || !success.Res {
			return &systemv3.Project{}, fmt.Errorf("unable to create mapping in authz; %v", err)
		}
	}

	return project, nil
}

func (s *projectService) deleteGroupRoleRelations(ctx context.Context, db bun.IDB, projectId uuid.UUID, project *systemv3.Project) (*systemv3.Project, error) {
	// delete previous entries
	err := dao.DeleteX(ctx, db, "project_id", projectId, &models.ProjectGroupRole{})
	if err != nil {
		return &systemv3.Project{}, err
	}

	pgnr := []models.ProjectGroupNamespaceRole{}
	err = dao.DeleteXR(ctx, db, "project_id", projectId, &pgnr)
	if err != nil {
		return &systemv3.Project{}, err
	}

	_, err = s.azc.DeletePolicies(ctx, &authzv1.Policy{Proj: project.GetMetadata().GetName()})
	if err != nil {
		return &systemv3.Project{}, fmt.Errorf("unable to delete project group-role relations from authz; %v", err)
	}
	return project, nil
}

func (s *projectService) deleteProjectAccountRelations(ctx context.Context, db bun.IDB, projectId uuid.UUID, project *systemv3.Project) (*systemv3.Project, error) {
	err := dao.DeleteX(ctx, db, "project_id", projectId, &models.ProjectAccountResourcerole{})
	if err != nil {
		return &systemv3.Project{}, fmt.Errorf("unable to delete project; %v", err)
	}

	pgnr := []models.ProjectAccountNamespaceRole{}
	err = dao.DeleteXR(ctx, db, "project_id", projectId, &pgnr)
	if err != nil {
		return &systemv3.Project{}, err
	}

	_, err = s.azc.DeletePolicies(ctx, &authzv1.Policy{Proj: project.GetMetadata().GetName()})
	if err != nil {
		return &systemv3.Project{}, fmt.Errorf("unable to delete project user-role relations from authz; %v", err)
	}
	return project, nil
}

// Update the users(account) mapped to each project
func (s *projectService) createProjectAccountRelations(ctx context.Context, db bun.IDB, projectId uuid.UUID, project *systemv3.Project) (*systemv3.Project, error) {
	var parrs []models.ProjectAccountResourcerole
	var panrs []models.ProjectAccountNamespaceRole
	var ugs []*authzv1.Policy

	for _, ur := range project.GetSpec().GetUserRoles() {
		// FIXME: do combined lookup
		entity, err := dao.GetUserIdByEmail(ctx, db, ur.User, &models.KratosIdentities{})
		if err != nil {
			return &systemv3.Project{}, fmt.Errorf("unable to find user '%v'", ur.User)
		}
		rentity, err := dao.GetByName(ctx, db, ur.Role, &models.Role{})
		if err != nil {
			return &systemv3.Project{}, fmt.Errorf("unable to find user '%v'", ur.User)
		}

		if acc, ok := entity.(*models.KratosIdentities); ok {
			if role, ok := rentity.(*models.Role); ok {
				switch role.Scope {
				case "project":
					parr := models.ProjectAccountResourcerole{
						CreatedAt:      time.Now(),
						ModifiedAt:     time.Now(),
						Trash:          false,
						AccountId:      acc.ID,
						ProjectId:      projectId,
						RoleId:         role.ID,
						OrganizationId: role.OrganizationId,
						PartnerId:      role.PartnerId,
						Active:         true,
					}
					parrs = append(parrs, parr)
					ugs = append(ugs, &authzv1.Policy{
						Sub:  "u:" + ur.User,
						Proj: project.Metadata.Name,
						Org:  project.Metadata.Organization,
						Ns:   "*",
						Obj:  role.Name,
					})
				case "namespace":
					panrObj := models.ProjectAccountNamespaceRole{
						CreatedAt:      time.Now(),
						ModifiedAt:     time.Now(),
						Trash:          false,
						AccountId:      acc.ID,
						PartnerId:      role.PartnerId,
						OrganizationId: role.OrganizationId,
						RoleId:         role.ID,
						ProjectId:      projectId,
						Namespace:      ur.GetNamespace(),
						Active:         true,
					}
					panrs = append(panrs, panrObj)

					ugs = append(ugs, &authzv1.Policy{
						Sub:  "u:" + ur.User,
						Proj: project.Metadata.Name,
						Org:  project.Metadata.Organization,
						Ns:   ur.GetNamespace(),
						Obj:  role.Name,
					})
				default:
					if err != nil {
						return project, fmt.Errorf("other scoped roles are not handled")
					}
				}
			}
		}
	}
	if len(parrs) == 0 && len(panrs) == 0 {
		return project, nil
	}
	if len(parrs) > 0 {
		_, err := dao.Create(ctx, db, &parrs)
		if err != nil {
			return &systemv3.Project{}, err
		}
	}
	if len(panrs) > 0 {
		_, err := dao.Create(ctx, db, &panrs)
		if err != nil {
			return &systemv3.Project{}, err
		}
	}

	// TODO: revert our db inserts if this fails
	// Just FYI, the succcess can be false if we delete the db directly but casbin has it available internally
	_, err := s.azc.CreatePolicies(ctx, &authzv1.Policies{Policies: ugs})
	if err != nil {
		return &systemv3.Project{}, fmt.Errorf("unable to create mapping in authz; %v", err)
	}

	return project, nil
}
