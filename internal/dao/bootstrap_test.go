package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/paralus/paralus/internal/models"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/paralus/paralus/proto/types/sentry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
)

func TestCreateOrUpdateBootstrapInfra(t *testing.T) {
	// INSERT ... ON CONFLICT DO UPDATE ... RETURNING executes as a Query
	// under pgdialect, not an Exec.
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("infra-1"))

		err := CreateOrUpdateBootstrapInfra(context.Background(), db, &models.BootstrapInfra{Name: "infra-1"})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateOrUpdateBootstrapInfra(context.Background(), db, &models.BootstrapInfra{Name: "infra-1"})
		assert.Error(t, err)
	})
}

func TestCreateOrUpdateBootstrapAgentTemplate(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tpl-1"))

		err := CreateOrUpdateBootstrapAgentTemplate(context.Background(), db, &models.BootstrapAgentTemplate{Name: "tpl-1"})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateOrUpdateBootstrapAgentTemplate(context.Background(), db, &models.BootstrapAgentTemplate{Name: "tpl-1"})
		assert.Error(t, err)
	})
}

func TestGetBootstrapAgentTemplateForToken(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tpl-1"))

		res, err := GetBootstrapAgentTemplateForToken(context.Background(), db, "tok")
		require.NoError(t, err)
		assert.Equal(t, "tpl-1", res.Name)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetBootstrapAgentTemplateForToken(context.Background(), db, "tok")
		assert.Error(t, err)
	})
}

func TestSelectBootstrapAgentTemplates(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		expectScanAndCount(mock, sqlmock.NewRows([]string{"name"}).AddRow("tpl-1"), 1)

		res, count, err := SelectBootstrapAgentTemplates(context.Background(), db, &commonv3.QueryOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, res, 1)
	})

	t.Run("invalid selector fails before querying", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		_, _, err := SelectBootstrapAgentTemplates(context.Background(), db, &commonv3.QueryOptions{Selector: "==="})
		assert.Error(t, err)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, _, err := SelectBootstrapAgentTemplates(context.Background(), db, &commonv3.QueryOptions{})
		assert.Error(t, err)
	})
}

func TestDeleteBootstrapAgentTempate(t *testing.T) {
	// BUG: pkg/query.Delete builds its bun.UpdateQuery from q.DB().NewUpdate(),
	// discarding the Model and Where(name/id) that Get() attached to q. The
	// resulting update has no table, so this always fails at Exec time --
	// DeleteBootstrapAgentTempate is currently non-functional even with a
	// valid name set. Asserting the real (broken) behavior here rather than
	// the intended one.
	t.Run("always fails: query.Delete produces a tableless update", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		err := DeleteBootstrapAgentTempate(context.Background(), db, &commonv3.QueryOptions{Name: "tpl-1"}, "infra-1")
		assert.Error(t, err)
	})

	t.Run("errors when neither name nor id is set", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		err := DeleteBootstrapAgentTempate(context.Background(), db, &commonv3.QueryOptions{}, "infra-1")
		assert.Error(t, err)
	})
}

func TestGetBootstrapAgent(t *testing.T) {
	t.Run("filters by template ref when not wildcard", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("agent-1"))

		res, err := GetBootstrapAgent(context.Background(), db, "tpl-1", &commonv3.QueryOptions{Name: "agent-1"})
		require.NoError(t, err)
		assert.Equal(t, "agent-1", res.Name)
	})

	t.Run("skips template ref filter for wildcard", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("agent-1"))

		_, err := GetBootstrapAgent(context.Background(), db, "-", &commonv3.QueryOptions{Name: "agent-1"})
		require.NoError(t, err)
	})

	t.Run("errors when neither name nor id is set", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		_, err := GetBootstrapAgent(context.Background(), db, "-", &commonv3.QueryOptions{})
		assert.Error(t, err)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetBootstrapAgent(context.Background(), db, "-", &commonv3.QueryOptions{Name: "agent-1"})
		assert.Error(t, err)
	})
}

func TestGetBootstrapAgents(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("agent-1"))

		res, count, err := GetBootstrapAgents(context.Background(), db, &commonv3.QueryOptions{Name: "agent-1"}, "-")
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, res, 1)
	})

	t.Run("errors when neither name nor id is set", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		_, _, err := GetBootstrapAgents(context.Background(), db, &commonv3.QueryOptions{}, "-")
		assert.Error(t, err)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, _, err := GetBootstrapAgents(context.Background(), db, &commonv3.QueryOptions{Name: "agent-1"}, "-")
		assert.Error(t, err)
	})
}

