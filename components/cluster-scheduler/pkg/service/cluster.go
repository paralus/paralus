package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper"
	clstrutil "github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/dao"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/models"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/patch"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	bun "github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _log = log.GetLogger()

var clusterNodeSyncMutexMap = make(map[string]*sync.Mutex)

const (
	clusterNotifyChan = "cluster:notify"
)

type ClusterService interface {
	Close() error
	// create Cluster
	Create(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	// get cluster
	Select(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	// get cluster
	Get(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error)
	// create or update cluster
	Update(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	// delete cluster
	Delete(ctx context.Context, cluster *infrav3.Cluster) error
	// list cluster
	List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error)
	//update cluster status
	UpdateClusterStatus(ctx context.Context, current *infrav3.Cluster) error
	//register cluster
	Register(ctx context.Context, token string) (*infrav3.ClusterToken, error)
	//listen clusters
	ListenClusters(ctx context.Context, mChan chan<- commonv3.Metadata)
}

// clusterService implements ClusterService
type clusterService struct {
	dao          pg.EntityDAO
	cns          ClusterNodesService
	cdao         dao.ClusterDao
	eobdao       dao.ClusterOperatorBootstrapDao
	pcdao        dao.ProjectClusterDao
	ctdao        dao.ClusterTokenDao
	downloadData bootstrapper.DownloadData
}

// NewClusterService return new cluster service
func NewClusterService(db *bun.DB, data *bootstrapper.DownloadData) ClusterService {
	entityDao := pg.NewEntityDAO(db)
	return &clusterService{
		dao:          entityDao,
		cns:          NewClusterNodesService(db),
		cdao:         dao.NewClusterDao(entityDao),
		eobdao:       dao.NewClusterOperatorBootstrapDao(entityDao),
		pcdao:        dao.NewProjectClusterDao(entityDao),
		ctdao:        dao.NewClusterTokenDao(entityDao),
		downloadData: *data,
	}
}

func (es *clusterService) Create(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	var errormsg string

	partId, err := uuid.Parse(cluster.Metadata.Partner)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, err
	}

	orgId, err := uuid.Parse(cluster.Metadata.Organization)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, err
	}

	projId, err := uuid.Parse(cluster.Metadata.Project)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, err
	}

	/*reqAuth, err := c.requestAuth(r, ctx, ps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		returnrequestAuth
	}*/

	if cluster.Metadata.Name == "" {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid cluster data, Name is missing.",
		}
		return cluster, fmt.Errorf("invalid cluster data, name is missing")
	}

	if cluster.Spec.ClusterType == "" {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid cluster data, Cluster type is missing.",
		}
		return cluster, fmt.Errorf("invalid cluster data, cluster type is missing")
	}

	clusterGeneration, err := clstrutil.GetClusterGeneration(cluster.Spec.ClusterType)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid cluster data, Cluster generation is invalid.",
		}
		return cluster, fmt.Errorf("invalid cluster data, cluster generation is invalid")
	}

	if !clstrutil.HasValidCharacters(strings.ToLower(cluster.Metadata.Name)) {
		errormsg = "cluster name contains invalid characters. valid characters are `[A-Z][a-z][0-9]-`"
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          errormsg,
		}
		return cluster, fmt.Errorf(errormsg)
	}
	if len(cluster.Metadata.Name) > 63 {
		errormsg = "maximum characters allowed for cluster name is 63. please try another name"
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          errormsg,
		}
		return cluster, fmt.Errorf(errormsg)
	}

	clusterPresent, err := es.dao.GetEntityByName(ctx, cluster.Metadata.Name, uuid.NullUUID{UUID: orgId, Valid: true},
		uuid.NullUUID{UUID: partId, Valid: true}, &models.Cluster{})
	if err != nil && err.Error() == "sql: no rows in result set" {
		_log.Infof("Skipping as first time cluster create ")
	} else if clusterPresent != nil {
		errormsg = "cluster name is already taken. please try another name"
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          errormsg,
		}
		return cluster, fmt.Errorf(errormsg)
	}

	metro := &models.Metro{}
	if cluster.Spec.Metro != nil && cluster.Spec.Metro.Name != "" {
		if mdb, err := es.dao.GetEntityByName(ctx, cluster.Spec.Metro.Name, uuid.NullUUID{UUID: orgId, Valid: true}, uuid.NullUUID{UUID: partId, Valid: true}, metro); err != nil {
			errormsg = "Invalid cluster location, provide a valid metro name"
			cluster.Status = &commonv3.Status{
				ConditionType:   "Create",
				ConditionStatus: commonv3.ConditionStatus_StatusFailed,
				Reason:          errormsg,
			}
			return cluster, err
		} else {
			metro = mdb.(*models.Metro)
		}
	}

	// Labels should have been populated by now. Perform last minute validations and fixups
	if len(cluster.Metadata.Labels) == 0 {
		cluster.Metadata.Labels = make(map[string]string)
	}
	if err := clstrutil.ValidateCustomLabels(cluster.Metadata.Labels); err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, err
	}
	clusterLabels := clstrutil.ExtractV2ClusterLabels(cluster.Metadata.Labels, nil, cluster.Metadata.Name, cluster.Spec.ClusterType, metro.Name)

	if clusterLabels == nil {
		clusterLabels = make(map[string]string)
	}
	clusterLabels[constants.ClusterLabelKey] = cluster.Metadata.Name
	lbsBytes, _ := json.Marshal(clusterLabels)

	edb := &models.Cluster{
		Name:           strings.ToLower(cluster.Metadata.Name),
		CreatedAt:      time.Now(),
		ModifiedAt:     time.Now(),
		Trash:          false,
		DisplayName:    cluster.Metadata.Name,
		ProjectId:      projId,
		ShareMode:      infrav3.ClusterShareMode_CUSTOM.String(),
		OrganizationId: orgId,
		PartnerId:      partId,
		MetroId:        metro.ID,
		Labels:         json.RawMessage(lbsBytes),
		BlueprintRef:   constants.DefaultBlueprint,
		ClusterType:    cluster.Spec.ClusterType,
	}

	if len(cluster.Metadata.Annotations) > 0 {
		annBytes, _ := json.Marshal(cluster.Metadata.Annotations)
		edb.Annotations = json.RawMessage(annBytes)
	}
	if cluster.Spec.ProxyConfig != nil {
		pcfgsByts, _ := json.Marshal(cluster.Spec.ProxyConfig)
		edb.ProxyConfig = json.RawMessage(pcfgsByts)
	}

	cluster.Spec.ClusterData = &infrav3.ClusterData{
		ClusterStatus: &infrav3.ClusterStatus{
			Conditions: clstrutil.DefaultClusterConditions,
		},
	}
	clstrutil.SetClusterCondition(cluster, clstrutil.NewClusterBootstrapAgent(commonv3.RafayConditionStatus_Pending, "created"))
	cnds, _ := json.Marshal(clstrutil.DefaultClusterConditions)
	edb.Conditions = json.RawMessage(cnds)

	cluster.Spec.ClusterData.Health = infrav3.Health_EDGE_IGNORE

	err = es.cdao.CreateCluster(ctx, edb)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, err
	}
	// if project is set create project cluster
	var pc *models.ProjectCluster
	pcList := make([]models.ProjectCluster, 0)
	if edb.ProjectId != uuid.Nil {
		pc = &models.ProjectCluster{
			ProjectID: edb.ProjectId,
			ClusterID: edb.ID,
		}
		err = es.pcdao.CreateProjectCluster(ctx, pc)
		if err != nil {
			cluster.Status = &commonv3.Status{
				ConditionType:   "Create",
				ConditionStatus: commonv3.ConditionStatus_StatusFailed,
				Reason:          err.Error(),
			}
			return cluster, err
		}
		pcList = append(pcList, *pc)
	}
	_log.Infow("Created the cluster: ", "Cluster", edb)

	//TODO: do we need create nodes while creating cluster or bootstrapping ?
	/*// add nodes
	for _, node := range cluster.Spec.ClusterData.ClusterStatus.Nodes {
		err := es.cndao.CreateOrUpdateNode(edb, &models.ClusterNodes{
			ClusterId:      edb.ID,
			OrganizationId: edb.OrganizationId,
			PartnerId:      edb.PartnerId,
			ProjectId:      edb.ProjectId,
			Name:           node.Metadata.Name,
			DisplayName:    node.Metadata.Name,
			Unschedulable:  node.Spec.Unschedulable,
			//Labels: json.RawMessage(node.Metadata.Labels[]),
			//TODO: lot more like taints and annotations etc
		})
		if err != nil {
			return nil, err
		}
	}*/

	clusterResp := prepareClusterResponse(cluster, edb, *metro, pcList, nil)

	if clusterGeneration == constants.Cluster_V2 && edb.PartnerId != uuid.Nil && edb.OrganizationId != uuid.Nil {
		operatorSpecStr, err := clstrutil.GetClusterOperatorYaml(ctx, &es.downloadData, clusterResp)
		if err != nil {
			_log.Errorw("Error downloading v2 cluster operator yaml", "Error", err)
			cluster.Status = &commonv3.Status{
				ConditionType:   "Create",
				ConditionStatus: commonv3.ConditionStatus_StatusFailed,
				Reason:          err.Error(),
			}
			return cluster, err
		}
		//fmt.Println(resp)
		_log.Infow("Creating cluster operator yaml", "clusterid", edb.ID)
		operatorSpecEncoded := base64.StdEncoding.EncodeToString([]byte(operatorSpecStr))
		boostrapData := models.ClusterOperatorBootstrap{
			ClusterId:   edb.ID,
			YamlContent: operatorSpecEncoded,
		}
		es.eobdao.CreateOperatorBootstrap(ctx, &boostrapData)
	}
	//TODO: while integrating events framework
	//CreateClusterEvent(c, "cluster.create.success", e, reqAuth.projectID, r, reqAuth.sd)

	return cluster, nil
}

