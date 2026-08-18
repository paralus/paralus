package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGroupPermissions(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"group_name"}).AddRow("team-a"))

	res, err := GetGroupPermissions(context.Background(), db, []string{"team-a"}, uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetGroupPermissions_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetGroupPermissions(context.Background(), db, []string{"team-a"}, uuid.New(), uuid.New())
	assert.Error(t, err)
}

func TestGetGroupProjectsByPermission(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"permission_name"}).AddRow("project.read"))

	res, err := GetGroupProjectsByPermission(context.Background(), db, []string{"team-a"}, uuid.New(), uuid.New(), "project.read")
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetGroupProjectsByPermission_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetGroupProjectsByPermission(context.Background(), db, nil, uuid.New(), uuid.New(), "x")
	assert.Error(t, err)
}

func TestGetGroupPermissionsByProjectIDPermissions(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"project_name"}).AddRow("proj-a"))

	res, err := GetGroupPermissionsByProjectIDPermissions(context.Background(), db, []string{"team-a"}, uuid.New(), uuid.New(), []string{"p1"}, []string{"perm1"})
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetGroupPermissionsByProjectIDPermissions_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetGroupPermissionsByProjectIDPermissions(context.Background(), db, nil, uuid.New(), uuid.New(), nil, nil)
	assert.Error(t, err)
}

func TestGetProjectByGroup(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"project_name"}).AddRow("proj-a"))

	res, err := GetProjectByGroup(context.Background(), db, []string{"team-a"}, uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetProjectByGroup_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetProjectByGroup(context.Background(), db, nil, uuid.New(), uuid.New())
	assert.Error(t, err)
}

func TestGetAccountPermissions(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"permission_name"}).AddRow("cluster.view"))

	res, err := GetAccountPermissions(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetAccountPermissions_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetAccountPermissions(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
	assert.Error(t, err)
}

func TestIsPartnerSuperAdmin(t *testing.T) {
	t.Run("detects super admin", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role_name"}).AddRow("SUPER_ADMIN"))

		isPartnerAdmin, isSuperAdmin, err := IsPartnerSuperAdmin(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.True(t, isSuperAdmin)
		assert.False(t, isPartnerAdmin)
	})

	t.Run("detects partner admin", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role_name"}).AddRow("PARTNER_ADMIN"))

		isPartnerAdmin, isSuperAdmin, err := IsPartnerSuperAdmin(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.False(t, isSuperAdmin)
		assert.True(t, isPartnerAdmin)
	})

	t.Run("neither when no matching rows", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role_name"}))

		isPartnerAdmin, isSuperAdmin, err := IsPartnerSuperAdmin(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.False(t, isSuperAdmin)
		assert.False(t, isPartnerAdmin)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, _, err := IsPartnerSuperAdmin(context.Background(), db, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestGetAccountProjectsByPermission(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"permission_name"}).AddRow("cluster.view"))

	res, err := GetAccountProjectsByPermission(context.Background(), db, uuid.New(), uuid.New(), uuid.New(), "cluster.view")
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetAccountProjectsByPermission_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetAccountProjectsByPermission(context.Background(), db, uuid.New(), uuid.New(), uuid.New(), "x")
	assert.Error(t, err)
}

func TestGetAccountProjectByPermission(t *testing.T) {
	t.Run("applies partner and org filters when set", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"permission_name"}).AddRow("cluster.view"))

		_, err := GetAccountProjectByPermission(context.Background(), db, uuid.New(), uuid.New(), uuid.New(), "cluster.view")
		require.NoError(t, err)
	})

	t.Run("skips partner/org filters when nil", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"permission_name"}).AddRow("cluster.view"))

		_, err := GetAccountProjectByPermission(context.Background(), db, uuid.New(), uuid.Nil, uuid.Nil, "cluster.view")
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetAccountProjectByPermission(context.Background(), db, uuid.New(), uuid.Nil, uuid.Nil, "x")
		assert.Error(t, err)
	})
}

