package reconcile

import (
	"context"
	"time"

	"github.com/paralus/paralus/pkg/common"
	"github.com/paralus/paralus/pkg/event"
	"github.com/paralus/paralus/pkg/query"
	"github.com/paralus/paralus/pkg/sentry/cryptoutil"
	"github.com/paralus/paralus/pkg/service"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/uptrace/bun"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

const (
	numClusterWorkers          = 3
	clusterEventHandleDuration = time.Second * 10
)

// ClusterEventHandler is the interface for handling cluster events
type ClusterEventHandler interface {
	// ClusterHook should be registred in cluster service to listen on
	// Cluster create/update/delete
	ClusterHook() event.Handler

	// ClusterWorkloadHook should be registered in cluster service to listen on
	// Task/Namespace/Workload ready events
	ClusterWorkloadHook() event.Handler
	// AntiEntropyHook should be used to handle anti entropy for
	// cluster
	AntiEntropyHook() event.Handler

	// Handle runs the placement event handler
	Handle(stop <-chan struct{})
}

type clusterEventHandler struct {
	cs service.ClusterService

	// cluster work queue
	cwq workqueue.RateLimitingInterface

	// cluster workload work queue
	wwq workqueue.RateLimitingInterface

	// required for cluster event reconciler
	db *bun.DB
	bs service.BootstrapService
	pf cryptoutil.PasswordFunc
}

// NewClusterEventHandler returns new cluster event handler
func NewClusterEventHandler(cs service.ClusterService, db *bun.DB, bs service.BootstrapService, pf cryptoutil.PasswordFunc) ClusterEventHandler {
	return &clusterEventHandler{
		cs:  cs,
		cwq: workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
		wwq: workqueue.NewRateLimitingQueue(workqueue.DefaultItemBasedRateLimiter()),
		db:  db,
		bs:  bs,
		pf:  pf,
	}
}

func (h *clusterEventHandler) ClusterHook() event.Handler {
	return event.HandlerFuncs{
		OnChangeFunc: func(r event.Resource) {
			h.cwq.Add(resourceToKey(r))
		},
	}
}

func (h *clusterEventHandler) ClusterWorkloadHook() event.Handler {
	return event.HandlerFuncs{
		OnChangeFunc: func(r event.Resource) {
			h.wwq.Add(resourceToKey(r))
		},
	}
}

func (h *clusterEventHandler) AntiEntropyHook() event.Handler {
	return event.HandlerFuncs{
		OnChangeFunc: func(r event.Resource) {
			h.cwq.Add(resourceToKey(r))
		},
	}
}

func (h *clusterEventHandler) processNextCluster() bool {
	item, shutdown := h.cwq.Get()
	if shutdown {
		return false
	}

	defer h.cwq.Done(item)

	ev := keyToResource(item.(string))

	h.handleClusterEvent(ev)

	return true
}

func (h *clusterEventHandler) runClusterWorker() {
	for h.processNextCluster() {
	}
}

func (h *clusterEventHandler) Handle(stop <-chan struct{}) {

	for i := 0; i < numClusterWorkers; i++ {
		go wait.Until(h.runClusterWorker, time.Second, stop)
		go wait.Until(h.runClusterWorkloadWorker, time.Second, stop)
	}

	<-stop
}

func (h *clusterEventHandler) processNextClusterWorkload() bool {
	item, shutdown := h.wwq.Get()
	if shutdown {
		return false
	}

	defer h.wwq.Done(item)

	ev := keyToResource(item.(string))

	h.handleClusterWorkloadEvent(ev)

	return true
}

func (h *clusterEventHandler) runClusterWorkloadWorker() {
	for h.processNextClusterWorkload() {
	}
}

func (h *clusterEventHandler) handleClusterEvent(ev event.Resource) {
	ctx, cancel := context.WithTimeout(context.Background(), clusterEventHandleDuration)
	defer cancel()

	var cluster *infrav3.Cluster
	var err error

	if ev.ID != "" {
		cluster, err = h.cs.Select(ctx, &infrav3.Cluster{
			Metadata: &commonv3.Metadata{Id: ev.ID, Project: ev.ProjectID},
		}, true)
	} else {

		cluster, err = h.cs.Get(ctx,
			query.WithName(ev.Name),
			query.WithPartnerID(ev.PartnerID),
			query.WithOrganizationID(ev.OrganizationID),
			query.WithProjectID(ev.ProjectID),
		)
	}

	if err != nil {
		_log.Infow("unable to get cluster for event", "event", ev, "error", err)
		return
	}

	//Update back the Ids
	cluster.Metadata.Project = ev.ProjectID
	cluster.Metadata.Organization = ev.OrganizationID
	cluster.Metadata.Partner = ev.PartnerID
	cluster.Metadata.Id = ev.ID

	ctx = context.WithValue(ctx, common.SessionDataKey, &commonv3.SessionData{
		Username: ev.Username,
		Account:  ev.Account,
	})

	_log.Debugw("handling cluster reconcile", "cluster", cluster.Metadata, "event", ev, "cluster status", cluster.Spec.ClusterData.ClusterStatus)

	reconciler := NewClusterReconciler(h.cs, h.db, h.bs, h.pf)
	err = reconciler.Reconcile(ctx, cluster)
	if err != nil {
		_log.Infow("unable to reconcile cluster", "error", err, "event", "ev")
		return
	}
	_log.Debugw("successfully reconciled cluster for event", "event", ev)
}

func (h *clusterEventHandler) handleClusterWorkloadEvent(ev event.Resource) {

	ctx, cancel := context.WithTimeout(context.Background(), clusterEventHandleDuration)
	defer cancel()

	var cluster *infrav3.Cluster
	var err error

	if ev.ID != "" {
		cluster, err = h.cs.Select(ctx, &infrav3.Cluster{
			Metadata: &commonv3.Metadata{Id: ev.ID, Project: ev.ProjectID},
		}, true)
	} else {
		cluster, err = h.cs.Get(ctx,
			query.WithName(ev.Name),
			query.WithPartnerID(ev.PartnerID),
			query.WithOrganizationID(ev.OrganizationID),
			query.WithProjectID(ev.ProjectID),
		)
	}

	if err != nil {
		_log.Infow("unable to get cluster for event", "event", ev, "error", err)
		return
	}

	_log.Debugw("handling cluster reconcile", "cluster", cluster.Metadata, "event", ev)

	reconciler := NewClusterConditionReconciler(h.cs)
	err = reconciler.Reconcile(ctx, cluster)
	if err != nil {
		_log.Infow("unable to reconcile cluster workload event", "error", err, "event", "ev")
		return
	}
	_log.Debugw("successfully reconciled cluster workload event", "event", ev)
}
