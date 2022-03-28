package service

import (
	"context"

	"github.com/RafayLabs/rcloud-base/internal/dao"
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

func remove(l []string, item string) []string {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
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
