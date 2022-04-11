package service

import (
	"context"

	"github.com/RafayLabs/rcloud-base/internal/dao"
	"github.com/RafayLabs/rcloud-base/pkg/common"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func unique(items []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range items {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func containsu(s []uuid.UUID, id uuid.UUID) bool {
	for _, v := range s {
		if v == id {
			return true
		}
	}
	return false
}

func remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func diff(before, after []string) ([]string, []string, []string) {
	cu := []string{}
	uu := []string{}
	du := []string{}

	for _, u := range after {
		if contains(before, u) {
			uu = append(uu, u)
		} else {
			cu = append(du, u)
		}
	}
	for _, u := range before {
		if !contains(uu, u) && !contains(du, u) {
			du = append(cu, u)
		}
	}
	return cu, uu, du
}

// Given two lists, return newly created, unchanged and deleted items
func diffu(before, after []uuid.UUID) ([]uuid.UUID, []uuid.UUID, []uuid.UUID) {
	cu := []uuid.UUID{}
	uu := []uuid.UUID{}
	du := []uuid.UUID{}

	for _, u := range after {
		if containsu(before, u) {
			uu = append(uu, u)
		} else {
			cu = append(du, u)
		}
	}
	for _, u := range before {
		if !containsu(uu, u) && !containsu(du, u) {
			du = append(cu, u)
		}
	}
	return cu, uu, du
}

func getPartnerOrganization(ctx context.Context, db bun.IDB, partner, org string) (uuid.UUID, uuid.UUID, error) {
	partnerId, err := dao.GetPartnerId(ctx, db, partner)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	organizationId, err := dao.GetOrganizationId(ctx, db, org)
	if err != nil {
		return partnerId, uuid.Nil, err
	}
	return partnerId, organizationId, nil

}

func GetSessionDataFromContext(ctx context.Context) (*commonv3.SessionData, bool) {
	s, ok := ctx.Value(common.SessionDataKey).(*commonv3.SessionData)
	return s, ok
}
