package testutil

import (
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/service"
)

func NewClusterService() (service.ClusterService, error) {
	db := GetDB()

	cs := service.NewClusterService(db, db, nil, nil)
	return cs, nil
}
