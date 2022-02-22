package tail

import (
	"github.com/rs/xid"
)

// {
// 	"level": "info",
// 	"ts": 1591154299.638035,
// 	"msg": "audit",
// 	"xid": "brbhcuug10lf2qfjj1hg",
// 	"method": "GET",
// 	"url": "/apis/metrics.k8s.io/v1beta1?timeout=32s",
// 	"serverName": "x3mxvkr.user.relay.rafay.local",
// 	"user": "rafay-core-debug",
// 	"remoteAddr": "127.0.0.1:63513",
// 	"statusCode": 503,
// 	"written": 20,
// 	"duration": 0.002513407
// }

// LogMsg represents log message
type LogMsg struct {
	//Level      string  `json:"level"`
	//Message    string  `json:"msg"`
	Timestamp  string  `json:"ts"`
	XID        xid.ID  `json:"xid"`
	Method     string  `json:"method"`
	URL        string  `json:"url"`
	Query      string  `json:"query"`
	ServerName string  `json:"serverName"`
	User       string  `json:"user"`
	RemoteAddr string  `json:"remoteAddr"`
	StatusCode int     `json:"statusCode"`
	Written    int64   `json:"written"`
	Duration   float64 `json:"duration"`
}

// AuditMsg represents audit message
type AuditMsg struct {
	Timestamp      string  `json:"ts"`
	XID            xid.ID  `json:"id"`
	StatusCode     int     `json:"sc"`
	UserName       string  `json:"un"`
	OrganizationID string  `json:"o"`
	PartnerID      string  `json:"p"`
	RemoteAddr     string  `json:"ra"`
	Duration       float64 `json:"d"`
	ClusterName    string  `json:"cn"`
	APIVersion     string  `json:"av"`
	Kind           string  `json:"k"`
	Namespace      string  `json:"ns"`
	Name           string  `json:"n"`
	Method         string  `json:"m"`
	URL            string  `json:"url"`
	Query          string  `json:"q"`
	Written        int64   `json:"w"`
	SessionType    string  `json:"st"`
}

// Reset resets audit message
func (m *AuditMsg) Reset() {
	m.Timestamp = ""
	m.XID = xid.NilID()
	m.StatusCode = 0
	m.UserName = ""
	m.OrganizationID = ""
	m.PartnerID = ""
	m.RemoteAddr = ""
	m.Duration = 0
	m.ClusterName = ""
	m.APIVersion = ""
	m.Kind = ""
	m.Namespace = ""
	m.Name = ""
	m.Method = ""
	m.URL = ""
	m.Query = ""
	m.Written = 0
}
