package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProjectCluster struct {
	bun.BaseModel `bun:"table:cluster_project_cluster,alias:projectcluster"`

	ProjectID uuid.UUID `bun:"project_id,type:uuid,notnull"`
	ClusterID uuid.UUID `bun:"cluster_id,type:uuid,notnull"`
}
