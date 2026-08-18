package query

import (
	"testing"

	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterLabels(t *testing.T) {
	t.Run("empty selector adds no filter", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.NotContains(t, sql, "WHERE")
	})

	t.Run("Equals uses the contains clause", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "env=prod"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "labels")
		assert.Contains(t, sql, "@>")
	})

	t.Run("NotEquals uses the not-contains clause", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "env!=prod"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "NOT")
		assert.Contains(t, sql, "@>")
	})

	t.Run("In uses an OR group of contains clauses", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "env in (prod,staging)"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "OR")
		assert.Contains(t, sql, "@>")
	})

	t.Run("NotIn uses a group of not-contains clauses", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "env notin (prod,staging)"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "NOT")
		assert.Contains(t, sql, "@>")
	})

	t.Run("Exists uses the key-exists clause", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "env"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "labels ? ")
	})

	t.Run("DoesNotExist uses the key-does-not-exist clause", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "!env"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "NOT")
		assert.Contains(t, sql, "labels ? ")
	})

	t.Run("multiple requirements are ANDed", func(t *testing.T) {
		db := newTestDB(t)
		q, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "env=prod,tier=frontend"})
		require.NoError(t, err)
		sql := renderSelect(t, q)
		assert.Contains(t, sql, "AND")
	})

	t.Run("invalid selector syntax errors", func(t *testing.T) {
		db := newTestDB(t)
		_, err := FilterLabels(db.NewSelect().Model(&testEntity{}), &commonv3.QueryOptions{Selector: "==="})
		assert.Error(t, err)
	})
}