func (s *clusterService) Select(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	clstr := &infrav3.Cluster{
		ApiVersion: constants.ApiVersion,
		Kind:       constants.ClusterKind,
	}

	id, err := uuid.Parse(cluster.Metadata.Id)
	if err != nil {
		id = uuid.Nil
	}
	c, err := s.cdao.GetCluster(ctx, &models.Cluster{ID: id, Name: cluster.Metadata.Name})
	if err != nil {
		clstr.Status = &commonv3.Status{
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return clstr, err
	}
	nodes, err := s.cns.GetClusterNodes(ctx, c.ID)
	if err != nil {
		clstr.Status = &commonv3.Status{
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return clstr, err
	}
	projects, err := s.pcdao.GetProjectsForCluster(ctx, c.ID)
	if err != nil {
		clstr.Status = &commonv3.Status{
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return clstr, err
	}

	entity, err := s.dao.GetByID(ctx, c.MetroId, &models.Metro{})
	if err != nil {
		_log.Errorf("failed to fetch metro details", err)
	}
	metro := entity.(*models.Metro)

	//TODO: Get cluster workload information

	clstr = prepareClusterResponse(clstr, c, *metro, projects, nodes)

	return clstr, nil
}

func (s *clusterService) Get(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
	queryOptions := commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(&queryOptions)
	}

	cluster := &infrav3.Cluster{
		Metadata: &commonv3.Metadata{
			Name:         queryOptions.Name,
			Partner:      queryOptions.PartnerID,
			Organization: queryOptions.OrganizationID,
			Project:      queryOptions.ProjectID,
			Id:           queryOptions.ID,
		},
	}
	return s.Select(ctx, cluster)
}

func prepareClusterResponse(clstr *infrav3.Cluster, c *models.Cluster, metro models.Metro, projects []models.ProjectCluster, nodes []infrav3.ClusterNode) *infrav3.Cluster {

	var lbls map[string]string
	if c.Labels != nil {
		json.Unmarshal(c.Labels, lbls)
	}
	var ann map[string]string
	if c.Annotations != nil {
		json.Unmarshal(c.Annotations, ann)
	}
	clstr.Metadata = &commonv3.Metadata{
		Name:         c.Name,
		Description:  c.DisplayName,
		Labels:       lbls,
		Annotations:  ann,
		Project:      c.ProjectId.String(),
		Organization: c.OrganizationId.String(),
		Partner:      c.PartnerId.String(),
		Id:           c.ID.String(),
		ModifiedAt:   timestamppb.New(c.ModifiedAt),
	}
	sm, _ := strconv.Atoi(c.ShareMode)
	var proxy infrav3.ProxyConfig
	if c.ProxyConfig != nil {
		json.Unmarshal(c.ProxyConfig, proxy)
	}
	var pcs []*infrav3.ProjectCluster
	if len(projects) > 0 {
		pcs = make([]*infrav3.ProjectCluster, len(projects)-1)
		for _, pc := range projects {
			pcs = append(pcs, &infrav3.ProjectCluster{
				ProjectID: pc.ProjectID.String(),
				ClusterID: pc.ClusterID.String(),
			})
		}
	}
	var nds []*infrav3.ClusterNode
	if len(nodes) > 0 {
		for _, node := range nodes {
			nds = append(nds, &node)
		}
		clstr.Spec.ClusterData.Nodes = nds
	}
	var conditions []*infrav3.ClusterCondition
	if c.Conditions != nil {
		json.Unmarshal(c.Conditions, &conditions)
	}
	clstr.Spec = &infrav3.ClusterSpec{
		ClusterType: c.ClusterType,
		Metro: &infrav3.Metro{
			Name:    metro.Name,
			City:    metro.City,
			State:   metro.State,
			Country: metro.Country,
		},
		OverrideSelector: c.OverrideSelector,
		ShareMode:        infrav3.ClusterShareMode(sm),
		ProxyConfig:      &proxy,
		ClusterData: &infrav3.ClusterData{
			ClusterBlueprint: c.BlueprintRef,
			Projects:         pcs,
			Nodes:            nds,
			ClusterStatus: &infrav3.ClusterStatus{
				Conditions:         conditions,
				Token:              c.Token,
				PublishedBlueprint: c.BlueprintRef,
			},
		},
	}
	clstr.Status = &commonv3.Status{
		ConditionStatus: commonv3.ConditionStatus_StatusOK,
		LastUpdated:     timestamppb.New(c.ModifiedAt),
	}
	return clstr
}

func (cs *clusterService) Update(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {

	var errormsg string

	if cluster.Metadata.Name == "" {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Update",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid cluster data, Name is missing.",
		}
		return cluster, fmt.Errorf("invalid cluster data, name is missing")
	}

	edb, err := cs.dao.GetByName(ctx, cluster.Metadata.Name, &models.Cluster{})
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Update",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, fmt.Errorf(errormsg)
	}
	cdb := edb.(*models.Cluster)

	oid := cdb.OrganizationId
	pid := cdb.PartnerId

	if cluster.Spec.ClusterType == "" {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Update",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid cluster data, Cluster type is missing.",
		}
		return cluster, fmt.Errorf("invalid cluster data, cluster type is missing")
	}

	_, err = clstrutil.GetClusterGeneration(cluster.Spec.ClusterType)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Update",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid cluster data, Cluster generation is invalid.",
		}
		return cluster, fmt.Errorf("invalid cluster data, cluster generation is invalid")
	}

	if len(cluster.Metadata.Labels) == 0 {
		cluster.Metadata.Labels = make(map[string]string)
	}
	if err := clstrutil.ValidateCustomLabels(cluster.Metadata.Labels); err != nil {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
		}
		return cluster, err
	}
	clusterLabels := clstrutil.ExtractV2ClusterLabels(cluster.Metadata.Labels, nil, cluster.Metadata.Name, cluster.Spec.ClusterType, "")

	if clusterLabels == nil {
		clusterLabels = make(map[string]string)
	}
	clusterLabels[constants.ClusterLabelKey] = cluster.Metadata.Name
	lbsBytes, _ := json.Marshal(clusterLabels)

	//update editable fields
	cdb.ModifiedAt = time.Now()
	cdb.OverrideSelector = cluster.Spec.OverrideSelector
	cdb.ShareMode = cluster.Spec.ShareMode.String()
	cdb.Labels = json.RawMessage(lbsBytes)

	//location of cluster is updated
	if cluster.Spec.Metro != nil && cdb.MetroId.String() != cluster.Spec.Metro.Id {
		metro := &models.Metro{}
		if cluster.Spec.Metro.Name != "" {
			if mdb, err := cs.dao.GetEntityByName(ctx, cluster.Spec.Metro.Name, uuid.NullUUID{UUID: oid, Valid: true}, uuid.NullUUID{UUID: pid, Valid: true}, metro); err != nil {
				errormsg = "Invalid cluster location, provide a valid metro name"
				cluster.Status = &commonv3.Status{
					ConditionType:   "Update",
					ConditionStatus: commonv3.ConditionStatus_StatusFailed,
					Reason:          errormsg,
				}
				return cluster, err
			} else {
				metro = mdb.(*models.Metro)
			}
			cdb.MetroId = metro.ID
		}
	}
	if len(cluster.Metadata.Annotations) > 0 {
		annBytes, _ := json.Marshal(cluster.Metadata.Annotations)
		cdb.Annotations = json.RawMessage(annBytes)
	}
	if cluster.Spec.ProxyConfig != nil {
		pcfgsByts, _ := json.Marshal(cluster.Spec.ProxyConfig)
		cdb.ProxyConfig = json.RawMessage(pcfgsByts)
	}

	if cluster.Spec.ClusterData != nil {
		cdb.BlueprintRef = cluster.Spec.ClusterData.ClusterBlueprint
		if cluster.Spec.ClusterData.ClusterStatus != nil {
			cdb.PublishedBlueprint = cluster.Spec.ClusterData.ClusterStatus.PublishedBlueprint
		}

		for _, node := range cluster.Spec.ClusterData.Nodes {
			node.Metadata.Project = cluster.Metadata.Project
			node.Metadata.Organization = cluster.Metadata.Organization
			node.Metadata.Partner = cluster.Metadata.Partner
			cs.cns.CreateOrUpdateNode(ctx, uuid.MustParse(cluster.Metadata.Id), node)
		}
	}
	err = cs.cdao.UpdateCluster(ctx, cdb)
	if err != nil {
		cluster.Status = &commonv3.Status{
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			ConditionType:   "Update",
			LastUpdated:     timestamppb.New(cdb.ModifiedAt),
		}
		return cluster, err
	}

	cluster.Status = &commonv3.Status{
		ConditionStatus: commonv3.ConditionStatus_StatusOK,
		ConditionType:   "Update",
		LastUpdated:     timestamppb.New(cdb.ModifiedAt),
	}

	return cluster, nil

	//TODO: revisit later
	/*notifyCluster(s.db.WithContext(ctx), c)

	ev := event.Resource{
		PartnerID:      c.PartnerID,
		OrganizationID: c.OrganizationID,
		ProjectID:      c.ProjectID,
		Name:           c.Name,
		EventType:      event.ResourceDelete,
		ID:             c.ID,
	}

	// for _, h := range s.clusterHandlers {
	// 	h.OnChange(ev)
	// }
	for _, h := range s.placementHandlers {
		h.OnChange(ev)
	}*/
}

