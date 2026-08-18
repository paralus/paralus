package query

import (
	"context"
	"testing"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithMeta(t *testing.T) {
	meta := &commonv3.Metadata{
		Name: "n1", Id: "id1", Partner: "p1", Organization: "o1", Project: "proj1",
		Labels: map[string]string{"k": "v"}, Annotations: map[string]string{"a": "b"},
	}
	opts := &commonv3.QueryOptions{}
	WithMeta(meta)(opts)

	assert.Equal(t, "n1", opts.Name)
	assert.Equal(t, "id1", opts.ID)
	assert.Equal(t, "p1", opts.Partner)
	assert.Equal(t, "o1", opts.Organization)
	assert.Equal(t, "proj1", opts.Project)
	assert.Equal(t, "n1", opts.DisplayName)
	assert.Equal(t, map[string]string{"k": "v"}, opts.Labels)
	assert.Equal(t, map[string]string{"a": "b"}, opts.Annotations)
}

func TestWithOptions(t *testing.T) {
	in := &commonv3.QueryOptions{Name: "n1", Limit: 5, Offset: 100}
	opts := &commonv3.QueryOptions{}
	WithOptions(in)(opts)

	assert.Equal(t, "n1", opts.Name)
	assert.Equal(t, int64(5), opts.Limit)
	assert.Equal(t, int64(100), opts.Offset)
}

func TestSimpleOptionSetters(t *testing.T) {
	opts := &commonv3.QueryOptions{}

	WithIgnoreScopeDefault()(opts)
	assert.True(t, opts.IgnoreScopeDefault)

	WithExtended()(opts)
	assert.True(t, opts.Extended)

	WithGlobalScope()(opts)
	assert.True(t, opts.GlobalScope)

	WithName("n1")(opts)
	assert.Equal(t, "n1", opts.Name)

	WithSelector("env=prod")(opts)
	assert.Equal(t, "env=prod", opts.Selector)

	WithDeleted()(opts)
	assert.True(t, opts.Deleted)

	WithPartnerID("p1")(opts)
	assert.Equal(t, "p1", opts.Partner)

	WithOrganizationID("o1")(opts)
	assert.Equal(t, "o1", opts.Organization)

	WithProjectID("proj1")(opts)
	assert.Equal(t, "proj1", opts.Project)
}

func TestPaginate(t *testing.T) {
	t.Run("applies the default limit when unset", func(t *testing.T) {
		db := newTestDB(t)
		q := Paginate(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{})
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "LIMIT 10")
	})

	t.Run("caps the limit at MaxLimit", func(t *testing.T) {
		db := newTestDB(t)
		q := Paginate(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Limit: 1000})
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "LIMIT 50")
	})

	t.Run("applies the offset", func(t *testing.T) {
		db := newTestDB(t)
		q := Paginate(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Limit: 5, Offset: 20})
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "LIMIT 5")
		assert.Contains(t, sql, "OFFSET 20")
	})

	t.Run("skips pagination entirely for a negative limit", func(t *testing.T) {
		db := newTestDB(t)
		q := Paginate(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Limit: -1})
		sql := renderSelect(t, q)
		assert.NotContains(t, sql, "LIMIT")
	})

	t.Run("caps the offset at MaxOffset", func(t *testing.T) {
		db := newTestDB(t)
		q := Paginate(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Limit: 5, Offset: 999999999})
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "OFFSET 100000")
	})
}

func TestSelect(t *testing.T) {
	t.Run("applies partner/organization/project filters", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Select(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{
			Partner: "p1", Organization: "o1", Project: "proj1",
		})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		// The column projection always includes these names, so assert on
		// the actual WHERE-clause shape (?TableAlias substitution leaves
		// the column name unquoted), not just substring presence.
		assert.Contains(t, sql, `"te".partner_id = 'p1'`)
		assert.Contains(t, sql, `"te".organization_id = 'o1'`)
		assert.Contains(t, sql, `"te".project_id = 'proj1'`)
	})

	t.Run("skips scope filters under global scope", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Select(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{
			Partner: "p1", GlobalScope: true,
		})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.NotContains(t, sql, "partner_id = ")
	})

	t.Run("propagates an invalid selector error", func(t *testing.T) {
		db := newTestDB(t)
		_, err := Select(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "==="})
		assert.Error(t, err)
	})
}

func TestGetAccountID(t *testing.T) {
	t.Run("returns the account under normal scope", func(t *testing.T) {
		id, err := GetAccountID(&commonv3.QueryOptions{Account: "acct-1"})
		require.NoError(t, err)
		assert.Equal(t, "acct-1", id)
	})

	t.Run("returns empty under global scope", func(t *testing.T) {
		id, err := GetAccountID(&commonv3.QueryOptions{Account: "acct-1", GlobalScope: true})
		require.NoError(t, err)
		assert.Empty(t, id)
	})
}

