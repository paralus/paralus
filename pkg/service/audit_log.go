package service

import (
	v1 "github.com/paralus/paralus/proto/rpc/audit"
	"github.com/uptrace/bun"
)

type AuditLogService interface {
	GetAuditLog(req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error)
	GetAuditLogByProjects(req *v1.GetAuditLogSearchRequest) (res *v1.GetAuditLogSearchResponse, err error)
}

func NewAuditLogElasticSearchService(url string, auditPattern string, logPrefix string) (AuditLogService, error) {
	auditQuery, err := NewElasticSearchQuery(url, auditPattern, logPrefix)
	if err != nil {
		return nil, err
	}
	return &auditLogElasticSearchService{auditQuery: auditQuery}, nil
}

func NewAuditLogDatabaseService(db *bun.DB, tag string) (AuditLogService, error) {
	return &auditLogDatabaseService{db: db, tag: tag}, nil
}

type RelayAuditService interface {
	GetRelayAudit(req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error)
	GetRelayAuditByProjects(req *v1.RelayAuditRequest) (res *v1.RelayAuditResponse, err error)
}

func NewRelayAuditDatabaseService(db *bun.DB, tag string) (RelayAuditService, error) {
	return &relayAuditDatabaseService{db: db, tag: tag}, nil
}

func NewRelayAuditElasticSearchService(url string, auditPattern string, logPrefix string) (RelayAuditService, error) {
	relayQuery, err := NewElasticSearchQuery(url, auditPattern, logPrefix)
	if err != nil {
		return nil, err
	}
	return &relayAuditElasticSearchService{relayQuery: relayQuery}, nil
}