func (cs *clusterService) Delete(ctx context.Context, cluster *infrav3.Cluster) error {
	//TODO:
	/*err = bootstrapagent.DeleteForCluster(ctx, s.sentryPool, s.ClusterService, query.WithMeta(cluster))
	if err != nil {
		return nil, err
	}*/

	cluster, err := cs.Select(ctx, cluster)
	if err != nil {
		return err
	}

	_log.Infow("deleting cluster", "name", cluster.Metadata.Name)

	_log.Debugw("setting cluster condition to pending delete", "name", cluster.Metadata.Name, "conditions", cluster.Spec.ClusterData.ClusterStatus.Conditions)
	clstrutil.SetClusterCondition(cluster, clstrutil.NewClusterDelete(constants.Pending, "deleted"))

	err = cs.UpdateClusterStatus(ctx, cluster)
	if err != nil {
		return errors.Wrapf(err, "could not update cluster %s status to pending delete", cluster.Metadata.Name)
	}

	_log.Infow("updated cluster status to pending delete", "name", cluster.Metadata.Name)

	return nil

}

func (cs *clusterService) deleteCluster(ctx context.Context, cluster *infrav3.Cluster) error {

	c := models.Cluster{
		ID:        uuid.MustParse(cluster.Metadata.Id),
		ProjectId: uuid.MustParse(cluster.Metadata.Project),
	}
	err := cs.pcdao.DeleteProjectsForCluster(ctx, uuid.MustParse(cluster.Metadata.Id))
	if err != nil {
		return errors.Wrapf(err, "could not delete projects for cluster %s", cluster.Metadata.Name)
	}
	return cs.cdao.DeleteCluster(ctx, &c)
}

