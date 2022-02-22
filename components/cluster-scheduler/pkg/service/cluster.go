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

	clstrutil "github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/cluster"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/cluster/dao"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/cluster/util"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/internal/models"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper"
	bootstrapagent "github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper/agent"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/patch"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/event"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	cmodels "github.com/RafaySystems/rcloud-base/components/common/pkg/models"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/scheduler"
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
	Select(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error)
	// get cluster
	Get(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error)
	// create or update cluster
	Update(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error)
	// delete cluster
	Delete(ctx context.Context, cluster *infrav3.Cluster) error
	// list cluster
	List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error)
	//update cluster status
	UpdateClusterConditionStatus(ctx context.Context, current *infrav3.Cluster) error
	// update cluster annotations
	UpdateClusterAnnotations(ctx context.Context, cluster *infrav3.Cluster) error
	//register cluster
	Register(ctx context.Context, token string) (*infrav3.ClusterToken, error)
	//listen clusters
	ListenClusters(ctx context.Context, mChan chan<- commonv3.Metadata)
	//Get cluster projects
	GetClusterProjects(ctx context.Context, cluster *infrav3.Cluster) ([]models.ProjectCluster, error)
	//Validate and update cluster status
	UpdateStatus(ctx context.Context, current *infrav3.Cluster, opts ...query.Option) error
	// Get Namespaces for cluster and conditions
	GetNamespacesForConditions(ctx context.Context, conditions []scheduler.ClusterNamespaceCondition, clusterID string) (*scheduler.ClusterNamespaceList, error)
	// Get Namespaces for given cluster
	GetNamespaces(ctx context.Context, clusterID string) (*scheduler.ClusterNamespaceList, error)
	// Get Namespace
	GetNamespace(ctx context.Context, namespace string, clusterID string) (*scheduler.ClusterNamespace, error)
	// Update Namespace Status
	UpdateNamespaceStatus(ctx context.Context, current *scheduler.ClusterNamespace) error
	// Get Namespace hashes
	GetNamespaceHashes(ctx context.Context, clusterID string) ([]infrav3.NameHash, error)
	//Add event handlers
	AddEventHandler(evh event.Handler)
}

// clusterService implements ClusterService
type clusterService struct {
	dao             pg.EntityDAO
	adao            pg.EntityDAO
	cdao            dao.ClusterDao
	eobdao          dao.ClusterOperatorBootstrapDao
	pcdao           dao.ProjectClusterDao
	ctdao           dao.ClusterTokenDao
	cndao           dao.ClusterNamespacesDao
	downloadData    bootstrapper.DownloadData
	clusterHandlers []event.Handler
	sentryPool      sentryrpc.SentryPool
}

// NewClusterService return new cluster service
func NewClusterService(db *bun.DB, adb *bun.DB, data *bootstrapper.DownloadData, sentrypool sentryrpc.SentryPool) ClusterService {
	entityDao := pg.NewEntityDAO(db)
	return &clusterService{
		dao:          entityDao,
		adao:         pg.NewEntityDAO(adb),
		cdao:         dao.NewClusterDao(entityDao),
		eobdao:       dao.NewClusterOperatorBootstrapDao(entityDao),
		pcdao:        dao.NewProjectClusterDao(entityDao),
		ctdao:        dao.NewClusterTokenDao(entityDao),
		cndao:        dao.NewClusterNamespacesDao(entityDao),
		downloadData: *data,
		sentryPool:   sentrypool,
	}
}

