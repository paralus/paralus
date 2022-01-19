package match

import (
	"github.com/RafaySystems/rcloud-base/components/common/pkg/query"
	commonv3 "github.com/RafaySystems/rcloud-base/components/common/proto/types/commonpb/v3"
	"k8s.io/apimachinery/pkg/labels"
)

// Matcher is the interface for matching objects which implement Metadata interface
type Matcher interface {
	Match(meta commonv3.Metadata) bool
}

// New returns new matcher for options
func New(opts ...query.Option) (Matcher, error) {
	options := &commonv3.QueryOptions{}
	for _, opt := range opts {
		opt(options)
	}

	selector, err := labels.Parse(options.Selector)
	if err != nil {
		return nil, err
	}

	return &matcher{
		partnerID:      options.PartnerID,
		organizationID: options.OrganizationID,
		projectID:      options.ProjectID,
		selector:       selector,
		name:           options.Name,
	}, nil

}

type matcher struct {
	partnerID      string
	organizationID string
	projectID      string
	selector       labels.Selector
	name           string
}

var _ Matcher = (*matcher)(nil)

func (m *matcher) Match(meta commonv3.Metadata) bool {

	if meta.GetPartner() != m.partnerID {
		return false
	}

	if meta.GetOrganization() != m.organizationID {
		return false
	}

	if meta.GetProject() != m.projectID {
		return false
	}

	if !m.selector.Matches(labels.Set(meta.GetLabels())) {
		return false
	}

	if !(m.name == "") && m.name != meta.GetName() {
		return false
	}

	return true
}
