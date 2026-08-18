package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGetAuditLogger(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")

	logger := GetAuditLogger(&AuditOptions{
		LogPath:    logPath,
		MaxSizeMB:  1,
		MaxBackups: 1,
		MaxAgeDays: 1,
	})
	require.NotNil(t, logger)

	// The encoder config only sets TimeKey/EncodeTime -- every other key
	// (MessageKey, LevelKey, ...) is left as "", and zap omits a field
	// entirely when its key is empty. So the log line carries only the
	// timestamp plus whatever explicit zap.Field values are passed; the
	// message text itself ("test message" here) never appears in output.
	logger.Info("test message", zap.String("foo", "bar"))
	require.NoError(t, logger.Sync())

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"timestamp"`)
	assert.Contains(t, string(data), `"foo":"bar"`)
	assert.NotContains(t, string(data), "test message")
}
