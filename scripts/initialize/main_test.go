package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	userv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// newMockBunDB returns a *bun.DB backed by go-sqlmock, same pattern used
// throughout internal/dao's tests.
func newMockBunDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { sqldb.Close() })
	return bun.NewDB(sqldb, pgdialect.New()), mock
}

func writePermissionFile(t *testing.T, dir, name string) {
	content := `{
		"name": "` + name + `",
		"base_url": "/v1/` + name + `",
		"description": "desc",
		"scope": "org"
	}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".json"), []byte(content), 0644))
}

func TestAddResourcePermissions(t *testing.T) {
	t.Run("creates a new permission when none exists yet", func(t *testing.T) {
		dir := t.TempDir()
		writePermissionFile(t, dir, "perm-create")

		db, mock := newMockBunDB(t)
		mock.ExpectQuery(`(?i)SELECT`).WillReturnError(sql.ErrNoRows)
		// dao.Create inserts a model with a default-tagged uuid pk, which
		// pgdialect executes as "INSERT ... RETURNING", i.e. a Query, not
		// an Exec.
		mock.ExpectQuery(`(?i)INSERT`).WillReturnRows(sqlmock.NewRows([]string{"id"}))

		err := addResourcePermissions(db, dir)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("updates an existing permission in place", func(t *testing.T) {
		dir := t.TempDir()
		writePermissionFile(t, dir, "perm-update")

		db, mock := newMockBunDB(t)
		mock.ExpectQuery(`(?i)SELECT`).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("perm-update"))
		mock.ExpectExec(`(?i)UPDATE`).WillReturnResult(sqlmock.NewResult(0, 1))

		err := addResourcePermissions(db, dir)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propagates a malformed-JSON error", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not json"), 0644))

		db, _ := newMockBunDB(t)
		err := addResourcePermissions(db, dir)
		assert.Error(t, err)
	})

	t.Run("propagates a non-ErrNoRows lookup error", func(t *testing.T) {
		dir := t.TempDir()
		writePermissionFile(t, dir, "perm-lookup-fail")

		db, mock := newMockBunDB(t)
		mock.ExpectQuery(`(?i)SELECT`).WillReturnError(errors.New("connection reset"))

		err := addResourcePermissions(db, dir)
		assert.Error(t, err)
	})
}

// fakeGroupService implements service.GroupService with function fields.
type fakeGroupService struct {
	getByNameFn func(ctx context.Context, g *userv3.Group) (*userv3.Group, error)
	createFn    func(ctx context.Context, g *userv3.Group) (*userv3.Group, error)
}

func (f *fakeGroupService) Create(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
	return f.createFn(ctx, g)
}
func (f *fakeGroupService) GetByID(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
	panic("not implemented")
}
func (f *fakeGroupService) GetByName(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
	return f.getByNameFn(ctx, g)
}
func (f *fakeGroupService) Update(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
	panic("not implemented")
}
func (f *fakeGroupService) Delete(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
	panic("not implemented")
}
func (f *fakeGroupService) List(ctx context.Context, opts ...query.Option) (*userv3.GroupList, error) {
	panic("not implemented")
}

func TestCreateDefaultGroup(t *testing.T) {
	t.Run("returns the existing group without creating a new one", func(t *testing.T) {
		existing := &userv3.Group{Metadata: &commonv3.Metadata{Name: "All Local Users"}}
		gs := &fakeGroupService{getByNameFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
			return existing, nil
		}}

		got, err := createDefaultGroup(context.Background(), gs, "All Local Users", "p", "o", "desc", "DEFAULT_USERS", nil)
		require.NoError(t, err)
		assert.Same(t, existing, got)
	})

	t.Run("creates a new group when GetByName reports 'not found'", func(t *testing.T) {
		var created *userv3.Group
		gs := &fakeGroupService{
			getByNameFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
				return nil, errors.New("group not found")
			},
			createFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
				created = g
				return g, nil
			},
		}

		got, err := createDefaultGroup(context.Background(), gs, "Organization Admins", "p", "o", "desc", "DEFAULT_ADMINS", nil)
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, "Organization Admins", got.Metadata.Name)
		assert.Equal(t, "p", got.Metadata.Partner)
		assert.Equal(t, "o", got.Metadata.Organization)
	})

	t.Run("creates a new group when GetByName reports 'no rows in result set'", func(t *testing.T) {
		gs := &fakeGroupService{
			getByNameFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
				return nil, errors.New("sql: no rows in result set")
			},
			createFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
				return g, nil
			},
		}

		_, err := createDefaultGroup(context.Background(), gs, "grp", "p", "o", "desc", "TYPE", nil)
		require.NoError(t, err)
	})

	t.Run("propagates an unrelated lookup error without creating anything", func(t *testing.T) {
		want := errors.New("connection reset")
		called := false
		gs := &fakeGroupService{
			getByNameFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
				return nil, want
			},
			createFn: func(ctx context.Context, g *userv3.Group) (*userv3.Group, error) {
				called = true
				return nil, nil
			},
		}

		_, err := createDefaultGroup(context.Background(), gs, "grp", "p", "o", "desc", "TYPE", nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, want)
		assert.False(t, called, "should not call Create when the lookup error is not a not-found error")
	})
}
