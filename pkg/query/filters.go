package query

import v1 "github.com/paralus/paralus/proto/rpc/audit"

type QueryFilters interface {
	GetType() string
	GetUser() string
	GetClient() string
	GetTimefrom() string
	GetPortal() string
	GetCluster() string
	GetNamespace() string
	GetKind() string
	GetMethod() string
	GetQueryString() string
	GetProjects() []string
}

var _ QueryFilters = (*v1.AuditLogQueryFilter)(nil)
var _ QueryFilters = (*v1.RelayAuditQueryFilter)(nil)
