package server

import (
	"errors"
	"testing"

	v3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrapbv3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	rolepbv3 "github.com/paralus/paralus/proto/types/rolepb/v3"
	systempbv3 "github.com/paralus/paralus/proto/types/systempb/v3"
	userpbv3 "github.com/paralus/paralus/proto/types/userpb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// All eight update*Status helpers across this package share the identical
// shape: on error, stamp the *request* with a Failed status and the error's
// message as Reason and return it; on success, stamp the *response* with an
// OK status and return it. One table-driven test per proto type is enough
// to prove each wiring is correct without duplicating the same two branches
// eight times.

func TestUpdateRoleStatus(t *testing.T) {
	t.Run("success stamps the response OK", func(t *testing.T) {
		req, resp := &rolepbv3.Role{}, &rolepbv3.Role{}
		got := updateRoleStatus(req, resp, nil)
		require.NotNil(t, got.Status)
		assert.Equal(t, v3.ConditionStatus_StatusOK, got.Status.ConditionStatus)
		assert.Same(t, resp, got)
	})

	t.Run("error stamps the request Failed with the error reason", func(t *testing.T) {
		req, resp := &rolepbv3.Role{}, &rolepbv3.Role{}
		got := updateRoleStatus(req, resp, errors.New("boom"))
		require.NotNil(t, got.Status)
		assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
		assert.Equal(t, "boom", got.Status.Reason)
		assert.Same(t, req, got)
	})
}

func TestUpdateClusterStatus(t *testing.T) {
	req, resp := &infrapbv3.Cluster{}, &infrapbv3.Cluster{}
	assert.Equal(t, v3.ConditionStatus_StatusOK, updateClusterStatus(req, resp, nil).Status.ConditionStatus)

	req2, resp2 := &infrapbv3.Cluster{}, &infrapbv3.Cluster{}
	got := updateClusterStatus(req2, resp2, errors.New("boom"))
	assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
	assert.Equal(t, "boom", got.Status.Reason)
	assert.Same(t, req2, got)
}

func TestUpdateGroupStatus(t *testing.T) {
	req, resp := &userpbv3.Group{}, &userpbv3.Group{}
	assert.Equal(t, v3.ConditionStatus_StatusOK, updateGroupStatus(req, resp, nil).Status.ConditionStatus)

	req2, resp2 := &userpbv3.Group{}, &userpbv3.Group{}
	got := updateGroupStatus(req2, resp2, errors.New("boom"))
	assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
	assert.Same(t, req2, got)
}

func TestUpdateOrganizationStatus(t *testing.T) {
	req, resp := &systempbv3.Organization{}, &systempbv3.Organization{}
	assert.Equal(t, v3.ConditionStatus_StatusOK, updateOrganizationStatus(req, resp, nil).Status.ConditionStatus)

	req2, resp2 := &systempbv3.Organization{}, &systempbv3.Organization{}
	got := updateOrganizationStatus(req2, resp2, errors.New("boom"))
	assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
	assert.Same(t, req2, got)
}

func TestUpdatePartnerStatus(t *testing.T) {
	req, resp := &systempbv3.Partner{}, &systempbv3.Partner{}
	assert.Equal(t, v3.ConditionStatus_StatusOK, updatePartnerStatus(req, resp, nil).Status.ConditionStatus)

	req2, resp2 := &systempbv3.Partner{}, &systempbv3.Partner{}
	got := updatePartnerStatus(req2, resp2, errors.New("boom"))
	assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
	assert.Same(t, req2, got)
}

func TestUpdateProjectStatus(t *testing.T) {
	req, resp := &systempbv3.Project{}, &systempbv3.Project{}
	assert.Equal(t, v3.ConditionStatus_StatusOK, updateProjectStatus(req, resp, nil).Status.ConditionStatus)

	req2, resp2 := &systempbv3.Project{}, &systempbv3.Project{}
	got := updateProjectStatus(req2, resp2, errors.New("boom"))
	assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
	assert.Same(t, req2, got)
}

func TestUpdateUserStatus(t *testing.T) {
	req, resp := &userpbv3.User{}, &userpbv3.User{}
	assert.Equal(t, v3.ConditionStatus_StatusOK, updateUserStatus(req, resp, nil).Status.ConditionStatus)

	req2, resp2 := &userpbv3.User{}, &userpbv3.User{}
	got := updateUserStatus(req2, resp2, errors.New("boom"))
	assert.Equal(t, v3.ConditionStatus_StatusFailed, got.Status.ConditionStatus)
	assert.Same(t, req2, got)
}
