package audit

import (
	"net/http"
	"strings"

	"github.com/RafaySystems/rcloud-base/components/relay/pkg/sessions"
	"github.com/felixge/httpsnoop"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// options holds audit options
type options struct {
	basePath   string
	maxSizeMB  int
	maxBackups int
	maxAgeDays int
}

// Option is the functional argument for audit options
type Option func(o *options)

// WithBasePath sets base path for audit logs
func WithBasePath(path string) Option {
	return func(o *options) {
		o.basePath = path
	}
}

// WithMaxSizeMB sets max size of audit file before it is rotated
func WithMaxSizeMB(size int) Option {
	return func(o *options) {
		o.maxSizeMB = size
	}
}

// WithMaxBackups sets maximum number of backed up rotated audit logs to maintain
func WithMaxBackups(backups int) Option {
	return func(o *options) {
		o.maxBackups = backups
	}
}

// WithMaxAgeDays sets age after which backed up rotated audit logs will be deleted
func WithMaxAgeDays(ageDays int) Option {
	return func(o *options) {
		o.maxAgeDays = ageDays
	}
}

type audit struct {
	logger *zap.Logger
}

func (a *audit) handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		fields := make([]zapcore.Field, 0, 10)
		fields = append(fields, zap.String("xid", xid.New().String()))
		fields = append(fields, zap.String("method", r.Method))
		fields = append(fields, zap.String("url", r.URL.EscapedPath()))
		fields = append(fields, zap.String("query", r.URL.RawQuery))
		fields = append(fields, zap.String("serverName", r.TLS.ServerName))
		fields = append(fields, zap.String("user", r.TLS.PeerCertificates[0].Subject.CommonName))
		fields = append(fields, zap.String("remoteAddr", strings.Split(r.RemoteAddr, ":")[0]))
		fields = append(fields, zap.Int("statusCode", metrics.Code))
		fields = append(fields, zap.Int64("written", metrics.Written))
		fields = append(fields, zap.Duration("duration", metrics.Duration))
		if r.Header.Get("X-Rafay-Audit") == "" {
			a.logger.Info("access", fields...)
		} else {
			a.logger.Info("audit", fields...)
		}

		if metrics.Code == http.StatusUnauthorized || metrics.Code == http.StatusBadGateway {
			sessKey := r.Header.Get("X-Rafay-Sessionkey")
			sessions.SetSessionErrorFlag(sessKey)
		}

	})
}

// WrapWithAudit wrap audit handler around next http.Handler
func WrapWithAudit(next http.Handler, opts ...Option) http.Handler {

	auditOpts := &options{}
	for _, opt := range opts {
		opt(auditOpts)
	}

	audit := &audit{logger: getAuditLogger(auditOpts)}
	return audit.handler(next)
}