func (es *clusterService) Create(ctx context.Context, cluster *infrav3.Cluster) (*infrav3.Cluster, error) {
	var errormsg string
	if cluster.Metadata.Project == "" {
		cluster.Status = &commonv3.Status{
			ConditionType:   "Create",
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          "Invalid project association, project is missing.",
		}
		return cluster, fmt.Errorf("invalid cluster data, project is missing")
	}

	var proj cmodels.Project
	_, err := es.adao.GetByName(ctx, cluster.Metadata.Project, &proj)
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

	clusterPresent, err := es.dao.GetByNamePartnerOrg(ctx, cluster.Metadata.Name, uuid.NullUUID{UUID: proj.PartnerId, Valid: true},
		uuid.NullUUID{UUID: proj.OrganizationId, Valid: true}, &models.Cluster{})
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
		if mdb, err := es.dao.GetByNamePartnerOrg(ctx, cluster.Spec.Metro.Name, uuid.NullUUID{UUID: proj.PartnerId, Valid: true}, uuid.NullUUID{UUID: uuid.Nil, Valid: false}, metro); err != nil {
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
	if err := util.ValidateCustomLabels(cluster.Metadata.Labels); err != nil {
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
		ProjectId:      proj.ID,
		ShareMode:      infrav3.ClusterShareMode_CUSTOM.String(),
		OrganizationId: proj.OrganizationId,
		PartnerId:      proj.PartnerId,
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
	fmt.Print("conditions while create :: ", clstrutil.DefaultClusterConditions)
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

	clusterResp := es.prepareClusterResponse(ctx, cluster, edb, metro, pcList, true)

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
		_log.Infow("Creating cluster operator yaml", "clusterid", edb.ID)
		operatorSpecEncoded := base64.StdEncoding.EncodeToString([]byte(operatorSpecStr))
		boostrapData := models.ClusterOperatorBootstrap{
			ClusterId:   edb.ID,
			YamlContent: operatorSpecEncoded,
		}
		es.eobdao.CreateOperatorBootstrap(ctx, &boostrapData)
	}

	ev := event.Resource{
		PartnerID:      edb.PartnerId.String(),
		OrganizationID: edb.OrganizationId.String(),
		ProjectID:      edb.ProjectId.String(),
		Name:           edb.Name,
		EventType:      event.ResourceCreate,
		ID:             edb.ID.String(),
	}

	for _, h := range es.clusterHandlers {
		h.OnChange(ev)
	}

	return cluster, nil
}

func (s *clusterService) Select(ctx context.Context, cluster *infrav3.Cluster, isExtended bool) (*infrav3.Cluster, error) {

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
	var projects []models.ProjectCluster
	if isExtended {
		projects, err = s.pcdao.GetProjectsForCluster(ctx, c.ID)
		if err != nil {
			clstr.Status = &commonv3.Status{
				ConditionStatus: commonv3.ConditionStatus_StatusFailed,
				Reason:          err.Error(),
				LastUpdated:     timestamppb.Now(),
			}
			return clstr, err
		}
	}

	var metro *models.Metro
	if c.MetroId != uuid.Nil {
		entity, err := s.dao.GetByID(ctx, c.MetroId, &models.Metro{})
		if err != nil {
			_log.Errorf("failed to fetch metro details", err)
		}
		metro = entity.(*models.Metro)
	}

	//TODO: Get cluster workload information
	clstr = s.prepareClusterResponse(ctx, clstr, c, metro, projects, isExtended)

	return clstr, nil
}

func (s *clusterService) Get(ctx context.Context, opts ...query.Option) (*infrav3.Cluster, error) {
	queryOptions := commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(&queryOptions)
	}

	clstr := &infrav3.Cluster{
		ApiVersion: constants.ApiVersion,
		Kind:       constants.ClusterKind,
	}

	id, err := uuid.Parse(queryOptions.ClusterID)
	if err != nil {
		id = uuid.Nil
	}
	c, err := s.cdao.GetCluster(ctx, &models.Cluster{ID: id, Name: queryOptions.Name})
	if err != nil {
		clstr.Status = &commonv3.Status{
			ConditionStatus: commonv3.ConditionStatus_StatusFailed,
			Reason:          err.Error(),
			LastUpdated:     timestamppb.Now(),
		}
		return clstr, err
	}
	var projects []models.ProjectCluster
	if queryOptions.Extended {
		projects, err = s.pcdao.GetProjectsForCluster(ctx, c.ID)
		if err != nil {
			clstr.Status = &commonv3.Status{
				ConditionStatus: commonv3.ConditionStatus_StatusFailed,
				Reason:          err.Error(),
				LastUpdated:     timestamppb.Now(),
			}
			return clstr, err
		}
	}

	s.prepareClusterResponse(ctx, clstr, c, nil, projects, queryOptions.Extended)
	clstr.Metadata.Id = c.ID.String()
	clstr.Metadata.Project = c.ProjectId.String()
	clstr.Metadata.Organization = c.OrganizationId.String()
	clstr.Metadata.Partner = c.PartnerId.String()

	return clstr, nil
}

