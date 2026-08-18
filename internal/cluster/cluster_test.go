package cluster

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/paralus/paralus/internal/cluster/constants"
	"github.com/paralus/paralus/pkg/common"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	infrav3 "github.com/paralus/paralus/proto/types/infrapb/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetClusterGeneration(t *testing.T) {
	tests := []struct {
		name        string
		clusterType string
		want        constants.ClusterGeneration
		wantErr     bool
	}{
		{"empty type defaults to V2", "", constants.Cluster_V2, false},
		{"imported maps to V2", "imported", constants.Cluster_V2, false},
		{"unknown type errors", "bogus", constants.Cluster_Verr, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetClusterGeneration(tt.clusterType)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHasValidCharacters(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"lowercase alnum with dashes", "my-cluster-1", true},
		{"single lowercase letter", "a", true},
		{"uppercase rejected", "MyCluster", false},
		{"leading digit rejected", "1cluster", false},
		{"leading dash rejected", "-cluster", false},
		{"empty string rejected", "", false},
		{"underscore rejected", "my_cluster", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, HasValidCharacters(tt.input))
		})
	}
}

func TestExtractV2ClusterLabels(t *testing.T) {
	t.Run("fills mandatory labels and metro location", func(t *testing.T) {
		res := ExtractV2ClusterLabels(nil, nil, "My Edge", "imported", "US East")
		assert.Equal(t, "my-edge", res[constants.ClusterLabelKey])
		assert.Equal(t, "imported", res[constants.ClusterTypeKey])
		assert.Equal(t, "us-east", res[constants.ClusterLocationKey])
	})

	t.Run("omits location when metro name is empty", func(t *testing.T) {
		res := ExtractV2ClusterLabels(nil, nil, "edge", "imported", "")
		_, ok := res[constants.ClusterLocationKey]
		assert.False(t, ok)
	})

	t.Run("maps gpu and gpu_vendor edge data keys", func(t *testing.T) {
		res := ExtractV2ClusterLabels(map[string]string{"gpu": "true", "gpu_vendor": "NVIDIA Corp"}, nil, "edge", "imported", "")
		assert.Equal(t, "true", res[constants.ClusterGPU])
		assert.Equal(t, "nvidia-corp", res[constants.ClusterGPUVendor])
	})

	t.Run("passes through unrecognized edge data keys as-is", func(t *testing.T) {
		res := ExtractV2ClusterLabels(map[string]string{"custom": "value"}, nil, "edge", "imported", "")
		assert.Equal(t, "value", res["custom"])
	})

	t.Run("ignores platform-managed paralus.dev labels to avoid overwriting", func(t *testing.T) {
		res := ExtractV2ClusterLabels(nil, map[string]string{
			constants.ClusterLabelKey: "should-be-ignored",
			constants.ClusterGPU:      "should-be-ignored",
		}, "edge", "imported", "")
		assert.Equal(t, "edge", res[constants.ClusterLabelKey])
		_, ok := res[constants.ClusterGPU]
		assert.False(t, ok)
	})

	t.Run("keeps other paralus.dev labels not in the ignore list", func(t *testing.T) {
		res := ExtractV2ClusterLabels(nil, map[string]string{
			constants.ClusterID: "cluster-id-123",
		}, "edge", "imported", "")
		assert.Equal(t, "cluster-id-123", res[constants.ClusterID])
	})

	t.Run("keeps non-paralus.dev platform labels", func(t *testing.T) {
		res := ExtractV2ClusterLabels(nil, map[string]string{"other/key": "val"}, "edge", "imported", "")
		assert.Equal(t, "val", res["other/key"])
	})
}

func TestGetClusterOperatorYaml(t *testing.T) {
	t.Run("renders successfully and base64-encodes BootstrapCA", func(t *testing.T) {
		data := &common.DownloadData{ControlAddr: "control.example.com", APIAddr: "api.example.com"}
		clusterObj := &infrav3.Cluster{
			Metadata: &commonv3.Metadata{Name: "cluster-1"},
			Spec: &infrav3.ClusterSpec{
				ProxyConfig: &infrav3.ProxyConfig{BootstrapCA: "raw-ca-cert"},
			},
		}

		out, err := GetClusterOperatorYaml(context.Background(), data, clusterObj)
		require.NoError(t, err)
		assert.NotEmpty(t, out)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("raw-ca-cert")), clusterObj.Spec.ProxyConfig.BootstrapCA)
	})

	t.Run("handles a nil ProxyConfig", func(t *testing.T) {
		data := &common.DownloadData{}
		clusterObj := &infrav3.Cluster{
			Metadata: &commonv3.Metadata{Name: "cluster-1"},
			Spec:     &infrav3.ClusterSpec{},
		}

		_, err := GetClusterOperatorYaml(context.Background(), data, clusterObj)
		require.NoError(t, err)
	})
}
