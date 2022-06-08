package query

import (
	"errors"
	"fmt"
	"time"

	"github.com/paralus/paralus/internal/random"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/uptrace/bun"
)

const (
	// DefaultLimit is the default limit if no limit is set in query options
	DefaultLimit = 10
	// MaxLimit is the max limit for page size
	MaxLimit = 50
	// MaxOffset is the max offset
	MaxOffset = 100000
	// DefaultOrderBy is the default column used to order results
	DefaultOrderBy = "created_at"
	orderASC       = "ASC"
	orderDESC      = "DESC"
	// DefaultOrder is the order of the results
	DefaultOrder = orderASC
)

const (
	nameWithAliasQ           = "?TableAlias.name = ?"
	idWithAliasQ             = "?TableAlias.id = ?"
	partnerIDWithAliasQ      = "?TableAlias.partner_id = ?"
	organizationIDWithAliasQ = "?TableAlias.organization_id = ?"
	projectIDwithAliasQ      = "?TableAlias.project_id = ?"

	nameQ        = "name = ?"
	deletedAtQ   = "deleted_at = ?"
	modifiedAtQ  = "modified_at = ?"
	displayNameQ = "display_name = ?"
	labelsQ      = "labels = ?"
	annotationsQ = "annotations = ?"
	idQ          = "id = ?"
)

const (
	scopeOrganization = "organization"
	scopePartner      = "partner"
	scopeProject      = "project"
	scopeProjects     = "projects"
	scopeUser         = "user"
	scopeCluster      = "cluster"
	scopeSSOUser      = "ssouser"
)

var (
	// ErrNoName is returned when name is not set in query option
	// trying to build query for get/update/delete
	ErrNoName     = errors.New("name not set in options")
	ErrNoNameOrID = errors.New("neither name nor id is set in options")
)

// Option is the functional query option signature
type Option func(*commonv3.QueryOptions)

type setOption func(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error)

// WithMeta sets meta in query options
func WithMeta(o *commonv3.Metadata) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Name = o.GetName()
		opts.ID = o.GetId()
		opts.Partner = o.GetPartner()
		opts.Organization = o.GetOrganization()
		opts.Project = o.GetProject()
		opts.DisplayName = o.GetName()
		opts.Labels = o.GetLabels()
		opts.Annotations = o.GetAnnotations()
	}
}

// WithOptions copies options to query options
func WithOptions(in *commonv3.QueryOptions) Option {
	return func(opts *commonv3.QueryOptions) {
		*opts = *in
		opts.Limit = getLimit(opts)
		opts.Offset = getOffset(opts)
	}
}

// WithIgnoreScopeDefault ignores default values for scope when building queries
func WithIgnoreScopeDefault() Option {
	return func(opts *commonv3.QueryOptions) {
		opts.IgnoreScopeDefault = true
	}
}

// WithExtended sets extended in query options
func WithExtended() Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Extended = true
	}
}

// WithGlobalScope should be used to query resources in global scope
// partnerID, orgID, projectID = 0, 0, 0
func WithGlobalScope() Option {
	return func(opts *commonv3.QueryOptions) {
		opts.GlobalScope = true
	}
}

func getLimit(opts *commonv3.QueryOptions) int64 {
	// if limit is < 0 send all records
	if opts.Limit == 0 {
		return DefaultLimit
	}

	if opts.Limit > MaxLimit {
		return MaxLimit
	}
	return opts.Limit
}

func getOffset(opts *commonv3.QueryOptions) int64 {
	if opts.Offset > MaxOffset {
		return MaxOffset
	}
	return opts.Offset
}

func Paginate(q *bun.SelectQuery, opts *commonv3.QueryOptions) *bun.SelectQuery {
	limit := getLimit(opts)

	if limit > 0 {
		offset := getOffset(opts)
		q = q.Limit(int(limit)).Offset(int(offset))
	}

	return q
}

