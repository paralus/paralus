package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/hasher"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/models"
	hash "github.com/RafaySystems/rcloud-base/components/common/pkg/hasher"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/persistence/provider/pg"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ClusterNodesService is the interface for cluster nodes operations
type ClusterNodesService interface {
	// create cluster nodes
	CreateOrUpdateNode(ctx context.Context, clusterId uuid.UUID, cn *infrav3.ClusterNode) error
	//get cluster nodes
	GetClusterNodes(ctx context.Context, clusterID uuid.UUID) ([]infrav3.ClusterNode, error)
}

// clusterNodesService implements ClusterNodesService
type clusterNodesService struct {
	dao pg.EntityDAO
}

// NewClusterNodesService return new cluster nodes service
func NewClusterNodesService(db *bun.DB) ClusterNodesService {
	return &clusterNodesService{
		dao: pg.NewEntityDAO(db),
	}
}

func (s *clusterNodesService) CreateOrUpdateNode(ctx context.Context, clusterId uuid.UUID, cn *infrav3.ClusterNode) error {

	nodeid, err := uuid.Parse(cn.Metadata.Id)
	if err != nil {
		nodeid = uuid.Nil
	}

	// add public lable if public ip is assigned
	if isPublic(cn.Status.Ips) {
		if cn.Metadata.Labels == nil {
			cn.Metadata.Labels = make(map[string]string)
		}
		cn.Metadata.Labels[constants.Public] = "true"
	} else {
		delete(cn.Metadata.Labels, constants.Public)
	}

	//add hash annotation
	if cn.Metadata.Annotations == nil {
		cn.Metadata.Annotations = make(map[string]string)
	}
	h, err := hasher.GetNodeHashFrom(cn.Metadata.Labels, cn.Spec.Taints, cn.Spec.Unschedulable)
	if err != nil {
		err = fmt.Errorf("unable to create hash %s", err.Error())
		return err
	}
	cn.Metadata.Annotations[hash.ObjectHash] = h

	node := &models.ClusterNodes{
		ID:             nodeid,
		ClusterId:      clusterId,
		Name:           cn.Metadata.Name,
		DisplayName:    cn.Metadata.Name,
		OrganizationId: uuid.MustParse(cn.Metadata.Organization),
		PartnerId:      uuid.MustParse(cn.Metadata.Partner),
		ProjectId:      uuid.MustParse(cn.Metadata.Project),
		ModifiedAt:     time.Now(),
		State:          strconv.Itoa(int(cn.Status.State)),
	}

	if lbls, err := json.Marshal(cn.Metadata.Labels); err == nil {
		node.Labels = json.RawMessage(lbls)
	}
	if ann, err := json.Marshal(cn.Metadata.Annotations); err == nil {
		node.Annotations = json.RawMessage(ann)
	}
	if tnts, err := json.Marshal(cn.Spec.Taints); err == nil {
		node.Taints = json.RawMessage(tnts)
	}
	if ips, err := json.Marshal(cn.Status.Ips); err == nil {
		node.Ips = json.RawMessage(ips)
	}
	if all, err := json.Marshal(cn.Status.Allocatable); err == nil {
		node.Allocatable = json.RawMessage(all)
	}
	if cndts, err := json.Marshal(cn.Status.Conditions); err == nil {
		node.Conditions = json.RawMessage(cndts)
	}
	if info, err := json.Marshal(cn.Status.NodeInfo); err == nil {
		node.NodeInfo = json.RawMessage(info)
	}

	if nodeid == uuid.Nil {
		s.dao.Create(ctx, cn)
	} else {
		s.dao.Update(ctx, nodeid, cn)
	}

	return err
}

func (s *clusterNodesService) GetClusterNodes(ctx context.Context, clusterID uuid.UUID) ([]infrav3.ClusterNode, error) {
	var cns []models.ClusterNodes

	entities, err := s.dao.GetX(ctx, "cluster_id", clusterID, cns)

	if entities == nil {
		return nil, nil
	}
	cns = entities.([]models.ClusterNodes)

	var clusterNodes []infrav3.ClusterNode
	for _, node := range cns {
		var lbls map[string]string
		if node.Labels != nil {
			json.Unmarshal(node.Labels, lbls)
		}
		var ann map[string]string
		if node.Annotations != nil {
			json.Unmarshal(node.Annotations, ann)
		}
		var tnts []*commonv3.Taint
		if node.Taints != nil {
			json.Unmarshal(node.Taints, tnts)
		}
		state, _ := strconv.Atoi(node.State)
		var cnds []*commonv3.NodeCondition
		if node.Conditions != nil {
			json.Unmarshal(node.Conditions, cnds)
		}
		var info *commonv3.NodeSystemInfo
		if node.NodeInfo != nil {
			json.Unmarshal(node.NodeInfo, info)
		}
		var capacity *infrav3.Resources
		if node.Capacity != nil {
			json.Unmarshal(node.Capacity, capacity)
		}
		var allocatable *infrav3.Resources
		if node.Allocatable != nil {
			json.Unmarshal(node.Allocatable, allocatable)
		}
		var allocated *infrav3.Resources
		if node.Allocated != nil {
			json.Unmarshal(node.Allocated, allocated)
		}
		var ips []*infrav3.ClusterNodeIP
		if node.Ips != nil {
			json.Unmarshal(node.Ips, ips)
		}
		clusterNodes = append(clusterNodes, infrav3.ClusterNode{
			Metadata: &commonv3.Metadata{
				Name:         node.Name,
				Description:  node.DisplayName,
				Labels:       lbls,
				Annotations:  ann,
				Id:           node.ID.String(),
				Project:      node.ProjectId.String(),
				Organization: node.OrganizationId.String(),
				Partner:      node.PartnerId.String(),
			},
			Spec: &infrav3.ClusterNodeSpec{
				Unschedulable: node.Unschedulable,
				Taints:        tnts,
			},
			Status: &infrav3.ClusterNodeStatus{
				State:       infrav3.ClusterNodeState(state),
				Conditions:  cnds,
				NodeInfo:    info,
				Capacity:    capacity,
				Allocatable: allocatable,
				Allocated:   allocated,
				Ips:         ips,
			},
		})
	}

	return clusterNodes, err
}

func isPublic(ips []*infrav3.ClusterNodeIP) bool {
	for _, nodeIP := range ips {
		if nodeIP.PublicIP != "" {
			return true
		}
	}
	return false
}
