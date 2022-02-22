package agent

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrapbv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
)

// GetRelayAgentsForCluster creates bootstrap agent for cluster
func GetRelaysConfigForCluster(ctx context.Context, sp sentryrpc.SentryPool, cluster *infrapbv3.Cluster) ([]Relay, error) {
	sc, err := sp.NewClient(ctx)
	if err != nil {
		_log.Infow("unable to create sentry client", "error", err)
		err = errors.Wrap(err, "unable to create sentry client")
		return nil, err
	}
	defer sc.Close()

	var relays []Relay

	resp, err := sc.GetBootstrapAgentTemplates(ctx, &commonv3.QueryOptions{
		GlobalScope: true,
		Selector:    "rafay.dev/defaultRelay=true",
	})
	if err != nil {
		err = errors.Wrap(err, "unable to get bootstrap agent template")
		return nil, err
	}

	for _, bat := range resp.Items {
		agent, err := sc.GetBootstrapAgent(ctx, &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{
				Id:           cluster.Metadata.Id,
				Name:         cluster.Metadata.Name,
				Partner:      cluster.Metadata.Partner,
				Organization: cluster.Metadata.Organization,
				Project:      cluster.Metadata.Project,
				/*RequestMeta: cluster.RequestMeta,*/
			},
			Spec: &sentry.BootstrapAgentSpec{
				TemplateRef: fmt.Sprintf("template/%s", bat.Metadata.Name),
			},
		})
		if err != nil {
			err = errors.Wrap(err, "unable to get bootstrap agent")
			return nil, err

		}

		endpoint := ""
		for _, host := range bat.Spec.Hosts {
			if host.Type == sentry.BootstrapTemplateHostType_HostTypeExternal {
				endpoint = host.Host
			}
		}

		if endpoint == "" {
			return nil, fmt.Errorf("no external endpoint for relay bootstrap template %s", bat.Metadata.Name)
		}

		relays = append(relays, Relay{
			agent.Spec.Token, getRelayBootstrapAddr(), endpoint, bat.Metadata.Name, bat.Spec.Token,
		})
	}

	return relays, nil
}