// Select builds query for selecting resources
func Select(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	var err error
	q, err = setRequestMeta(q, opts)
	if err != nil {
		return nil, err
	}

	q, err = FilterLabels(q, opts)
	if err != nil {
		return nil, err
	}

	//q = Order(q, opts)

	return q, nil
}

// GetAccountID returns account ID from QueryOptions
func GetAccountID(opts *commonv3.QueryOptions) (string, error) {
	switch {
	case opts.GlobalScope:
		return "", nil
	default:
		return opts.Account, nil

	}
}

// GetOrganizationID returns organization id from QueryOptions
func GetOrganizationID(opts *commonv3.QueryOptions) (string, error) {
	switch {
	case opts.GlobalScope:
		return "", nil
	default:
		return opts.Organization, nil
	}
}

func setRequestMeta(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	var err error
	for _, f := range []setOption{
		setPartner,
		setOrganization,
		setProject,
	} {
		q, err = f(q, opts)
		if err != nil {
			return nil, err
		}
	}

	return q, nil

}

func setProject(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	id := opts.Project

	if !opts.GlobalScope && id != "" {
		q = q.Where(projectIDwithAliasQ, id)
	}

	return q, nil
}

func setOrganization(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {

	id := opts.Organization

	if !opts.GlobalScope && id != "" {
		q = q.Where(organizationIDWithAliasQ, id)
	}

	return q, nil

}

func setPartner(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	id := opts.Partner

	// global scope takes precedence over ignore scope default
	if !opts.GlobalScope && id != "" {
		q = q.Where(partnerIDWithAliasQ, id)
	}

	return q, nil
}

// WithName sets name in query options
func WithName(name string) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Name = name
	}
}

// WithSelector sets selector in query options
func WithSelector(selector string) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Selector = selector
	}
}

// WithDeleted sets deleted in query options
func WithDeleted() Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Deleted = true
	}
}

// WithPartnerID sets partner id in query options
func WithPartnerID(partnerID string) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Partner = partnerID
	}
}

// WithOrganizationID sets organization id in query options
func WithOrganizationID(organizationID string) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Organization = organizationID
	}
}

// WithProjectID sets project id in query options
func WithProjectID(projectID string) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Project = projectID
	}
}

// GetClusterID returns cluster ID from QueryOptions
func GetClusterID(opts *commonv3.QueryOptions) (string, error) {
	switch {
	case opts.GlobalScope:
		return "", nil
	default:
		return opts.ClusterID, nil

	}
}

// Get builds query for getting resource
func Get(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {

	var err error
	if !opts.GlobalScope {
		q, err = setRequestMeta(q, opts)
	}

	if err != nil {
		return nil, err
	}

	if opts.Name == "" && opts.ID == "" {
		return nil, ErrNoNameOrID
	}

	if opts.Name != "" {
		q = q.Where(nameWithAliasQ, opts.Name)
	} else if opts.ID != "" {
		q = q.Where(idWithAliasQ, opts.ID)
	}

	return q, nil
}

// Update builds query for updating resource
func Update(uq *bun.UpdateQuery, opts *commonv3.QueryOptions) (*bun.UpdateQuery, error) {

	uq = uq.Set(modifiedAtQ, time.Now())
	if opts.DisplayName != "" {
		uq = uq.Set(displayNameQ, opts.DisplayName)
	}
	if opts.Labels != nil {
		uq = uq.Set(labelsQ, opts.Labels)
	}
	if opts.Annotations != nil {
		uq = uq.Set(annotationsQ, opts.Annotations)
	}

	return uq, nil
}

// Delete builds query for deleting resource
func Delete(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.UpdateQuery, error) {
	var err error
	q, err = Get(q, opts)
	if err != nil {
		return nil, err
	}

	uq := q.DB().NewUpdate()
	if opts.Deleted {
		uq = uq.Set(nameQ, fmt.Sprintf("%s-%s", opts.Name, random.NewRandomString(10)))
	}
	now := time.Now()

	uq = uq.Set(deletedAtQ, &now)

	return uq, nil
}