func (cs *clusterService) List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error) {

	queryOptions := commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(&queryOptions)
	}

	cdb, err := cs.cdao.ListClusters(ctx, queryOptions)
	if err != nil {
		return nil, err
	}

	clusters := infrav3.ClusterList{
		ApiVersion: constants.ApiVersion,
		Kind:       constants.ClusterListKind,
		Metadata: &commonv3.ListMetadata{
			Count:  int64(len(cdb)),
			Offset: queryOptions.Offset,
			Limit:  queryOptions.Limit,
		},
	}

	var items []*infrav3.Cluster
	for _, clstr := range cdb {
		nodes, err := cs.cns.GetClusterNodes(ctx, clstr.ID)
		if err != nil {
			return nil, err
		}
		projects, err := cs.pcdao.GetProjectsForCluster(ctx, clstr.ID)
		if err != nil {
			return nil, err
		}
		entity, err := cs.dao.GetByID(ctx, clstr.MetroId, &models.Metro{})
		if err != nil {
			return nil, err
		}
		metro := entity.(*models.Metro)
		//TODO: workload related stuff pending
		cluster := prepareClusterResponse(&infrav3.Cluster{}, &clstr, *metro, projects, nodes)
		items = append(items, cluster)
	}

	clusters.Items = items

	return &clusters, nil
}