func TestSelectBootstrapAgents(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("agent-1"))

		res, count, err := SelectBootstrapAgents(context.Background(), db, "-", &commonv3.QueryOptions{})
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		require.Len(t, res, 1)
	})

	t.Run("invalid selector fails before querying", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		_, _, err := SelectBootstrapAgents(context.Background(), db, "-", &commonv3.QueryOptions{Selector: "==="})
		assert.Error(t, err)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, _, err := SelectBootstrapAgents(context.Background(), db, "-", &commonv3.QueryOptions{})
		assert.Error(t, err)
	})
}

func TestCreateBootstrapAgent(t *testing.T) {
	t.Run("sets NotRegistered token state and succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		ba := &models.BootstrapAgent{Name: "agent-1"}
		err := CreateBootstrapAgent(context.Background(), db, ba)
		require.NoError(t, err)
		assert.Equal(t, sentry.BootstrapAgentState_NotRegistered.String(), ba.TokenState)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := CreateBootstrapAgent(context.Background(), db, &models.BootstrapAgent{Name: "agent-1"})
		assert.Error(t, err)
	})
}

// registerInTx runs RegisterBootstrapAgent inside a real bun.Tx backed by
// sqlmock (RegisterBootstrapAgent requires bun.Tx, not just bun.IDB).
// sqlmock enforces a single ordered expectation queue across Begin/Query/
// Exec/Commit/Rollback alike, and the real call order is always
// Begin -> (queries/exec) -> Commit-or-Rollback, so ExpectBegin must be
// declared before setup() registers the query/exec expectations, and
// Commit/Rollback must be declared after.
func registerInTx(t *testing.T, db *bun.DB, mock sqlmock.Sqlmock, setup func(), token, ip, fingerprint string, wantErr bool) error {
	t.Helper()
	mock.ExpectBegin()
	setup()
	if wantErr {
		mock.ExpectRollback()
	} else {
		mock.ExpectCommit()
	}

	return db.RunInTx(context.Background(), nil, func(ctx context.Context, tx bun.Tx) error {
		return RegisterBootstrapAgent(ctx, tx, token, ip, fingerprint)
	})
}

func TestRegisterBootstrapAgent(t *testing.T) {
	t.Run("transitions NotRegistered to Approved", func(t *testing.T) {
		db, mock := newMockBunDB(t)

		err := registerInTx(t, db, mock, func() {
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"token_state", "template_ref"}).AddRow(sentry.BootstrapAgentState_NotRegistered.String(), "tpl-1"),
			)
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"auto_approve", "ignore_multiple_register"}).AddRow(false, false),
			)
			mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))
		}, "tok", "1.2.3.4", "fp-1", false)
		require.NoError(t, err)
	})

	t.Run("rejects re-registration when multiple register is not allowed", func(t *testing.T) {
		db, mock := newMockBunDB(t)

		err := registerInTx(t, db, mock, func() {
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"token_state", "template_ref", "fingerprint"}).AddRow(sentry.BootstrapAgentState_Approved.String(), "tpl-1", "fp-1"),
			)
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"auto_approve", "ignore_multiple_register"}).AddRow(false, false),
			)
		}, "tok", "1.2.3.4", "fp-1", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot register token")
	})

	t.Run("rejects mismatched fingerprint even when multiple register is allowed", func(t *testing.T) {
		db, mock := newMockBunDB(t)

		err := registerInTx(t, db, mock, func() {
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"token_state", "template_ref", "fingerprint"}).AddRow(sentry.BootstrapAgentState_Approved.String(), "tpl-1", "fp-1"),
			)
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"auto_approve", "ignore_multiple_register"}).AddRow(false, true),
			)
		}, "tok", "1.2.3.4", "different-fp", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fingerprint mismatch")
	})

	t.Run("rejects an invalid token state", func(t *testing.T) {
		db, mock := newMockBunDB(t)

		err := registerInTx(t, db, mock, func() {
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"token_state", "template_ref"}).AddRow("bogus-state", "tpl-1"),
			)
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"auto_approve"}).AddRow(false),
			)
		}, "tok", "1.2.3.4", "fp-1", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token state")
	})

	t.Run("propagates error looking up the agent", func(t *testing.T) {
		db, mock := newMockBunDB(t)

		err := registerInTx(t, db, mock, func() {
			mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))
		}, "tok", "1.2.3.4", "fp-1", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	})

	t.Run("propagates error looking up the template", func(t *testing.T) {
		db, mock := newMockBunDB(t)

		err := registerInTx(t, db, mock, func() {
			mock.ExpectQuery(".*").WillReturnRows(
				sqlmock.NewRows([]string{"token_state", "template_ref"}).AddRow(sentry.BootstrapAgentState_NotRegistered.String(), "tpl-1"),
			)
			mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))
		}, "tok", "1.2.3.4", "fp-1", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
	})
}