func (s *clusterService) prepareClusterResponse(ctx context.Context, clstr *infrav3.Cluster, c *models.Cluster, metro *models.Metro, projects []models.ProjectCluster, isExtended bool) *infrav3.Cluster {

	var part cmodels.Partner
	_, err := s.adao.GetNameById(ctx, c.PartnerId, &part)
	if err != nil {
		_log.Infow("unable to fetch partner information, ", err.Error())
	}

	var org cmodels.Organization
	_, err = s.adao.GetNameById(ctx, c.OrganizationId, &org)
	if err != nil {
		_log.Infow("unable to fetch organization information, ", err.Error())
	}

	var proj cmodels.Project
	_, err = s.adao.GetNameById(ctx, c.ProjectId, &proj)
	if err != nil {
		_log.Infow("unable to fetch project information, ", err.Error())
	}

	var lbls map[string]string
	if isExtended && c.Labels != nil {
		json.Unmarshal(c.Labels, &lbls)
	}
	var ann map[string]string
	if isExtended && c.Annotations != nil {
		json.Unmarshal(c.Annotations, &ann)
	}
	clstr.Metadata = &commonv3.Metadata{
		Name:         c.Name,
		Description:  c.DisplayName,
		Labels:       lbls,
		Annotations:  ann,
		Project:      proj.Name,
		Organization: org.Name,
		Partner:      part.Name,
		ModifiedAt:   timestamppb.New(c.ModifiedAt),
	}

	sm, _ := strconv.Atoi(c.ShareMode)
	var proxy infrav3.ProxyConfig
	if c.ProxyConfig != nil {
		json.Unmarshal(c.ProxyConfig, &proxy)
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
	var conditions []*infrav3.ClusterCondition
	if isExtended && c.Conditions != nil {
		json.Unmarshal(c.Conditions, &conditions)
	}
	clstr.Spec = &infrav3.ClusterSpec{
		ClusterType:      c.ClusterType,
		OverrideSelector: c.OverrideSelector,
		ShareMode:        infrav3.ClusterShareMode(sm),
		ProxyConfig:      &proxy,
		ClusterData: &infrav3.ClusterData{
			ClusterBlueprint: c.BlueprintRef,
			Projects:         pcs,
			ClusterStatus: &infrav3.ClusterStatus{
				Conditions:         conditions,
				Token:              c.Token,
				PublishedBlueprint: c.BlueprintRef,
			},
		},
	}
	if metro != nil {
		clstr.Spec.Metro = &infrav3.Metro{
			Name:    metro.Name,
			City:    metro.City,
			State:   metro.State,
			Country: metro.Country,
		}
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

	if len(cluster.Metadata.Annotations) > 0 {
		annBytes, _ := json.Marshal(cluster.Metadata.Annotations)
		cdb.Annotations = json.RawMessage(annBytes)
	}

	//location of cluster is updated
	if cluster.Spec.Metro != nil && cdb.MetroId.String() != cluster.Spec.Metro.Id {
		metro := &models.Metro{}
		if cluster.Spec.Metro.Name != "" {
			if mdb, err := cs.dao.GetByNamePartnerOrg(ctx, cluster.Spec.Metro.Name, uuid.NullUUID{UUID: pid, Valid: true}, uuid.NullUUID{UUID: uuid.Nil, Valid: false}, metro); err != nil {
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

			if len(cluster.Spec.ClusterData.ClusterStatus.Conditions) > 0 {
				cndBytes, _ := json.Marshal(cluster.Spec.ClusterData.ClusterStatus.Conditions)
				cdb.Conditions = json.RawMessage(cndBytes)
			}
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

	cs.notifyCluster(ctx, cluster)

	ev := event.Resource{
		PartnerID:      cluster.Metadata.Partner,
		OrganizationID: cluster.Metadata.Organization,
		ProjectID:      cluster.Metadata.Project,
		Name:           cluster.Metadata.Name,
		EventType:      event.ResourceUpdate,
		ID:             cluster.Metadata.Id,
	}

	for _, h := range cs.clusterHandlers {
		h.OnChange(ev)
	}
	/*for _, h := range s.placementHandlers {
		h.OnChange(ev)
	}*/

	return cluster, nil
}

func (cs *clusterService) Delete(ctx context.Context, cluster *infrav3.Cluster) error {

	cluster, err := cs.Get(ctx, func(qo *commonv3.QueryOptions) {
		qo.Name = cluster.Metadata.Name
		qo.Project = cluster.Metadata.Project
		qo.Extended = true
	})
	if err != nil {
		return err
	}
	clusterId := cluster.Metadata.Id
	projectId := cluster.Metadata.Project

	err = bootstrapagent.DeleteForCluster(ctx, cs.sentryPool, *cluster, query.WithMeta(cluster.Metadata))
	if err != nil {
		return err
	}

	_log.Infow("deleting cluster", "name", cluster.Metadata.Name)

	_log.Debugw("setting cluster condition to pending delete", "name", cluster.Metadata.Name, "conditions", cluster.Spec.ClusterData.ClusterStatus.Conditions)
	clstrutil.SetClusterCondition(cluster, clstrutil.NewClusterDelete(constants.Pending, "deleted"))

	err = cs.UpdateClusterConditionStatus(ctx, cluster)
	if err != nil {
		return errors.Wrapf(err, "could not update cluster %s status to pending delete", cluster.Metadata.Name)
	}

	ev := event.Resource{
		PartnerID:      cluster.Metadata.Partner,
		OrganizationID: cluster.Metadata.Organization,
		ProjectID:      projectId,
		Name:           cluster.Metadata.Name,
		EventType:      event.ResourceDelete,
		ID:             clusterId,
	}

	for _, h := range cs.clusterHandlers {
		h.OnChange(ev)
	}

	return nil

}

func (cs *clusterService) deleteCluster(ctx context.Context, clusterId, projectId string) error {

	c := models.Cluster{
		ID:        uuid.MustParse(clusterId),
		ProjectId: uuid.MustParse(projectId),
	}
	err := cs.pcdao.DeleteProjectsForCluster(ctx, uuid.MustParse(clusterId))
	if err != nil {
		return errors.Wrapf(err, "could not delete projects for cluster %s", clusterId)
	}
	return cs.cdao.DeleteCluster(ctx, &c)
}

func (cs *clusterService) List(ctx context.Context, opts ...query.Option) (*infrav3.ClusterList, error) {

	queryOptions := commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(&queryOptions)
	}

	var proj cmodels.Project
	_, err := cs.adao.GetByName(ctx, queryOptions.Project, &proj)
	if err != nil {
		return nil, err
	}

	cdb, err := cs.cdao.ListClusters(ctx, commonv3.QueryOptions{
		Project:      proj.ID.String(),
		Organization: proj.OrganizationId.String(),
		Partner:      proj.PartnerId.String(),
	})
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
		cluster := cs.prepareClusterResponse(ctx, &infrav3.Cluster{}, &clstr, metro, projects, false)
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
			_log.Infow("Cluster not found for token", token, "error", err)
			return err
		}
		cluster = &infrav3.Cluster{}
		cluster = s.prepareClusterResponse(ctx, cluster, cdb, nil, nil, true)
		cluster.Spec.ClusterData.ClusterStatus.Conditions = []*infrav3.ClusterCondition{
			clstrutil.NewClusterRegister(commonv3.RafayConditionStatus_Success, "registered"),
		}

		err = s.UpdateClusterConditionStatus(ctx, cluster)
		if err != nil {
			_log.Infow("error updating cluster status", "error", err)
			return err
		}

		ev := event.Resource{
			PartnerID:      cdb.PartnerId.String(),
			OrganizationID: cdb.OrganizationId.String(),
			ProjectID:      cdb.ProjectId.String(),
			Name:           cdb.Name,
			EventType:      event.ResourceUpdateStatus,
			ID:             cdb.ID.String(),
		}

		for _, h := range s.clusterHandlers {
			h.OnChange(ev)
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

	s.notifyCluster(ctx, cluster)
	return clusterToken, nil
}

func (s *clusterService) UpdateClusterConditionStatus(ctx context.Context, current *infrav3.Cluster) error {

	existing, err := s.Get(ctx, func(qo *commonv3.QueryOptions) {
		qo.ClusterID = current.Metadata.Id
		qo.Name = current.Metadata.Name
		qo.Project = current.Metadata.Project
		qo.Extended = true
	})
	if err != nil {
		return err
	}

	_log.Debugw("existing cluster conditions", "name", existing.Spec.ClusterData.ClusterStatus.Conditions)
	_log.Debugw("current cluster conditions", "name", current.Spec.ClusterData.ClusterStatus.Conditions)

	err = patch.ClusterStatus(existing.Spec.ClusterData.ClusterStatus, current.Spec.ClusterData.ClusterStatus)
	if err != nil {
		_log.Debugw("failed to update cluster status, ", err.Error())
		return err
	}

	_log.Debugw("updated cluster conditions", "name", existing.Spec.ClusterData.ClusterStatus.Conditions)
	_log.Debugw("updated cluster object", "name", existing.Metadata.Name)

	//update the cluster
	_, err = s.Update(ctx, existing)
	if err != nil {
		return err
	}

	_log.Debugw("updated cluster in db", "name", existing.Metadata.Name)

	if clstrutil.IsClusterDeleted(existing) {
		err = s.deleteCluster(ctx, existing.Metadata.Id, existing.Metadata.Project)
		if err != nil {
			return err
		}
		_log.Debugw("deleted cluster in db", "name", existing.Metadata.Name)
	}

	// set current to patched existing
	*current = *existing

	return err
}

func (s *clusterService) UpdateClusterAnnotations(ctx context.Context, cluster *infrav3.Cluster) error {
	if len(cluster.Metadata.Annotations) > 0 {
		annBytes, _ := json.Marshal(cluster.Metadata.Annotations)
		return s.cdao.UpdateClusterAnnotations(ctx, &models.Cluster{
			ID:          uuid.MustParse(cluster.Metadata.Id),
			Annotations: json.RawMessage(annBytes),
		})
	}
	return nil
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

func (s *clusterService) GetClusterProjects(ctx context.Context, cluster *infrav3.Cluster) ([]models.ProjectCluster, error) {

	id, err := uuid.Parse(cluster.Metadata.Id)
	if err != nil {
		id = uuid.Nil
	}
	c, err := s.cdao.GetCluster(ctx, &models.Cluster{ID: id, Name: cluster.Metadata.Name})
	if err != nil {
		return nil, err
	}
	projects, err := s.pcdao.GetProjectsForCluster(ctx, c.ID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *clusterService) UpdateStatus(ctx context.Context, current *infrav3.Cluster, opts ...query.Option) error {
	queryOptions := commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(&queryOptions)
	}

	isAllowed, err := s.pcdao.ValidateClusterAccess(ctx, queryOptions)
	if err != nil {
		return err
	}

	if !isAllowed {
		return fmt.Errorf("forbidden: access to cluster %s from project in scope is not allowed", current.Metadata.Name)
	}

	err = s.UpdateClusterConditionStatus(ctx, current)
	if err != nil {
		return err
	}

	ev := event.Resource{
		PartnerID:      current.Metadata.Partner,
		OrganizationID: current.Metadata.Organization,
		ProjectID:      current.Metadata.Project,
		Name:           current.Metadata.Name,
		EventType:      event.ResourceUpdateStatus,
		ID:             current.Metadata.Id,
	}

	for _, h := range s.clusterHandlers {
		h.OnChange(ev)
	}

	s.notifyCluster(ctx, current)

	return nil
}

func (s *clusterService) notifyCluster(ctx context.Context, c *infrav3.Cluster) {
	b, err := json.Marshal(c.Metadata)
	if err != nil {
		_log.Infow("unable to marshal cluster meta", "error", err)
		return
	}

	err = s.cdao.Notify(clusterNotifyChan, string(b))
	if err != nil {
		_log.Infow("unable to send cluster notification", "error", err)
		return
	}
}

func (s *clusterService) AddEventHandler(evh event.Handler) {
	s.clusterHandlers = append(s.clusterHandlers, evh)
}

func (s *clusterService) Close() error {
	return s.dao.Close()
}