func (s *clusterService) Register(ctx context.Context, token string) (*infrav3.ClusterToken, error) {

	var clusterToken *infrav3.ClusterToken

	cluster := &infrav3.Cluster{
		ApiVersion: constants.ApiVersion,
		Kind:       constants.ClusterKind,
	}

	s.dao.GetInstance().RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		ct, err := s.ctdao.RegisterToken(ctx, token)
		if err != nil {
			_log.Infow("Token Not registered", "token", token, "error", err)
			return err
		}
		cdb, err := s.cdao.GetClusterForToken(ctx, token)
		if err != nil {
			return err
		}
		cluster = &infrav3.Cluster{}
		cluster = prepareClusterResponse(cluster, cdb, models.Metro{}, nil, nil)
		cluster.Spec.ClusterData.ClusterStatus.Conditions = []*infrav3.ClusterCondition{
			clstrutil.NewClusterRegister(commonv3.RafayConditionStatus_Success, "registered"),
		}

		err = s.UpdateClusterStatus(ctx, cluster)
		if err != nil {
			return err
		}
		tt, _ := strconv.Atoi(ct.TokenType)
		clusterToken = &infrav3.ClusterToken{
			Spec: &infrav3.ClusterTokenSpec{
				TokenType: infrav3.ClusterTokenType(tt),
			},
			Metadata: &commonv3.Metadata{
				Id:           ct.ID.String(),
				Project:      ct.ProjectId.String(),
				Organization: ct.OrganizationId.String(),
				Partner:      ct.PartnerId.String(),
			},
		}
		return nil
	})
	/*
		ev := event.Resource{
			EventType: event.ResourceUpdateStatus,
			ID:        c.ID,
		}

		for _, h := range s.clusterHandlers {
			h.OnChange(ev)
		}*/

	//notifyCluster(s.db.WithContext(ctx), c.Cluster)
	return clusterToken, nil
}

