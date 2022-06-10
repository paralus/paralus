package query

import (
	commonv3 "github.com/paralus/paralus/proto/types/commonpb/v3"
	"github.com/uptrace/bun"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	labelsContainsQ    = "?TableAlias.labels @> ?::jsonb"
	labelsNotContainsQ = "NOT (?TableAlias.labels @> ?::jsonb)"
	labelExistsQ       = "?TableAlias.labels ? ?::text"
	labelDoesnotExistQ = "NOT ?TableAlias.labels ? ?::text"
)

func getClause(operator selection.Operator, opts *commonv3.QueryOptions) string {
	switch operator {
	case selection.Equals, selection.In:
		return labelsContainsQ
	case selection.NotEquals, selection.NotIn:
		return labelsNotContainsQ
	case selection.Exists:
		return labelExistsQ
	case selection.DoesNotExist:
		return labelDoesnotExistQ
	default:
		return labelsContainsQ
	}
}

// FilterLabels adds query filter for labels based on the selector
func FilterLabels(q *bun.SelectQuery, opts *commonv3.QueryOptions) (*bun.SelectQuery, error) {
	sel, err := labels.Parse(opts.Selector)
	if err != nil {
		return nil, err
	}

	reqs, _ := sel.Requirements()
	for i := range reqs {
		clause := getClause(reqs[i].Operator(), opts)
		switch reqs[i].Operator() {
		case selection.Equals:
			q = q.Where(
				clause,
				map[string]string{
					reqs[i].Key(): reqs[i].Values().List()[0],
				},
			)
		case selection.NotEquals:
			q = q.Where(
				clause,
				map[string]string{
					reqs[i].Key(): reqs[i].Values().List()[0],
				},
			)
		case selection.In:
			key := reqs[i].Key()
			values := reqs[i].Values().List()
			q = q.WhereGroup("grp", func(orq *bun.SelectQuery) *bun.SelectQuery {
				for _, value := range values {
					orq = orq.WhereOr(
						clause,
						map[string]string{
							key: value,
						},
					)
				}
				return orq
			})
		case selection.NotIn:
			key := reqs[i].Key()
			values := reqs[i].Values().List()
			q = q.WhereGroup("grp", func(orq *bun.SelectQuery) *bun.SelectQuery {
				for _, value := range values {
					orq = orq.Where(
						clause,
						map[string]string{
							key: value,
						},
					)
				}
				return orq
			})
		case selection.Exists:
			key := reqs[i].Key()
			q = q.Where(
				clause,
				bun.Safe("?"),
				key,
			)
		case selection.DoesNotExist:
			key := reqs[i].Key()
			q = q.Where(
				clause,
				bun.Safe("?"),
				key,
			)
		default:
			continue
		}

	}

	return q, nil
}
