package service

import (
	v3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func statusFailed(err error) *v3.Status {
	// maybe we can parse the errors here and give better user facing errors
	return &v3.Status{
		ConditionType:   "StatusFailed",
		ConditionStatus: v3.ConditionStatus_StatusFailed,
		LastUpdated:     timestamppb.Now(),
		Reason:          err.Error(),
	}
}

func statusOK() *v3.Status {
	return &v3.Status{
		ConditionType:   "StatusOK",
		ConditionStatus: v3.ConditionStatus_StatusOK,
	}
}
