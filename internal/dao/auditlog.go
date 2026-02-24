package dao

import (
	"context"
	"strings"

	"github.com/paralus/paralus/internal/models"
	"github.com/paralus/paralus/pkg/audit"
	"github.com/paralus/paralus/pkg/query"
	"github.com/uptrace/bun"
)

func GetAuditLogAggregations(ctx context.Context, db *bun.DB, tag, field string, filters query.QueryFilters) ([]models.AggregatorData, error) {
	var adata []models.AggregatorData
	sq := db.NewSelect().Table("audit_logs").
		ColumnExpr("count(1) as count")

	switch field {
	case "type":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("data->>'type' as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("data->>'type'")
		} else {
			sq.ColumnExpr("data->>'type' as key").
				Where("tag = ?", tag).GroupExpr("data->>'type'")
		}
	case "username":
		if tag == audit.KUBECTL_API {
			sq.ColumnExpr("data->>'username' as key").
				Where("tag = ?", tag).GroupExpr("data->>'username'")
		} else if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->'actor'->'account'->>'username', data->>'username') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->'actor'->'account'->>'username', data->>'username')")
		} else {
			sq.ColumnExpr("data->'actor'->'account'->>'username' as key").
				Where("tag = ?", tag).GroupExpr("data->'actor'->'account'->>'username'")
		}
	case "project":
		if tag == audit.KUBECTL_API {
			sq.ColumnExpr("data->>'project' as key").
				Where("tag = ?", tag).GroupExpr("data->>'project'")
		} else if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("data->>'project' as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("data->>'project'")
		} else {
			sq.ColumnExpr("data->>'project' as key").
				Where("tag = ?", tag).GroupExpr("data->>'project'")
		}
	case "cluster":
		if tag == audit.KUBECTL_API {
			sq.ColumnExpr("data->>'cluster_name' as key").
				Where("tag = ?", tag).GroupExpr("data->>'cluster_name'")
		} else if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->'detail'->'meta'->>'cluster_name', data->>'cluster_name') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->'detail'->'meta'->>'cluster_name', data->>'cluster_name')")
		} else {
			sq.ColumnExpr("data->'detail'->'meta'->>'cluster_name' as key").
				Where("tag = ?", tag).GroupExpr("data->'detail'->'meta'->>'cluster_name'")
		}
	case "namespace":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("data->>'namespace' as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("data->>'namespace'")
		} else {
			sq.ColumnExpr("data->>'namespace' as key").
				Where("tag = ?", tag).GroupExpr("data->>'namespace'")
		}
	case "kind":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("data->>'kind' as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("data->>'kind'")
		} else {
			sq.ColumnExpr("data->>'kind' as key").
				Where("tag = ?", tag).GroupExpr("data->>'kind'")
		}
	case "method":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("data->>'method' as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("data->>'method'")
		} else {
			sq.ColumnExpr("data->>'method' as key").
				Where("tag = ?", tag).GroupExpr("data->>'method'")
		}
	}

	switch tag {
	case audit.KUBECTL_API:
		sq = buildRelayAuditQuery(sq, filters)
	case audit.KUBECTL_CMD:
		sq = buildRelayCommandQuery(sq, filters)
	case audit.SYSTEM:
		sq = buildQuery(sq, filters)
	}

	err := sq.Scan(ctx, &adata)
	return adata, err
}

func GetAuditLogs(ctx context.Context, db *bun.DB, tag string, filters query.QueryFilters) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	sq := db.NewSelect().Model(&logs)

	if tag == audit.KUBECTL_CMD {
		sq.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("tag = ?", audit.KUBECTL_CMD).
				WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
		})
	} else {
		sq.Where("tag = ?", tag)
	}

	switch tag {
	case audit.KUBECTL_API:
		sq = buildRelayAuditQuery(sq, filters)
	case audit.KUBECTL_CMD:
		sq = buildRelayCommandQuery(sq, filters)
	case audit.SYSTEM:
		sq = buildQuery(sq, filters)
	}
	err := sq.Order("time desc").Scan(ctx)
	return logs, err
}

func buildRelayCommandQuery(query *bun.SelectQuery, filters query.QueryFilters) *bun.SelectQuery {
	if len(filters.GetProjects()) > 0 {
		for _, project := range filters.GetProjects() {
			query.Where("data->>'project' = ?", project)
		}
	}

	if filters.GetUser() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->'actor'->'account'->>'username' = ?", filters.GetUser()).
				WhereOr("data->>'username' = ?", filters.GetUser())
		})
	}

	if filters.GetKind() != "" {
		query.Where("data->>'kind' = ?", filters.GetKind())
	}

	if filters.GetMethod() != "" {
		query.Where("data->>'method' = ?", filters.GetMethod())
	}

	if filters.GetNamespace() != "" {
		query.Where("data->>'namespace' = ?", filters.GetNamespace())
	}

	if filters.GetCluster() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->'detail'->'meta'->>'cluster_name' = ?", filters.GetCluster()).
				WhereOr("data->>'cluster_name' = ?", filters.GetCluster())
		})
	}

	if filters.GetClient() != "" {
		if filters.GetClient() == "KUBECTL" {
			query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Where("data->'client'->>'type' = ?", filters.GetClient()).
					WhereOr("tag = ? AND data->>'session_type' != 'browser shell'", audit.KUBECTL_API)
			})
		} else {
			query.Where("data->'client'->>'type' = ?", filters.GetClient())
		}
	}

	if filters.GetTimefrom() != "" {
		diff := strings.Split(filters.GetTimefrom(), "-")[1]
		query.Where("time between now() - interval ? and now()", diff)
	}

	return query
}

func buildRelayAuditQuery(query *bun.SelectQuery, filters query.QueryFilters) *bun.SelectQuery {
	if filters.GetUser() != "" {
		query.Where("data->>'username' = ?", filters.GetUser())
	}
	if filters.GetKind() != "" {
		query.Where("data->>'kind' = ?", filters.GetKind())
	}
	if filters.GetMethod() != "" {
		query.Where("data->>'method' = ?", filters.GetMethod())
	}
	if filters.GetNamespace() != "" {
		query.Where("data->>'namespace' = ?", filters.GetNamespace())
	}
	if filters.GetCluster() != "" {
		query.Where("data->>'cluster_name' = ?", filters.GetCluster())
	}
	if filters.GetTimefrom() != "" {
		diff := strings.Split(filters.GetTimefrom(), "-")[1]
		query.Where("time between now() - interval ? and now()", diff)
	}
	if len(filters.GetProjects()) > 0 {
		for _, project := range filters.GetProjects() {
			query.Where("data->>'project' = ?", project)
		}
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

	if filters.GetCluster() != "" {
		query.Where("data->'detail'->'meta'->>'cluster_name' = ?", filters.GetCluster())
	}

	if filters.GetClient() != "" {
		query.Where("data->'client'->>'type' = ?", filters.GetClient())
	}

	if filters.GetTimefrom() != "" {
		diff := strings.Split(filters.GetTimefrom(), "-")[1]
		query.Where("time between now() - interval ? and now()", diff)
	}

	return query
}
