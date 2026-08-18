package dao

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/paralus/paralus/pkg/audit"
	rpcv1 "github.com/paralus/paralus/proto/rpc/audit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAuditLogAggregations(t *testing.T) {
	fields := []string{"type", "username", "project", "cluster", "namespace", "kind", "method", "unknown-field"}
	tags := []string{audit.SYSTEM, audit.KUBECTL_CMD, audit.KUBECTL_API, "other-tag"}

	for _, field := range fields {
		for _, tag := range tags {
			t.Run(tag+"/"+field, func(t *testing.T) {
				db, mock := newMockBunDB(t)
				mock.ExpectQuery(".*").WillReturnRows(
					sqlmock.NewRows([]string{"count", "key"}).AddRow(int64(3), "x"),
				)

				res, err := GetAuditLogAggregations(context.Background(), db, tag, field, &rpcv1.AuditLogQueryFilter{
					User: "bob", Kind: "Pod", Method: "GET", Namespace: "ns-1", Cluster: "c1",
					Projects: []string{"p1", "p2"}, Type: "create", Client: "cli", Timefrom: "now-1h",
				})
				require.NoError(t, err)
				require.Len(t, res, 1)
				assert.Equal(t, int64(3), res[0].Count)
			})
		}
	}

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetAuditLogAggregations(context.Background(), db, audit.SYSTEM, "type", &rpcv1.AuditLogQueryFilter{})
		assert.Error(t, err)
	})
}

func TestGetAuditLogs(t *testing.T) {
	t.Run("kubectl_api tag applies relay filters", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"tag"}).AddRow(audit.KUBECTL_API))

		res, err := GetAuditLogs(context.Background(), db, audit.KUBECTL_API, &rpcv1.RelayAuditQueryFilter{
			User: "bob", Kind: "Pod", Method: "GET", Namespace: "ns-1", Cluster: "c1",
			Projects: []string{"p1"}, Timefrom: "now-1h",
		})
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("system tag applies generic filters", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"tag"}).AddRow(audit.SYSTEM))

		res, err := GetAuditLogs(context.Background(), db, audit.SYSTEM, &rpcv1.AuditLogQueryFilter{
			User: "bob", Type: "create", Client: "cli", Cluster: "c1",
			Projects: []string{"p1"}, Timefrom: "now-1h",
		})
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("unrecognized tag applies no extra filters", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"tag"}).AddRow("custom"))

		res, err := GetAuditLogs(context.Background(), db, "custom", &rpcv1.AuditLogQueryFilter{})
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("propagates query error", func(t *testing.T) {
		db, mock := newMockBunDB(t)
		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		_, err := GetAuditLogs(context.Background(), db, audit.SYSTEM, &rpcv1.AuditLogQueryFilter{})
		assert.Error(t, err)
	})
}
