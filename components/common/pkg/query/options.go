package query

import (
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
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
	partnerIDWithAliasQ      = "?TableAlias.partner_id = ?"
	organizationIDWithAliasQ = "?TableAlias.organization_id = ?"
	projectIDwithAliasQ      = "?TableAlias.project_id = ?"
)

// Option is the functional query option signature
type Option func(*commonv3.QueryOptions)

type setOption func(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error)

// WithMeta sets meta in query options
func WithMeta(o *commonv3.Metadata) Option {
	return func(opts *commonv3.QueryOptions) {
		opts.Name = o.GetName()
		opts.ID = o.GetId()
		opts.PartnerID = o.GetPartner()
		opts.OrganizationID = o.GetOrganization()
		opts.ProjectID = o.GetProject()
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

func setRequestMeta(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	var err error
	for _, f := range []setOption{
		setPartnerID,
		setOrganizationID,
		setProjectID,
	} {
		q, err = f(q, opts)
		if err != nil {
			return nil, err
		}
	}

	return q, nil

}

func setProjectID(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	id := opts.ProjectID

	if !opts.GlobalScope && opts.IgnoreScopeDefault && id != "" {
		return q, nil
	}

	q = q.Where(projectIDwithAliasQ, id)

	return q, nil
}

func setOrganizationID(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {

	id := opts.OrganizationID

	if !opts.GlobalScope && opts.IgnoreScopeDefault && id != "" {
		return q, nil
	}

	q = q.Where(organizationIDWithAliasQ, id)

	return q, nil

}

func setPartnerID(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	id := opts.PartnerID

	// global scope takes precedence over ignore scope default
	if !opts.GlobalScope && opts.IgnoreScopeDefault && id != "" {
		return q, nil
	}

	q = q.Where(partnerIDWithAliasQ, id)

	return q, nil
}
