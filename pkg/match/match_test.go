package match

import (
	"testing"

	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("invalid selector returns an error", func(t *testing.T) {
		_, err := New(query.WithSelector("==="))
		assert.Error(t, err)
	})

	t.Run("valid options build a matcher", func(t *testing.T) {
		m, err := New(query.WithPartnerID("p1"), query.WithOrganizationID("o1"), query.WithProjectID("proj1"), query.WithName("n1"))
		require.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestMatcher_Match(t *testing.T) {
	newMatcher := func(t *testing.T, opts ...query.Option) Matcher {
		m, err := New(opts...)
		require.NoError(t, err)
		return m
	}

	t.Run("matches on partner, organization, and project", func(t *testing.T) {
		m := newMatcher(t, query.WithPartnerID("p1"), query.WithOrganizationID("o1"), query.WithProjectID("proj1"))
		assert.True(t, m.Match(commonv3.Metadata{Partner: "p1", Organization: "o1", Project: "proj1"}))
	})

	t.Run("rejects mismatched partner", func(t *testing.T) {
		m := newMatcher(t, query.WithPartnerID("p1"))
		assert.False(t, m.Match(commonv3.Metadata{Partner: "different"}))
	})

	t.Run("rejects mismatched organization", func(t *testing.T) {
		m := newMatcher(t, query.WithOrganizationID("o1"))
		assert.False(t, m.Match(commonv3.Metadata{Organization: "different"}))
	})

	t.Run("rejects mismatched project", func(t *testing.T) {
		m := newMatcher(t, query.WithProjectID("proj1"))
		assert.False(t, m.Match(commonv3.Metadata{Project: "different"}))
	})

	t.Run("empty name in options matches any name", func(t *testing.T) {
		m := newMatcher(t)
		assert.True(t, m.Match(commonv3.Metadata{Name: "anything"}))
	})

	t.Run("explicit name must match exactly", func(t *testing.T) {
		m := newMatcher(t, query.WithName("expected"))
		assert.True(t, m.Match(commonv3.Metadata{Name: "expected"}))
		assert.False(t, m.Match(commonv3.Metadata{Name: "other"}))
	})

	t.Run("label selector must match", func(t *testing.T) {
		m := newMatcher(t, query.WithSelector("env=prod"))
		assert.True(t, m.Match(commonv3.Metadata{Labels: map[string]string{"env": "prod"}}))
		assert.False(t, m.Match(commonv3.Metadata{Labels: map[string]string{"env": "staging"}}))
	})

	t.Run("empty selector matches everything", func(t *testing.T) {
		m := newMatcher(t)
		assert.True(t, m.Match(commonv3.Metadata{Labels: map[string]string{"anything": "goes"}}))
	})
}
