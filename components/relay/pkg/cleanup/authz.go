package cleanup

import (
	"context"
	"strconv"
	"time"

	clientutil "github.com/RafaySystems/rcloud-base/components/common/pkg/controller/client"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

var (
	_log = log.GetLogger()
)

const (
	// how often should the cleanup routine run
	sweepInterval = time.Minute * 10
	// any authz which is olderThan this should be cleaned up
	olderThan = time.Hour * 8

	rafayRelayLabel   = "rafay-relay"
	authzRefreshLabel = "authz-refreshed"
	rafaySystemNS     = "rafay-system"

	maxResults = 50
)

func getMatchingLabelSelector() (sel client.MatchingLabelsSelector, err error) {
	var relayLabelReq, refreshLabelReq *labels.Requirement

	relayLabelReq, err = labels.NewRequirement(rafayRelayLabel, selection.Equals, []string{"true"})
	if err != nil {
		_log.Infow("unable to create relay label requirment", "error", err)
		return
	}

	refreshLabelReq, err = labels.NewRequirement(
		authzRefreshLabel,
		selection.LessThan,
		[]string{strconv.FormatInt(time.Now().Add(-olderThan).Unix(), 10)},
	)
	if err != nil {
		_log.Infow("unable to create authz refresh requirment", "error", err)
		return
	}
	sel.Selector = labels.NewSelector().Add(*relayLabelReq, *refreshLabelReq)

	return
}

func cleanServiceAccounts(ctx context.Context, c client.Client, sel client.MatchingLabelsSelector) {
	nCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var saList corev1.ServiceAccountList

	err := c.List(nCtx, &saList, sel, client.InNamespace(rafaySystemNS), client.Limit(maxResults))
	if err != nil {
		_log.Infow("unable to list service accounts", "error", err)
		return
	}

	for _, item := range saList.Items {
		_log.Infow("found stale service account", "name", item.Name)
		err = c.Delete(ctx, &item)
		if err != nil {
			_log.Infow("unable to delete stale service account", "name", item.Name, "error", err)
		}
	}

}

func cleanClusterRoleBindings(ctx context.Context, c client.Client, sel client.MatchingLabelsSelector) {
	nCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var crbList rbacv1.ClusterRoleBindingList

	err := c.List(nCtx, &crbList, sel, client.Limit(maxResults))
	if err != nil {
		_log.Infow("unable to list cluster role bindings", "error", err)
		return
	}

	for _, item := range crbList.Items {
		_log.Infow("found stale cluster role binding", "name", item.Name)
		err = c.Delete(ctx, &item)
		if err != nil {
			_log.Infow("unable to delete cluster role binding", "name", item.Name, "error", err)
		}
	}
}

func cleanRoleBindings(ctx context.Context, c client.Client, sel client.MatchingLabelsSelector) {
	nCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var rbList rbacv1.RoleBindingList

	err := c.List(nCtx, &rbList, sel, client.Limit(maxResults))
	if err != nil {
		_log.Infow("unable to list role bindings", "error", err)
		return
	}

	for _, item := range rbList.Items {
		_log.Infow("found stale role binding", "name", item.Name)
		err = c.Delete(ctx, &item)
		if err != nil {
			_log.Infow("unable to delete role binding", "name", item.Name, "error", err)
		}
	}
}

// StaleAuthz cleans up state authorizations provisioned in the cluster
func StaleAuthz(ctx context.Context) {
	c, err := clientutil.New()
	if err != nil {
		_log.Fatalw("unable to create client for cleaning stale authz entries", "error", err)
	}

	ticker := time.NewTicker(sweepInterval)
sweepLoop:
	for ; true; <-ticker.C {
		select {
		case <-ctx.Done():
			break sweepLoop
		default:
		}

		sel, err := getMatchingLabelSelector()
		if err != nil {
			_log.Infow("unable to build matching label selector", "error", err)
			continue
		}

		_log.Infow("finding authz matching", "selector", sel)

		cleanClusterRoleBindings(ctx, c, sel)
		cleanRoleBindings(ctx, c, sel)
		cleanServiceAccounts(ctx, c, sel)
	}
}
