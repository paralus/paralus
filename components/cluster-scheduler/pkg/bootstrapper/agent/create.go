package agent

import (
	"context"
	"encoding/json"
	"fmt"

	configrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/config"
	sentryrpc "github.com/RafaySystems/rcloud-base/components/common/proto/rpc/sentry"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	infrapbv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
	"github.com/RafaySystems/rcloud-base/components/common/proto/types/sentry"
	"github.com/pkg/errors"
	"github.com/rs/xid"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/RafaySystems/rcloud-base/components/common/pkg/log"
)

var (
	_log = log.GetLogger()
)

const (
	maxDials = 2
)

func getRelayBootstrapAddr() string {
	return viper.GetString("SENTRY_BOOTSTRAP_ADDR")
}

type Relay struct {
	Token         string `json:"token"`
	Addr          string `json:"addr"`
	Endpoint      string `json:"endpoint"`
	Name          string `json:"name"`
	TemplateToken string `json:"templateToken"`
}

// CreateForCluster creates bootstrap agent for cluster
func CreateForCluster(ctx context.Context, sp sentryrpc.SentryPool, cp configrpc.ConfigPool, cluster *infrapbv3.Cluster) error {
	sc, err := sp.NewClient(ctx)
	if err != nil {
		err = errors.Wrap(err, "unable to create sentry client")
		return err
	}
	defer sc.Close()

	var relays []Relay

	resp, err := sc.GetBootstrapAgentTemplates(ctx, &commonv3.QueryOptions{
		GlobalScope: true,
		Selector:    "rafay.dev/defaultRelay=true",
	})
	if err != nil {
		err = errors.Wrap(err, "unable to get bootstrap agent template")
		return err
	}

	// create bootstrap agent
	for _, bat := range resp.Items {
		found := true
		agent, err := sc.GetBootstrapAgent(ctx, &sentry.BootstrapAgent{
			Metadata: &commonv3.Metadata{
				Name: cluster.Metadata.Id,
				//RequestMeta: cluster.RequestMeta, TODO
			},
			Spec: &sentry.BootstrapAgentSpec{
				TemplateRef: fmt.Sprintf("template/%s", bat.Metadata.Name),
			},
		})
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.NotFound:
					found = false
				default:
					return err
				}
			} else {
				err = errors.Wrap(err, "unable to get bootstrap agent")
				return err
			}

		}

		if !found {
			agent = &sentry.BootstrapAgent{
				Metadata: &commonv3.Metadata{
					Name:        cluster.Metadata.Id,
					Description: cluster.Metadata.Name,
					Labels: map[string]string{
						"rafay.dev/clusterName": cluster.Metadata.Name,
					},
					Partner:      cluster.Metadata.Partner,
					Organization: cluster.Metadata.Organization,
					Project:      cluster.Metadata.Project,
				},
				Spec: &sentry.BootstrapAgentSpec{
					TemplateRef: bat.Metadata.Name,
					Token:       xid.New().String(),
				},
			}

			for _, project := range cluster.Spec.ClusterData.Projects {
				agent.Metadata.Labels[fmt.Sprintf("project/%s", project.ProjectID)] = ""
			}

			_, err := sc.CreateBootstrapAgent(ctx, agent)
			if err != nil {
				_log.Infow("unable to create bootstrap agent", "error", err, "agent", *agent)
				err = errors.Wrap(err, "unable to create bootstrap agent")
				return err
			}
		}
		endpoint := ""
		for _, host := range bat.Spec.Hosts {
			if host.Type == sentry.BootstrapTemplateHostType_HostTypeExternal {
				endpoint = host.Host
			}
		}

		if endpoint == "" {
			return fmt.Errorf("no external endpoint for bootstrap template %s", bat.Metadata.Name)
		}

		relays = append(relays, Relay{
			agent.Spec.Token, getRelayBootstrapAddr(), endpoint, bat.Metadata.Name, bat.Spec.Token,
		})
	}

	relaysBytes, _ := json.Marshal(relays)
	if cluster.Metadata.Annotations == nil {
		cluster.Metadata.Annotations = make(map[string]string)
	}
	cluster.Metadata.Annotations["relays"] = string(relaysBytes)

	//TODO: to revisit during gitops config component
	/* relayAgentConfig := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "relay-agent-config",
		},
		Data: map[string]string{
			"clusterID": cluster.Metadata.Id,
			"relays":    string(relaysBytes),
			"maxDials":  strconv.FormatInt(maxDials, 10),
			// "dialoutProxy":               "",
			// "dialoutProxyAuthentication": "",
		},
	}
	relayAgentConfigObj, err := runtime.FromObject(relayAgentConfig)
	if err != nil {
		return err
	}

	override := &config.Override{
		Metadata: &commonv3.Metadata{
			Name:         fmt.Sprintf("relay-override-%s", cluster.Metadata.Name),
			Partner:      cluster.Metadata.Partner,
			Organization: cluster.Metadata.Organization,
			Labels: map[string]string{
				config.OverrideScope:   config.OverrideScopeSpecificCluster,
				config.OverrideCluster: cluster.Metadata.Name,
			},
			Annotations: map[string]string{
				"rafay.dev/weight": "10",
			},
		},
		Spec: &config.OverrideSpec{
			ResourceSelector: "rafay.dev/system=true",
			Overrides: []*clusterv2.StepObject{
				relayAgentConfigObj,
			},
		},
	}

	// create override
	cc, err := cp.NewClient(ctx)
	if err != nil {
		err = errors.Wrap(err, "unable to create config client")
		return err
	}
	defer cc.Close()

	_, err = cc.UpdateOverride(ctx, override)
	if err != nil {
		return err
	} */

	return nil
}
