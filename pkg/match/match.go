package match

import (
	"github.com/paralus/paralus/pkg/query"
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
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
		partner:      options.Partner,
		organization: options.Organization,
		project:      options.Project,
		selector:     selector,
		name:         options.Name,
	}, nil

}

type matcher struct {
	partner      string
	organization string
	project      string
	selector     labels.Selector
	name         string
}

var _ Matcher = (*matcher)(nil)

func (m *matcher) Match(meta commonv3.Metadata) bool {

	if meta.GetPartner() != m.partner {
		return false
	}

	if meta.GetOrganization() != m.organization {
		return false
	}

	if meta.GetProject() != m.project {
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