func TestGetOrganizationID(t *testing.T) {
	t.Run("returns the org under normal scope", func(t *testing.T) {
		id, err := GetOrganizationID(&commonv3.QueryOptions{Organization: "o1"})
		require.NoError(t, err)
		assert.Equal(t, "o1", id)
	})

	t.Run("returns empty under global scope", func(t *testing.T) {
		id, err := GetOrganizationID(&commonv3.QueryOptions{Organization: "o1", GlobalScope: true})
		require.NoError(t, err)
		assert.Empty(t, id)
	})
}

func TestGetClusterID(t *testing.T) {
	t.Run("returns the cluster id under normal scope", func(t *testing.T) {
		id, err := GetClusterID(&commonv3.QueryOptions{ClusterID: "c1"})
		require.NoError(t, err)
		assert.Equal(t, "c1", id)
	})

	t.Run("returns empty under global scope", func(t *testing.T) {
		id, err := GetClusterID(&commonv3.QueryOptions{ClusterID: "c1", GlobalScope: true})
		require.NoError(t, err)
		assert.Empty(t, id)
	})
}

func TestGet(t *testing.T) {
	t.Run("filters by name when set", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Get(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Name: "n1"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, `"te".name = 'n1'`)
	})

	t.Run("filters by id when name is unset", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Get(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{ID: "id1"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, `"te".id = 'id1'`)
	})

	t.Run("errors when neither name nor id is set", func(t *testing.T) {
		db := newTestDB(t)
		_, err := Get(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{})
		assert.ErrorIs(t, err, ErrNoNameOrID)
	})

	t.Run("skips scope filters under global scope", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Get(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Name: "n1", Partner: "p1", GlobalScope: true})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.NotContains(t, sql, "partner_id = ")
	})

	t.Run("propagates invalid selector error from setRequestMeta path", func(t *testing.T) {
		// Get only calls setRequestMeta (not FilterLabels), so an invalid
		// selector does NOT surface here -- documenting that Get and Select
		// have different validation surfaces.
		db := newTestDB(t)
		_, err := Get(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Name: "n1", Selector: "==="})
		assert.NoError(t, err)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("always sets modified_at", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Update(db.NewUpdate().Model((*testEntity)(nil)).Where("id = ?", "x"), &commonv3.QueryOptions{})
		require.NoError(t, err)
		sql := renderUpdate(t, q)
		assert.Contains(t, sql, "modified_at")
	})

	t.Run("sets display_name/labels/annotations when present", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Update(db.NewUpdate().Model((*testEntity)(nil)).Where("id = ?", "x"), &commonv3.QueryOptions{
			DisplayName: "dn1",
			Labels:      map[string]string{"k": "v"},
			Annotations: map[string]string{"a": "b"},
		})
		require.NoError(t, err)
		sql := renderUpdate(t, q)
		assert.Contains(t, sql, "display_name")
		assert.Contains(t, sql, "labels")
		assert.Contains(t, sql, "annotations")
	})

	t.Run("omits display_name/labels/annotations when absent", func(t *testing.T) {
		db := newTestDB(t)
		q, err := Update(db.NewUpdate().Model((*testEntity)(nil)).Where("id = ?", "x"), &commonv3.QueryOptions{})
		require.NoError(t, err)
		sql := renderUpdate(t, q)
		assert.NotContains(t, sql, "display_name")
		assert.NotContains(t, sql, `"labels"`)
		assert.NotContains(t, sql, "annotations")
	})
}

func TestDelete(t *testing.T) {
	// BUG (also documented in internal/dao tests): Delete builds its
	// bun.UpdateQuery via q.DB().NewUpdate(), discarding the Model and
	// Where(name/id) that Get() attached to q, so the resulting query has no
	// table and fails at Exec time regardless of valid input.
	t.Run("errors when neither name nor id is set (Get's own validation)", func(t *testing.T) {
		db := newTestDB(t)
		_, err := Delete(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{})
		assert.ErrorIs(t, err, ErrNoNameOrID)
	})

	t.Run("BUG: produces a tableless update even with a valid name", func(t *testing.T) {
		db := newTestDB(t)
		uq, err := Delete(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Name: "n1"})
		require.NoError(t, err) // Delete itself doesn't error...
		_, execErr := uq.Exec(context.Background())
		assert.Error(t, execErr) // ...but the returned query can never execute.
	})
}
