package service

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/RafaySystems/rcloud-base/internal/dao"
	"github.com/RafaySystems/rcloud-base/internal/models"
	providers "github.com/RafaySystems/rcloud-base/internal/provider/kratos"
	"github.com/RafaySystems/rcloud-base/pkg/common"
	userrpcv3 "github.com/RafaySystems/rcloud-base/proto/rpc/user"
	authzv1 "github.com/RafaySystems/rcloud-base/proto/types/authz"
	v3 "github.com/RafaySystems/rcloud-base/proto/types/commonpb/v3"
	userv3 "github.com/RafaySystems/rcloud-base/proto/types/userpb/v3"
)

const (
	userKind     = "User"
	userListKind = "UserList"
)

// GroupService is the interface for group operations
type UserService interface {
	// create user
	Create(context.Context, *userv3.User) (*userv3.User, error)
	// get user by id
	GetByID(context.Context, *userv3.User) (*userv3.User, error)
	// // get user by name
	GetByName(context.Context, *userv3.User) (*userv3.User, error)
	// create or update user
	Update(context.Context, *userv3.User) (*userv3.User, error)
	// delete user
	Delete(context.Context, *userv3.User) (*userrpcv3.DeleteUserResponse, error)
	// list users
	List(context.Context, *userv3.User) (*userv3.UserList, error)
	// retrieve the cli config for the logged in user
	RetrieveCliConfig(ctx context.Context, req *userrpcv3.ApiKeyRequest) (*common.CliConfigDownloadData, error)
}

type userService struct {
	ap  providers.AuthProvider
	db  *bun.DB
	azc AuthzService
	ks  ApiKeyService
	cc  common.CliConfigDownloadData
}

type userTraits struct {
	Email       string
	FirstName   string
	LastName    string
	Description string
}

// FIXME: find a better way to do this
type parsedIds struct {
	Id           uuid.UUID
	Partner      uuid.UUID
	Organization uuid.UUID
}

func NewUserService(ap providers.AuthProvider, db *bun.DB, azc AuthzService, kss ApiKeyService, cfg common.CliConfigDownloadData) UserService {
	return &userService{ap: ap, db: db, azc: azc, ks: kss, cc: cfg}
}

func getUserTraits(traits map[string]interface{}) userTraits {
	// FIXME: is there a better way to do this?
	// All of these should ideally be available as we have the identities schema, but just in case
	email, ok := traits["email"]
	if !ok {
		email = ""
	}
	fname, ok := traits["first_name"]
	if !ok {
		fname = ""
	}
	lname, ok := traits["last_name"]
	if !ok {
		lname = ""
	}
	desc, ok := traits["desc"]
	if !ok {
		desc = ""
	}
	return userTraits{
		Email:       email.(string),
		FirstName:   fname.(string),
		LastName:    lname.(string),
		Description: desc.(string),
	}
}