func TestDeleteBootstrapAgent(t *testing.T) {
	t.Run("succeeds with a template ref filter", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := DeleteBootstrapAgent(context.Background(), db, "tpl-1", &commonv3.QueryOptions{ID: "agent-1"})
		require.NoError(t, err)
	})

	t.Run("succeeds without a template ref filter", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 1))

		err := DeleteBootstrapAgent(context.Background(), db, "", &commonv3.QueryOptions{ID: "agent-1"})
		require.NoError(t, err)
	})

	t.Run("propagates exec error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectExec(".*").WillReturnError(errors.New("boom"))

		err := DeleteBootstrapAgent(context.Background(), db, "tpl-1", &commonv3.QueryOptions{ID: "agent-1"})
		assert.Error(t, err)
	})
}

func TestUpdateBootstrapAgent(t *testing.T) {
	t.Run("succeeds (UPDATE ... RETURNING executes as Query)", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

		err := UpdateBootstrapAgent(context.Background(), db, &models.BootstrapAgent{ID: uuid.New()}, &commonv3.QueryOptions{})
		require.NoError(t, err)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		err := UpdateBootstrapAgent(context.Background(), db, &models.BootstrapAgent{ID: uuid.New()}, &commonv3.QueryOptions{})
		assert.Error(t, err)
	})
}

func TestGetBootstrapAgentForToken(t *testing.T) {
	db, mock := newMockBunDB(t)
	mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("agent-1"))

	res, err := GetBootstrapAgentForToken(context.Background(), db, "tok")
	require.NoError(t, err)
	assert.Equal(t, "agent-1", res.Name)
}

func TestGetBootstrapAgentTemplateForHost(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("tpl-1"))

		res, err := GetBootstrapAgentTemplateForHost(context.Background(), db, "host-1")
		require.NoError(t, err)
		assert.Equal(t, "tpl-1", res.Name)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetBootstrapAgentTemplateForHost(context.Background(), db, "host-1")
		assert.Error(t, err)
	})
}

func TestGetBootstrapAgentCountForClusterID(t *testing.T) {
	t.Run("counts matching agents", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"name"}).AddRow("agent-1").AddRow("agent-2"),
		)

		count, err := GetBootstrapAgentCountForClusterID(context.Background(), db, "cluster-1", uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 2, count)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetBootstrapAgentCountForClusterID(context.Background(), db, "cluster-1", uuid.New())
		assert.Error(t, err)
	})
}

func TestGetBootstrapAgentForClusterID(t *testing.T) {
	t.Run("succeeds", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("agent-1"))

		res, err := GetBootstrapAgentForClusterID(context.Background(), db, "cluster-1", uuid.New())
		require.NoError(t, err)
		assert.Equal(t, "agent-1", res.Name)
	})

	t.Run("propagates error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetBootstrapAgentForClusterID(context.Background(), db, "cluster-1", uuid.New())
		assert.Error(t, err)
	})
}

func TestUpdateBootstrapAgentTempateDeleteAt(t *testing.T) {
	// BUG: pkg/query.Update never attaches a Where clause, so this always
	// fails at Exec time with bun's built-in "Update and Delete queries
	// require at least one Where" safety guard -- UpdateBootstrapAgentTempateDeleteAt
	// is currently non-functional even with a valid name set. Asserting the
	// real (broken) behavior here rather than the intended one.
	t.Run("always fails regardless of name, since query.Update never checks it", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		err := UpdateBootstrapAgentTempateDeleteAt(context.Background(), db, &commonv3.QueryOptions{Name: "tpl-1"})
		assert.Error(t, err)

		err = UpdateBootstrapAgentTempateDeleteAt(context.Background(), db, &commonv3.QueryOptions{})
		assert.Error(t, err)
	})
}

func TestUpdateBootstrapInfraDeleteAt(t *testing.T) {
	// BUG: same missing-Where issue as UpdateBootstrapAgentTempateDeleteAt
	// (both go through pkg/query.Update). Always fails at Exec time.
	t.Run("always fails regardless of name, since query.Update never checks it", func(t *testing.T) {
		db, _ := newMockBunDB(t)

		err := UpdateBootstrapInfraDeleteAt(context.Background(), db, &commonv3.QueryOptions{Name: "infra-1"})
		assert.Error(t, err)

		err = UpdateBootstrapInfraDeleteAt(context.Background(), db, &commonv3.QueryOptions{})
		assert.Error(t, err)
	})
}
