package agent

import (
	"context"
	"fmt"

	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateProjectsForCluster updates projects for bootstrap agent for cluster
func UpdateProjectsForCluster(ctx context.Context, sp sentryrpc.SentryPool, cluster infrav3.Cluster) error {

	sc, err := sp.NewClient(ctx)
	if err != nil {
		return err
	}
	defer sc.Close()

	//TODO: handle multiple bootstrap agents
	// by fetching other agent templates in partner scope

	resp, err := sc.GetBootstrapAgentTemplates(ctx, &commonv3.QueryOptions{
		GlobalScope: true,
		Selector:    "rafay.dev/defaultRelay=true",
	})
	if err != nil {
		return err
	}

	// create bootstrap agent
	for _, bat := range resp.Items {

		agent := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{
				Id:           cluster.Metadata.Id,
				Name:         cluster.Metadata.Name,
				Partner:      cluster.Metadata.Partner,
				Organization: cluster.Metadata.Organization,
				Project:      cluster.Metadata.Project,
				Labels: map[string]string{
					"rafay.dev/clusterName": cluster.Metadata.Name,
				},
			},
			Spec: &sentry.BootstrapAgentSpec{
				TemplateRef: fmt.Sprintf("template/%s", bat.Metadata.Name),
			},
		}

		for _, project := range cluster.Spec.ClusterData.Projects {
			agent.Metadata.Labels[fmt.Sprintf("project/%s", project.ProjectID)] = ""
		}

		_, err = sc.UpdateBootstrapAgent(ctx, agent)

		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.NotFound:
				default:
					return err
				}
			} else {
				return err
			}

		}

	}
	return nil
}
