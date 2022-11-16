package dao

import (
	"context"
	"strings"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/query"
	"github.com/uptrace/bun"
)

func GetAuditLogAggregations(ctx context.Context, db *bun.DB, tag, field string, filters query.QueryFilters) ([]models.AggregatorData, error) {
	var adata []models.AggregatorData
	sq := db.NewSelect().Table("audit_logs").
		ColumnExpr("count(1) as count")

	switch field {
	case "type":
		sq.ColumnExpr("data->>'type' as key").
			Where("tag = ?", tag).GroupExpr("data->>'type'")
	case "username":
		if tag != "kubectl_api" {
			sq.ColumnExpr("data->'actor'->'account'->>'username' as key").
				Where("tag = ?", tag).GroupExpr("data->'actor'->'account'->>'username'")
		} else {
			sq.ColumnExpr("data->>'un' as key").
				Where("tag = ?", tag).GroupExpr("data->>'un'")
		}
	case "project":
		sq.ColumnExpr("data->>'project' as key").
			Where("tag = ?", tag).GroupExpr("data->>'project'")
	case "cluster":
		sq.ColumnExpr("data->>'cn' as key").
			Where("tag = ?", tag).GroupExpr("data->>'cn'")
	case "namespace":
		sq.ColumnExpr("data->>'n' as key").
			Where("tag = ?", tag).GroupExpr("data->>'n'")
	case "kind":
		sq.ColumnExpr("data->>'k' as key").
			Where("tag = ?", tag).GroupExpr("data->>'k'")
	case "method":
		sq.ColumnExpr("data->>'m' as key").
			Where("tag = ?", tag).GroupExpr("data->>'m'")
	}

	// add filters
	switch tag {
	case "kubectl_api":
		sq = buildRelayAuditQuery(sq, filters)
	case "system", "kubectl_cmd":
		sq = buildQuery(sq, filters)
	}

	err := sq.Scan(ctx, &adata)
	return adata, err
}

func GetAuditLogs(ctx context.Context, db *bun.DB, tag string, filters query.QueryFilters) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	sq := db.NewSelect().Model(&logs).
		Where("tag = ?", tag)

	switch tag {
	case "kubectl_api":
		sq = buildRelayAuditQuery(sq, filters)
	case "system", "kubectl_cmd":
		sq = buildQuery(sq, filters)
	}

	err := sq.Scan(ctx)
	return logs, err
}

func buildRelayAuditQuery(query *bun.SelectQuery, filters query.QueryFilters) *bun.SelectQuery {
	if filters.GetUser() != "" {
		query.Where("data->>'un' = ?", filters.GetUser())
	}

	if filters.GetKind() != "" {
		query.Where("data->>'k' = ?", filters.GetKind())
	}

	if filters.GetMethod() != "" {
		query.Where("data->>'m' = ?", filters.GetMethod())
	}
	if filters.GetNamespace() != "" {
		query.Where("data->>'ns' = ?", filters.GetNamespace())
	}
	if filters.GetCluster() != "" {
		query.Where("data->>'cn' = ?", filters.GetCluster())
	}
	if filters.GetTimefrom() != "" {
		diff := strings.Split(filters.GetTimefrom(), "-")[1]
		query.Where("to_timestamp(data->>'ts', 'YYYY-MM-DD\"T\"HH:MI:SS') between now() - interval ? and now()", diff)
	}
	return query
}

func buildQuery(query *bun.SelectQuery, filters query.QueryFilters) *bun.SelectQuery {
	if len(filters.GetProjects()) > 0 {
		for _, project := range filters.GetProjects() {
			query.Where("data->>'project' = ?", project)
		}
	}

	if filters.GetType() != "" {
		query.Where("data->>'type' = ?", filters.GetType())
	}

	if filters.GetUser() != "" {
		query.Where("data->'actor'->'account'->>'username' = ?", filters.GetUser())
	}

	if filters.GetClient() != "" {
		query.Where("data->'client'->>'type' = ?", filters.GetClient())
	}

	if filters.GetTimefrom() != "" {
		diff := strings.Split(filters.GetTimefrom(), "-")[1]
		query.Where("to_timestamp(data->>'timestamp', 'YYYY-MM-DD\"T\"HH:MI:SS') between now() - interval ? and now()", diff)
	}

	return query
}
