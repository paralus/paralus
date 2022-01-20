package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalConfig(t *testing.T) {
	assert.Equal(t, configFileDefaultPath, getLocalConfigPath(), "read from default connection config path if environment variable is not set")

	os.Setenv(configFilePathEnvironmentVariable, "dir/custom_path.json")
	assert.Equal(t, "dir/custom_path.json", getLocalConfigPath())
}