// Map roles to accounts
func (s *userService) createUserRoleRelations(ctx context.Context, db bun.IDB, user *userv3.User, ids parsedIds) (*userv3.User, error) {
	projectNamespaceRoles := user.GetSpec().GetProjectNamespaceRoles()

	// TODO: add transactions
	var panrs []models.ProjectAccountNamespaceRole
	var pars []models.ProjectAccountResourcerole
	var ars []models.AccountResourcerole
	var ps []*authzv1.Policy
	for _, pnr := range projectNamespaceRoles {
		role := pnr.GetRole()
		entity, err := dao.GetIdByName(ctx, db, role, &models.Role{})
		if err != nil {
			return user, fmt.Errorf("unable to find role '%v'", role)
		}
		var roleId uuid.UUID
		if rle, ok := entity.(*models.Role); ok {
			roleId = rle.ID
		} else {
			return user, fmt.Errorf("unable to find role '%v'", role)
		}

		project := pnr.GetProject()
		org := user.GetMetadata().GetOrganization()
		namespaceId := pnr.GetNamespace() // TODO: lookup id from name

		switch {
		case pnr.Namespace != nil:
			projectId, err := dao.GetProjectId(ctx, db, project)
			if err != nil {
				return user, fmt.Errorf("unable to find project '%v'", project)
			}
			panr := models.ProjectAccountNamespaceRole{
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				AccountId:      ids.Id,
				ProjectId:      projectId,
				NamespaceId:    namespaceId,
				Active:         true,
			}
			panrs = append(panrs, panr)

			ps = append(ps, &authzv1.Policy{
				Sub:  "u:" + user.GetMetadata().GetName(),
				Ns:   strconv.FormatInt(namespaceId, 10),
				Proj: project,
				Org:  org,
				Obj:  role,
			})
		case project != "":
			projectId, err := dao.GetProjectId(ctx, db, project)
			if err != nil {
				return user, fmt.Errorf("unable to find project '%v'", project)
			}
			par := models.ProjectAccountResourcerole{
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				Default:        true,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				AccountId:      ids.Id,
				ProjectId:      projectId,
				Active:         true,
			}
			pars = append(pars, par)

			ps = append(ps, &authzv1.Policy{
				Sub:  "u:" + user.GetMetadata().GetName(),
				Ns:   "*",
				Proj: project,
				Org:  org,
				Obj:  role,
			})
		default:
			ar := models.AccountResourcerole{
				CreatedAt:      time.Now(),
				ModifiedAt:     time.Now(),
				Trash:          false,
				Default:        true,
				RoleId:         roleId,
				PartnerId:      ids.Partner,
				OrganizationId: ids.Organization,
				AccountId:      ids.Id,
				Active:         true,
			}
			ars = append(ars, ar)

			ps = append(ps, &authzv1.Policy{
				Sub:  "u:" + user.GetMetadata().GetName(),
				Ns:   "*",
				Proj: "*",
				Org:  org,
				Obj:  role,
			})
		}
	}
	if len(panrs) > 0 {
		_, err := dao.Create(ctx, db, &panrs)
		if err != nil {
			return &userv3.User{}, err
		}
	}
	if len(pars) > 0 {
		_, err := dao.Create(ctx, db, &pars)
		if err != nil {
			return &userv3.User{}, err
		}
	}
	if len(ars) > 0 {
		_, err := dao.Create(ctx, db, &ars)
		if err != nil {
			return &userv3.User{}, err
		}
	}

	if len(ps) > 0 {
		success, err := s.azc.CreatePolicies(ctx, &authzv1.Policies{Policies: ps})
		if err != nil || !success.Res {
			return &userv3.User{}, fmt.Errorf("unable to create mapping in authz; %v", err)
		}
	}

	return user, nil
}

