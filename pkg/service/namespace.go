package service

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/RafayLabs/rcloud-base/internal/cluster/dao"
	"github.com/RafayLabs/rcloud-base/internal/models"
	"github.com/RafayLabs/rcloud-base/pkg/converter"
	"github.com/RafayLabs/rcloud-base/pkg/patch"
	commonv3 "github.com/RafayLabs/rcloud-base/proto/types/commonpb/v3"
	"github.com/RafayLabs/rcloud-base/proto/types/controller"
	infrav3 "github.com/RafayLabs/rcloud-base/proto/types/infrapb/v3"
	"github.com/RafayLabs/rcloud-base/proto/types/scheduler"
	"github.com/google/uuid"
)

func (s *clusterService) GetNamespacesForConditions(ctx context.Context, conditions []scheduler.ClusterNamespaceCondition, clusterID string) (*scheduler.ClusterNamespaceList, error) {

	cns, count, err := dao.GetNamespacesForConditions(ctx, s.db, uuid.MustParse(clusterID), conditions)
	if err != nil {
		return nil, err
	}

	cnl := scheduler.ClusterNamespaceList{}
	cnl.Metadata.Count = int64(count)

	var items []*scheduler.ClusterNamespace
	for _, cn := range cns {
		ns := &scheduler.NamespaceTemplate{}
		if err = json.Unmarshal(cn.Namespace, ns); err != nil {
			return nil, nil
		}
		cnd := make([]*scheduler.ClusterNamespaceCondition, 0, 10)
		if err = json.Unmarshal(cn.Conditions, &cnd); err != nil {
			return nil, nil
		}
		st := &controller.NamespaceStatus{}
		if err = json.Unmarshal(cn.Status, st); err != nil {
			return nil, nil
		}
		nstype, err := strconv.Atoi(cn.Type)
		if err != nil {
			return nil, nil
		}
		items = append(items, &scheduler.ClusterNamespace{
			Metadata: &commonv3.Metadata{
				Name: cn.Name,
			},
			Spec: &scheduler.ClusterNamespaceSpec{
				Type:      scheduler.ClusterNamespaceType(nstype),
				ClusterID: cn.ClusterId.String(),
				Namespace: ns,
			},
			Status: &scheduler.ClusterNamespaceStatus{
				Conditions: cnd,
				Status:     st,
			},
		})
	}
	cnl.Items = items

	return &cnl, nil
}

func (s *clusterService) GetNamespaces(ctx context.Context, clusterID string) (*scheduler.ClusterNamespaceList, error) {

	cns, err := dao.GetNamespaces(ctx, s.db, uuid.MustParse(clusterID))
	if err != nil {
		return nil, err
	}

	cnl := scheduler.ClusterNamespaceList{}

	var items []*scheduler.ClusterNamespace
	for _, cn := range cns {
		ns := &scheduler.NamespaceTemplate{}
		if err = json.Unmarshal(cn.Namespace, ns); err != nil {
			return nil, nil
		}
		cnd := make([]*scheduler.ClusterNamespaceCondition, 0, 10)
		if err = json.Unmarshal(cn.Conditions, &cnd); err != nil {
			return nil, nil
		}
		st := &controller.NamespaceStatus{}
		if err = json.Unmarshal(cn.Status, st); err != nil {
			return nil, nil
		}
		nstype, err := strconv.Atoi(cn.Type)
		if err != nil {
			return nil, nil
		}
		items = append(items, &scheduler.ClusterNamespace{
			Metadata: &commonv3.Metadata{
				Name: cn.Name,
			},
			Spec: &scheduler.ClusterNamespaceSpec{
				Type:      scheduler.ClusterNamespaceType(nstype),
				ClusterID: cn.ClusterId.String(),
				Namespace: ns,
			},
			Status: &scheduler.ClusterNamespaceStatus{
				Conditions: cnd,
				Status:     st,
			},
		})
	}
	cnl.Items = items
	cnl.Metadata.Count = int64(len(items))

	return &cnl, nil
}

func (s *clusterService) GetNamespace(ctx context.Context, namespace string, clusterID string) (*scheduler.ClusterNamespace, error) {

	cn, err := dao.GetNamespace(ctx, s.db, uuid.MustParse(clusterID), namespace)
	if err != nil {
		return nil, err
	}

	ns := &scheduler.NamespaceTemplate{}
	if err = json.Unmarshal(cn.Namespace, ns); err != nil {
		return nil, nil
	}
	cnd := make([]*scheduler.ClusterNamespaceCondition, 0, 10)
	if err = json.Unmarshal(cn.Conditions, &cnd); err != nil {
		return nil, nil
	}
	st := &controller.NamespaceStatus{}
	if err = json.Unmarshal(cn.Status, st); err != nil {
		return nil, nil
	}
	nstype, err := strconv.Atoi(cn.Type)
	if err != nil {
		return nil, nil
	}
	cns := &scheduler.ClusterNamespace{
		Metadata: &commonv3.Metadata{
			Name: cn.Name,
		},
		Spec: &scheduler.ClusterNamespaceSpec{
			Type:      scheduler.ClusterNamespaceType(nstype),
			ClusterID: cn.ClusterId.String(),
			Namespace: ns,
		},
		Status: &scheduler.ClusterNamespaceStatus{
			Conditions: cnd,
			Status:     st,
		},
	}

	return cns, nil
}

func (s *clusterService) UpdateNamespaceStatus(ctx context.Context, current *scheduler.ClusterNamespace) error {

	existing, err := s.GetNamespace(ctx, current.Metadata.Name, current.Spec.ClusterID)
	if err != nil {
		return err
	}

	err = patch.NamespaceStatus(existing.Status, current.Status)
	if err != nil {
		return err
	}

	cn := models.ClusterNamespace{
		ClusterId:  uuid.MustParse(existing.Spec.ClusterID),
		Name:       existing.Metadata.Name,
		Type:       existing.Spec.Type.String(),
		Namespace:  converter.ConvertToJsonRawMessage(existing.Spec.Namespace),
		Conditions: converter.ConvertToJsonRawMessage(existing.Status.Conditions),
		Status:     converter.ConvertToJsonRawMessage(existing.Status),
	}

	err = dao.UpdateNamespaceStatus(ctx, s.db, &cn)
	if err != nil {
		return err
	}

	//TODO: as part of gitops
	/*ev := event.Resource{
		EventType: event.ResourceUpdateStatus,
		ID:        namespace.ClusterID,
	}

	for _, h := range s.workloadHandlers {
		h.OnChange(ev)
	}*/

	return nil
}

func (s *clusterService) GetNamespaceHashes(ctx context.Context, clusterID string) ([]infrav3.NameHash, error) {
	nameHashes, err := dao.GetNamespaceHashes(ctx, s.db, uuid.MustParse(clusterID))
	return nameHashes, err
}
