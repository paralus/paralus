package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	bun "github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/utils"
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/models"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/internal/user/dao"
	"github.com/RafaySystems/rcloud-base/components/usermgmt/pkg/providers"
	userrpcv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/rpc/v3"
	userv3 "github.com/RafaySystems/rcloud-base/components/usermgmt/proto/types/userpb/v3"
)

const (
	userKind     = "User"
	userListKind = "UserList"
)

// GroupService is the interface for group operations
type UserService interface {
	Close() error
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
}

type userService struct {
	ap   providers.AuthProvider
	dao  pg.EntityDAO
	udao dao.UserDAO
	l    utils.Lookup
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

func NewUserService(ap providers.AuthProvider, db *bun.DB) UserService {
	return &userService{ap: ap, dao: pg.NewEntityDAO(db), udao: dao.NewUserDAO(db), l: utils.NewLookup(db)}
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
func (s *userService) updateUserRoleRelation(ctx context.Context, user *userv3.User, ids parsedIds) (*userv3.User, error) {
	projectNamespaceRoles := user.GetSpec().GetProjectNamespaceRoles()

	// TODO: add transactions
	var panrs []models.ProjectAccountNamespaceRole
	var pars []models.ProjectAccountResourcerole
	var ars []models.AccountResourcerole
	for _, pnr := range projectNamespaceRoles {
		role := pnr.GetRole()
		entity, err := s.dao.GetIdByName(ctx, role, &models.Role{})
		if err != nil {
			user.Status = statusFailed(fmt.Errorf("unable to find role '%v'", role))
			return user, err
		}
		var roleId uuid.UUID
		if rle, ok := entity.(*models.Role); ok {
			roleId = rle.ID
		} else {
			user.Status = statusFailed(fmt.Errorf("unable to find role '%v'", role))
			return user, err
		}

		project := pnr.GetProject()
		namespaceId := pnr.GetNamespace() // TODO: lookup id from name

		switch {
		case pnr.Namespace != nil:
			projectId, err := s.l.GetProjectId(ctx, project)
			if err != nil {
				user.Status = statusFailed(fmt.Errorf("unable to find project '%v'", project))
				return user, err
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
				NamesapceId:    namespaceId,
				Active:         true,
			}
			panrs = append(panrs, panr)
		case project != "":
			projectId, err := s.l.GetProjectId(ctx, project)
			if err != nil {
				user.Status = statusFailed(fmt.Errorf("unable to find project '%v'", project))
				return user, err
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
		}
	}
	if len(panrs) > 0 {
		_, err := s.dao.Create(ctx, &panrs)
		if err != nil {
			return user, err
		}
	}
	if len(pars) > 0 {
		_, err := s.dao.Create(ctx, &pars)
		if err != nil {
			return user, err
		}
	}
	if len(ars) > 0 {
		_, err := s.dao.Create(ctx, &ars)
		if err != nil {
			return user, err
		}
	}

	return user, nil
}

// FIXME: make this generic
func (s *userService) getPartnerOrganization(ctx context.Context, user *userv3.User) (uuid.UUID, uuid.UUID, error) {
	partner := user.GetMetadata().GetPartner()
	org := user.GetMetadata().GetOrganization()
	partnerId, err := s.l.GetPartnerId(ctx, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := s.l.GetOrganizationId(ctx, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func (s *userService) Create(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	// TODO: restrict endpoint to admin
	partnerId, organizationId, err := s.getPartnerOrganization(ctx, user)
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
		user.Status = statusFailed(err)
		return user, err
	}

	uid, _ := uuid.Parse(id)
	user, err = s.updateUserRoleRelation(ctx, user, parsedIds{Id: uid, Partner: partnerId, Organization: organizationId})
	if err != nil {
		user.Status = statusFailed(err)
		return user, err
	}

	rl, err := s.ap.GetRecoveryLink(ctx, id)
	fmt.Println("Recovery link:", rl) // TODO: email the recovery link to the user
	if err != nil {
		user.Status = statusFailed(err)
		return user, err
	}

	user.Status = statusOK()
	return user, nil
}

func (s *userService) identitiesModelToUser(ctx context.Context, user *userv3.User, usr *models.KratosIdentities) (*userv3.User, error) {
	traits := getUserTraits(usr.Traits)
	groups, err := s.udao.GetGroups(ctx, usr.ID)
	if err != nil {
		user.Status = statusFailed(err)
		return user, err
	}
	groupNames := []string{}
	for _, g := range groups {
		groupNames = append(groupNames, g.Name)
	}

	labels := make(map[string]string)

	roles, err := s.udao.GetRoles(ctx, usr.ID)
	if err != nil {
		user.Status = statusFailed(err)
		return user, err
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
		user.Status = statusFailed(err)
		return user, err
	}
	entity, err := s.dao.GetByID(ctx, uid, &models.KratosIdentities{})
	if err != nil {
		user.Status = statusFailed(err)
		return user, err
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		user, err := s.identitiesModelToUser(ctx, user, usr)
		if err != nil {
			user.Status = statusFailed(err)
			return user, err
		}

		user.Status = statusOK()
		return user, nil
	}
	user.Status = statusFailed(fmt.Errorf("unabele to fetch user '%v'", id))
	return user, nil

}

func (s *userService) GetByName(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	name := user.GetMetadata().GetName()
	entity, err := s.dao.GetByTraits(ctx, name, &models.KratosIdentities{})
	if err != nil {
		user.Status = statusFailed(err)
		return user, err
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		user, err := s.identitiesModelToUser(ctx, user, usr)
		if err != nil {
			user.Status = statusFailed(err)
			return user, err
		}

		user.Status = statusOK()
		return user, nil
	}
	fmt.Println("user:", user);
	return user, nil
}

func (s *userService) Update(ctx context.Context, user *userv3.User) (*userv3.User, error) {
	name := user.GetMetadata().GetName()
	entity, err := s.dao.GetIdByTraits(ctx, name, &models.KratosIdentities{})
	if err != nil {
		user.Status = statusFailed(fmt.Errorf("no user found with name '%v'", name))
		return user, err
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		partnerId, organizationId, err := s.getPartnerOrganization(ctx, user)
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
			user.Status = statusFailed(err)
			return user, err
		}

		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.AccountResourcerole{})
		if err != nil {
			user.Status = statusFailed(err)
			return nil, err
		}
		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.ProjectAccountResourcerole{})
		if err != nil {
			user.Status = statusFailed(err)
			return nil, err
		}
		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.ProjectAccountNamespaceRole{})
		if err != nil {
			user.Status = statusFailed(err)
			return nil, err
		}

		user, err = s.updateUserRoleRelation(ctx, user, parsedIds{Id: usr.ID, Partner: partnerId, Organization: organizationId})
		if err != nil {
			user.Status = statusFailed(err)
			return user, err
		}
	} else {
		user.Status = statusFailed(fmt.Errorf("unable to update user '%v'", name))
		return user, err
	}

	user.Status = statusOK()
	return user, nil
}

func (s *userService) Delete(ctx context.Context, user *userv3.User) (*userrpcv3.DeleteUserResponse, error) {
	name := user.GetMetadata().GetName()
	entity, err := s.dao.GetIdByTraits(ctx, name, &models.KratosIdentities{})
	if err != nil {
		return &userrpcv3.DeleteUserResponse{}, fmt.Errorf("no user founnd with username '%v'", name)
	}

	if usr, ok := entity.(*models.KratosIdentities); ok {
		err := s.ap.Delete(ctx, usr.ID.String())
		if err != nil {
			return &userrpcv3.DeleteUserResponse{}, err
		}

		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.GroupAccount{})
		if err != nil {
			return &userrpcv3.DeleteUserResponse{}, err
		}
		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.AccountResourcerole{})
		if err != nil {
			return &userrpcv3.DeleteUserResponse{}, err
		}
		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.ProjectAccountResourcerole{})
		if err != nil {
			return &userrpcv3.DeleteUserResponse{}, err
		}
		err = s.dao.DeleteX(ctx, "account_id", usr.ID, &models.ProjectAccountNamespaceRole{})
		if err != nil {
			return &userrpcv3.DeleteUserResponse{}, err
		}

		return &userrpcv3.DeleteUserResponse{}, nil
	}
	return &userrpcv3.DeleteUserResponse{}, fmt.Errorf("unable to find user '%v'", user.Metadata.Name)

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
	entities, err := s.dao.ListAll(ctx, &accs)
	if err != nil {
		return userList, err
	}
	if usrs, ok := entities.(*[]models.KratosIdentities); ok {
		for _, usr := range *usrs {
			user := &userv3.User{}
			user, err := s.identitiesModelToUser(ctx, user, &usr)
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

func (s *userService) Close() error {
	return s.dao.Close()
}