// FIXME: make this generic
func (s *userService) getPartnerOrganization(ctx context.Context, db bun.IDB, user *userv3.User) (uuid.UUID, uuid.UUID, error) {
	partner := user.GetMetadata().GetPartner()
	org := user.GetMetadata().GetOrganization()
	partnerId, err := dao.GetPartnerId(ctx, db, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := dao.GetOrganizationId(ctx, db, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func (s *userService) Create(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	// TODO: restrict endpoint to admin
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, user)
	if err != nil {
		return nil, fmt.Errorf("unable to get partner and org id")
	}

	// Kratos checks if the user is already available
	id, err := s.ap.Create(ctx, map[string]interface{}{
		"email":       user.GetMetadata().GetName(), // can be just username for API access
		"first_name":  user.GetSpec().GetFirstName(),
		"last_name":   user.GetSpec().GetLastName(),
		"description": user.GetMetadata().GetDescription(),
	})
	if err != nil {
		return &userv3.User{}, err
	}

	uid, _ := uuid.Parse(id)

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return &userv3.User{}, err
	}

	user, err = s.createUserRoleRelations(ctx, tx, user, parsedIds{Id: uid, Partner: partnerId, Organization: organizationId})
	if err != nil {
		tx.Rollback()
		return &userv3.User{}, err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		_log.Warn("unable to commit changes", err)
	}

	rl, err := s.ap.GetRecoveryLink(ctx, id)
	fmt.Println("Recovery link:", rl) // TODO: email the recovery link to the user
	if err != nil {
		return &userv3.User{}, err
	}

	return user, nil
}

func (s *userService) identitiesModelToUser(ctx context.Context, db bun.IDB, user *userv3.User, usr *models.KratosIdentities) (*userv3.User, error) {
	traits := getUserTraits(usr.Traits)
	groups, err := dao.GetGroups(ctx, db, usr.ID)
	if err != nil {
		return &userv3.User{}, err
	}
	groupNames := []string{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}

	labels := make(map[string]string)

	roles, err := dao.GetUserRoles(ctx, db, usr.ID)
	if err != nil {
		return &userv3.User{}, err
	}

	user.ApiVersion = apiVersion
	user.Kind = userKind
	user.Metadata = &v3.Metadata{
		Name:        traits.Email,
		Description: traits.Description,
		Labels:      labels,
		ModifiedAt:  timestamppb.New(usr.UpdatedAt),
	}
	user.Spec = &userv3.UserSpec{
		FirstName:             traits.FirstName,
		LastName:              traits.LastName,
		Groups:                groupNames,
		ProjectNamespaceRoles: roles,
	}

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	id := user.GetMetadata().GetId()
	uid, err := uuid.Parse(id)
	if err != nil {
		return &userv3.User{}, err
	}
	entity, err := dao.GetByID(ctx, s.db, uid, &models.KratosIdentities{})
	if err != nil {
		return &userv3.User{}, err
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		user, err := s.identitiesModelToUser(ctx, s.db, user, usr)
		if err != nil {
			return &userv3.User{}, err
		}

		return user, nil
	}
	return user, fmt.Errorf("unabele to fetch user '%v'", id)

}

func (s *userService) GetByName(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	name := user.GetMetadata().GetName()
	entity, err := dao.GetByTraits(ctx, s.db, name, &models.KratosIdentities{})
	if err != nil {
		return &userv3.User{}, err
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		user, err := s.identitiesModelToUser(ctx, s.db, user, usr)
		if err != nil {
			return &userv3.User{}, err
		}

		return user, nil
	}
	fmt.Println("user:", user)
	return user, nil
}

func (s *userService) deleteUserRoleRelations(ctx context.Context, db bun.IDB, userId uuid.UUID, user *userv3.User) error {
	err := dao.DeleteX(ctx, db, "account_id", userId, &models.AccountResourcerole{})
	if err != nil {
		return err
	}
	err = dao.DeleteX(ctx, db, "account_id", userId, &models.ProjectAccountResourcerole{})
	if err != nil {
		return err
	}
	err = dao.DeleteX(ctx, db, "account_id", userId, &models.ProjectAccountNamespaceRole{})
	if err != nil {
		return err
	}

	_, err = s.azc.DeletePolicies(ctx, &authzv1.Policy{Sub: "u:" + user.GetMetadata().GetName()})
	if err != nil {
		return fmt.Errorf("unable to delete user-role relations from authz; %v", err)
	}

	return nil
}

func (s *userService) Update(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	name := user.GetMetadata().GetName()
	entity, err := dao.GetIdByTraits(ctx, s.db, name, &models.KratosIdentities{})
	if err != nil {
		return &userv3.User{}, fmt.Errorf("no user found with name '%v'", name)
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		partnerId, organizationId, err := s.getPartnerOrganization(ctx, s.db, user)
		if err != nil {
			return nil, fmt.Errorf("unable to get partner and org id")
		}
		err = s.ap.Update(ctx, usr.ID.String(), map[string]interface{}{
			"email":       user.GetMetadata().GetName(),
			"first_name":  user.GetSpec().GetFirstName(),
			"last_name":   user.GetSpec().GetLastName(),
			"description": user.GetMetadata().GetDescription(),
		})
		if err != nil {
			return &userv3.User{}, err
		}

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &userv3.User{}, err
		}

		err = s.deleteUserRoleRelations(ctx, tx, usr.ID, user)
		if err != nil {
			tx.Rollback()
			return &userv3.User{}, err
		}

		user, err = s.createUserRoleRelations(ctx, tx, user, parsedIds{Id: usr.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			tx.Rollback()
			return &userv3.User{}, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}
		return user, nil
	}

	return &userv3.User{}, fmt.Errorf("unable to update user '%v'", name)

}

func (s *userService) Delete(ctx context.Context, user *userv3.User) (*userrpcv3.DeleteUserResponse, error) {
	name := user.GetMetadata().GetName()
	entity, err := dao.GetIdByTraits(ctx, s.db, name, &models.KratosIdentities{})
	if err != nil {
		return &userrpcv3.DeleteUserResponse{}, fmt.Errorf("no user founnd with username '%v'", name)
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {

		tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return &userrpcv3.DeleteUserResponse{}, err
		}

		err = s.deleteUserRoleRelations(ctx, s.db, usr.ID, user)
		if err != nil {
			tx.Rollback()
			return &userrpcv3.DeleteUserResponse{}, err
		}

		err = s.ap.Delete(ctx, usr.ID.String())
		if err != nil {
			tx.Rollback()
			return &userrpcv3.DeleteUserResponse{}, err
		}

		err = dao.DeleteX(ctx, s.db, "account_id", usr.ID, &models.GroupAccount{})
		if err != nil {
			tx.Rollback()
			return &userrpcv3.DeleteUserResponse{}, fmt.Errorf("unable to delete user; %v", err)
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			_log.Warn("unable to commit changes", err)
		}
		return &userrpcv3.DeleteUserResponse{}, nil
	}
	return &userrpcv3.DeleteUserResponse{}, fmt.Errorf("unable to delete user '%v'", user.Metadata.Name)

}

func (s *userService) List(ctx context.Context, _ *userv3.User) (*userv3.UserList, error) {
	var users []*userv3.User
	userList := &userv3.UserList{
		ApiVersion: apiVersion,
		Kind:       userListKind,
		Metadata: &v3.ListMetadata{
			Count: 0,
		},
	}
	var accs []models.KratosIdentities
	entities, err := dao.ListAll(ctx, s.db, &accs)
	if err != nil {
		return userList, err
	}
	if usrs, ok := entities.(*[]models.KratosIdentities); ok {
		for _, usr := range *usrs {
			user := &userv3.User{}
			user, err := s.identitiesModelToUser(ctx, s.db, user, &usr)
			if err != nil {
				return userList, err
			}
			users = append(users, user)
		}

		// update the list metadata and items response
		userList.Metadata = &v3.ListMetadata{
			Count: int64(len(users)),
		}
		userList.Items = users
	}

	return userList, nil
}

func (s *userService) RetrieveCliConfig(ctx context.Context, req *userrpcv3.ApiKeyRequest) (*common.CliConfigDownloadData, error) {
	// get the default project associated to this account
	ap, err := dao.GetDefaultAccountProject(ctx, s.db, uuid.MustParse(req.Id))
	if err != nil {
		return nil, err
	}
	// fetch the metadata information required to populate cli config
	var proj models.Project
	_, err = dao.GetByID(ctx, s.db, ap.ProjecttId, &proj)
	if err != nil {
		return nil, err
	}

	var org models.Organization
	_, err = dao.GetByID(ctx, s.db, ap.OrganizationId, &org)
	if err != nil {
		return nil, err
	}

	var part models.Partner
	_, err = dao.GetByID(ctx, s.db, ap.PartnerId, &part)
	if err != nil {
		return nil, err
	}

	// get the api key if exists, if not create a new one
	apikey, err := s.ks.Get(ctx, &userrpcv3.ApiKeyRequest{Username: req.Username})
	if err != nil {
		return nil, err
	}

	if apikey == nil {
		apikey, err = s.ks.Create(ctx, &userrpcv3.ApiKeyRequest{Username: req.Username, Id: req.Id})
		if err != nil {
			return nil, err
		}
	}

	cliConfig := &common.CliConfigDownloadData{
		Profile:      s.cc.Profile,
		RestEndpoint: s.cc.RestEndpoint,
		OpsEndpoint:  s.cc.OpsEndpoint,
		ApiKey:       apikey.Key,
		ApiSecret:    apikey.Secret,
		Project:      proj.Name,
		Organization: org.Name,
		Partner:      part.Name,
	}

	return cliConfig, nil

}