func (s *clusterService) UpdateClusterStatus(ctx context.Context, current *infrav3.Cluster) error {

	existing, err := s.Select(ctx, current)
	if err != nil {
		return err
	}

	_log.Debugw("existing cluster conditions", "name", existing.Spec.ClusterData.ClusterStatus.Conditions)
	_log.Debugw("retrieved cluster to update cluster status", "name", existing.Metadata.Name)
	_log.Debugw("current cluster conditions", "name", current.Spec.ClusterData.ClusterStatus.Conditions)

	err = patch.ClusterStatus(existing.Spec.ClusterData.ClusterStatus, current.Spec.ClusterData.ClusterStatus)
	if err != nil {
		return err
	}

	_log.Debugw("updated cluster conditions", "name", existing.Spec.ClusterData.ClusterStatus.Conditions)
	_log.Debugw("updated cluster object", "name", existing.Metadata.Name)

	//update the cluster
	_, err = s.Update(ctx, existing)

	_log.Debugw("updated cluster in db", "name", existing.Metadata.Name)

	if clstrutil.IsClusterDeleted(existing) {
		err = s.deleteCluster(ctx, existing)
		if err != nil {
			return err
		}
		_log.Debugw("deleted cluster in db", "name", existing.Metadata.Name)
	}

	// set current to patched existing
	*current = *existing

	return err
}

func (s *clusterService) ListenClusters(ctx context.Context, mChan chan<- commonv3.Metadata) {
	listener := pgdriver.NewListener(s.dao.GetInstance())
	listener.Listen(ctx, clusterNotifyChan)
	notifyChan := listener.Channel()
listenerLoop:
	for {
		select {
		case <-ctx.Done():
			break listenerLoop
		case n, ok := <-notifyChan:
			if !ok {
				break listenerLoop
			}

			var meta commonv3.Metadata
			err := json.Unmarshal([]byte(n.Payload), &meta)
			if err != nil {
				_log.Infow("unable to unmarshal cluster notification", "error", err)
				continue

			}

			select {
			case mChan <- meta:
			default:
			}

		}
	}

}

func (s *clusterService) Close() error {
	return s.dao.Close()
}
