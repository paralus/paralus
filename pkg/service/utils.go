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

type projectRole struct {
	Project *string
	Role    string
}
type userProjectRoles map[uuid.UUID][]projectRole

func getUserProjectRoles(ctx context.Context, db bun.IDB) (userProjectRoles, error) {
	roles, err := dao.ListUserRoles(ctx, db)
	if err != nil {
		return userProjectRoles{}, err
	}

	upr := userProjectRoles{}
	for _, role := range roles {
		upr[role.AccountId] = append(upr[role.AccountId], projectRole{Project: role.Project, Role: role.Role})
	}

	return upr, nil
}

func projectAvailable(r []projectRole, projects []string) bool {
	// This is an OR internally
	// ALL is when the permissions is not project bound
	all := false
	if contains(projects, "ALL") {
		all = true
		projects = remove(projects, "ALL")
	}
	for _, pr := range r {
		if pr.Project != nil {
			if contains(projects, *pr.Project) {
				return true
			}
		} else if all {
			return true
		}
	}
	return false
}

func roleAvailable(r []projectRole, role string) bool {
	for _, pr := range r {
		if pr.Role == role {
			return true
		}
	}
	return false
}

func filterUserProjectRoles(upr userProjectRoles, projects []string, role string) (userProjectRoles, error) {
	fupr := userProjectRoles{}
	for u, r := range upr {
		if (len(projects) == 0 || projectAvailable(r, projects)) && (role == "" || roleAvailable(r, role)) {
			fupr[u] = r
		}
	}
	return fupr, nil
}
