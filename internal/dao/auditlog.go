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
		sq.ColumnExpr("data->>'type' as key").
			Where("tag = ?", tag).GroupExpr("data->>'type'")
	case "username":
		if tag == audit.KUBECTL_API {
			sq.ColumnExpr("data->>'un' as key").
				Where("tag = ?", tag).GroupExpr("data->>'un'")
		} else if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->'actor'->'account'->>'username', data->>'un') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->'actor'->'account'->>'username', data->>'un')")
		} else {
			sq.ColumnExpr("data->'actor'->'account'->>'username' as key").
				Where("tag = ?", tag).GroupExpr("data->'actor'->'account'->>'username'")
		}
	case "project":
		if tag == audit.KUBECTL_API {
			sq.ColumnExpr("data->>'pr' as key").
				Where("tag = ?", tag).GroupExpr("data->>'pr'")
		} else if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->>'project', data->>'pr') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->>'project', data->>'pr')")
		} else {
			sq.ColumnExpr("data->>'project' as key").
				Where("tag = ?", tag).GroupExpr("data->>'project'")
		}
	case "cluster":
		if tag == audit.KUBECTL_API {
			sq.ColumnExpr("data->>'cn' as key").
				Where("tag = ?", tag).GroupExpr("data->>'cn'")
		} else if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->'detail'->'meta'->>'cluster_name', data->>'cn') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->'detail'->'meta'->>'cluster_name', data->>'cn')")
		} else {
			sq.ColumnExpr("data->'detail'->'meta'->>'cluster_name' as key").
				Where("tag = ?", tag).GroupExpr("data->'detail'->'meta'->>'cluster_name'")
		}
	case "namespace":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->>'namespace', data->>'ns', data->>'n') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->>'namespace', data->>'ns', data->>'n')")
		} else {
			sq.ColumnExpr("data->>'n' as key").
				Where("tag = ?", tag).GroupExpr("data->>'n'")
		}
	case "kind":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->>'kind', data->>'k') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->>'kind', data->>'k')")
		} else {
			sq.ColumnExpr("data->>'k' as key").
				Where("tag = ?", tag).GroupExpr("data->>'k'")
		}
	case "method":
		if tag == audit.KUBECTL_CMD {
			sq.ColumnExpr("COALESCE(data->>'method', data->>'m') as key").
				WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Where("tag = ?", audit.KUBECTL_CMD).
						WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
				}).
				GroupExpr("COALESCE(data->>'method', data->>'m')")
		} else {
			sq.ColumnExpr("data->>'m' as key").
				Where("tag = ?", tag).GroupExpr("data->>'m'")
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
				WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
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
			query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Where("data->>'project' = ?", project).
					WhereOr("data->>'pr' = ?", project)
			})
		}
	}

	if filters.GetUser() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->'actor'->'account'->>'username' = ?", filters.GetUser()).
				WhereOr("data->>'un' = ?", filters.GetUser())
		})
	}

	if filters.GetKind() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->>'kind' = ?", filters.GetKind()).
				WhereOr("data->>'k' = ?", filters.GetKind())
		})
	}

	if filters.GetMethod() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->>'method' = ?", filters.GetMethod()).
				WhereOr("data->>'m' = ?", filters.GetMethod())
		})
	}

	if filters.GetNamespace() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->>'namespace' = ?", filters.GetNamespace()).
				WhereOr("data->>'ns' = ?", filters.GetNamespace()).
				WhereOr("data->>'n' = ?", filters.GetNamespace())
		})
	}

	if filters.GetCluster() != "" {
		query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("data->'detail'->'meta'->>'cluster_name' = ?", filters.GetCluster()).
				WhereOr("data->>'cn' = ?", filters.GetCluster())
		})
	}

	if filters.GetClient() != "" {
		if filters.GetClient() == "KUBECTL" {
			query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Where("data->'client'->>'type' = ?", filters.GetClient()).
					WhereOr("tag = ? AND data->>'st' != 'browser shell'", audit.KUBECTL_API)
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
		query.Where("time between now() - interval ? and now()", diff)
	}
	if len(filters.GetProjects()) > 0 {
		for _, project := range filters.GetProjects() {
			query.Where("data->>'pr' = ?", project)
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