func TestGetDefaultAccountProject(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(uuid.New()))

	_, err := GetDefaultAccountProject(context.Background(), db, uuid.New())
	require.NoError(t, err)
}

func TestGetDefaultAccountProject_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetDefaultAccountProject(context.Background(), db, uuid.New())
	assert.Error(t, err)
}

func TestGetAccountPermissionsByProjectIDPermissions(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"permission_name"}).AddRow("cluster.view"))

	res, err := GetAccountPermissionsByProjectIDPermissions(context.Background(), db, uuid.New(), uuid.New(), uuid.New(), []uuid.UUID{uuid.New()}, []string{"cluster.view"})
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetAccountPermissionsByProjectIDPermissions_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetAccountPermissionsByProjectIDPermissions(context.Background(), db, uuid.New(), uuid.New(), uuid.New(), nil, nil)
	assert.Error(t, err)
}

func TestGetSSOUsersGroupProjectRole(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("a@b.com"))

	res, err := GetSSOUsersGroupProjectRole(context.Background(), db, uuid.New())
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetSSOUsersGroupProjectRole_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetSSOUsersGroupProjectRole(context.Background(), db, uuid.New())
	assert.Error(t, err)
}

func TestGetAcccountsWithApprovalPermission(t *testing.T) {
	t.Run("returns distinct usernames", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("a@b.com"))

		res, err := GetAcccountsWithApprovalPermission(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.Equal(t, []string{"a@b.com"}, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetAcccountsWithApprovalPermission(context.Background(), db, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestGetSSOAcccountsWithApprovalPermission(t *testing.T) {
	t.Run("de-duplicates usernames", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"username"}).AddRow("a@b.com").AddRow("a@b.com"),
		)

		res, err := GetSSOAcccountsWithApprovalPermission(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.Equal(t, []string{"a@b.com"}, res)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetSSOAcccountsWithApprovalPermission(context.Background(), db, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestIsOrgAdmin(t *testing.T) {
	t.Run("true when a matching admin/organization row exists", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role_name", "scope"}).AddRow("admin", "organization"),
		)

		got, err := IsOrgAdmin(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("false with no matching rows", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role_name", "scope"}))

		got, err := IsOrgAdmin(context.Background(), db, uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := IsOrgAdmin(context.Background(), db, uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestIsOrgReadOnly(t *testing.T) {
	t.Run("true when a matching admin_read_only/organization row exists", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"role_name", "scope"}).AddRow("admin_read_only", "organization"),
		)

		got, err := IsOrgReadOnly(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("false with no matching rows", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"role_name", "scope"}))

		got, err := IsOrgReadOnly(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		require.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := IsOrgReadOnly(context.Background(), db, uuid.New(), uuid.New(), uuid.New())
		assert.Error(t, err)
	})
}

func TestGetAccountBasics(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"username"}).AddRow("a@b.com"))

	res, err := GetAccountBasics(context.Background(), db, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "a@b.com", res.Username)
}

func TestGetAccountBasics_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetAccountBasics(context.Background(), db, uuid.New())
	assert.Error(t, err)
}

func TestGetAccountGroups(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("team-a"))

	res, err := GetAccountGroups(context.Background(), db, uuid.New())
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestGetAccountGroups_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetAccountGroups(context.Background(), db, uuid.New())
	assert.Error(t, err)
}

func TestGetDefaultUserGroup(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Default Users"))

	res, err := GetDefaultUserGroup(context.Background(), db, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "Default Users", res.Name)
}

func TestGetDefaultUserGroup_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetDefaultUserGroup(context.Background(), db, uuid.New())
	assert.Error(t, err)
}

func TestGetDefaultUserGroupAccount(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow(uuid.New()))

	_, err := GetDefaultUserGroupAccount(context.Background(), db, uuid.New(), uuid.New())
	require.NoError(t, err)
}

func TestGetDefaultUserGroupAccount_Error(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

	_, err := GetDefaultUserGroupAccount(context.Background(), db, uuid.New(), uuid.New())
	assert.Error(t, err)
}
