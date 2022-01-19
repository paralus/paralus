package cluster

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/bootstrapper"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/cluster/constants"
	"github.com/RafaySystems/rcloud-base/components/cluster-scheduler/pkg/internal/fixtures"
	"github.com/RafaySystems/rcloud-base/components/common/pkg/log"
	infrav3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/infrapb/v3"
)

var _log = log.GetLogger()

// GetClusterGeneration() looks up the cluster type and attempt to covert it into
// the cluster generation.
// Return error in case the type string is not recognized
func GetClusterGeneration(clusterType string) (constants.ClusterGeneration, error) {
	if clusterType == "" {
		return constants.Cluster_V2, nil
	}

	clusterTypeMap := map[string]constants.ClusterGeneration{
		"imported": constants.Cluster_V2,
	}

	g, ok := clusterTypeMap[clusterType]

	if ok {
		return g, nil
	}

	return constants.Cluster_Verr, fmt.Errorf("cluster type not found")
}

func HasValidCharacters(name string) bool {
	return regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(name)
}

func ExtractV2ClusterLabels(edgeDataLabels, clusterV2Labels map[string]string, edgeName, clusterType, metroName string) map[string]string {
	// Pre-fill mandatory labels
	res := map[string]string{
		constants.ClusterLabelKey: SanitizeLabelValues(edgeName),
		constants.ClusterTypeKey:  SanitizeLabelValues(clusterType),
	}
	if metroName != "" {
		res[constants.ClusterLocationKey] = SanitizeLabelValues(metroName)
	}
	// Copy over the labels passed by the user
	for key, value := range edgeDataLabels {
		switch key {
		case "gpu":
			res[constants.ClusterGPU] = value
		case "gpu_vendor":
			res[constants.ClusterGPUVendor] = SanitizeLabelValues(value)
		default:
			res[key] = value
		}
	}
	// Copy over the labels inserted by platform
	for key, value := range clusterV2Labels {
		if strings.HasPrefix(key, constants.RafayDomainLabel+"/") {
			switch key {
			case constants.ClusterGPU, constants.ClusterGPUVendor, constants.ClusterLabelKey, constants.ClusterTypeKey, constants.ClusterLocationKey:
				// Ignore the labels that infra adds so that we don't overwrite newer values
				continue
			default:
				res[key] = value
			}
		} else {
			res[key] = value
		}
	}
	return res
}

func GetClusterOperatorYaml(ctx context.Context, data *bootstrapper.DownloadData, cluster *infrav3.Cluster) (string, error) {

	_log.Infow("printing cluster in GetClusterOperatorYaml", "cluster", cluster)

	bb := new(bytes.Buffer)

	if cluster.Spec.ProxyConfig != nil {
		if cluster.Spec.ProxyConfig.BootstrapCA != "" {
			cluster.Spec.ProxyConfig.BootstrapCA = base64.StdEncoding.EncodeToString([]byte(cluster.Spec.ProxyConfig.BootstrapCA))
		}
	}

	err := fixtures.DownloadTemplate.Execute(bb, struct {
		DownloadData *bootstrapper.DownloadData
		Cluster      *infrav3.Cluster
	}{
		data, cluster,
	})
	if err != nil {
		_log.Errorw("error while downloading template GetClusterOperatorYaml", "cluster", cluster)
		return "", err
	}
	operatorSpec := string(bb.String())

	return operatorSpec, nil
}
