package agent

import (
	"context"
	"fmt"

	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteForCluster delete bootstrap agent
func DeleteForCluster(ctx context.Context, sp sentryrpc.SentryPool, cluster infrav3.Cluster, opts ...query.Option) error {

	sc, err := sp.NewClient(ctx)
	if err != nil {
		return err
	}
	defer sc.Close()

	resp, err := sc.GetBootstrapAgentTemplates(ctx, &commonv3.QueryOptions{
		GlobalScope: true,
		Selector:    "rafay.dev/defaultRelay=true",
	})
	if err != nil {
		return err
	}

	for _, bat := range resp.Items {

		agent := &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{
				Id:           cluster.Metadata.Id,
				Name:         cluster.Metadata.Name,
				Partner:      cluster.Metadata.Partner,
				Organization: cluster.Metadata.Organization,
				Project:      cluster.Metadata.Project,
			},
			Spec: &sentry.BootstrapAgentSpec{
				TemplateRef: fmt.Sprintf("template/%s", bat.Metadata.Name),
			},
		}

		_, err := sc.DeleteBootstrapAgent(ctx, agent)
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.NotFound:
					continue
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
