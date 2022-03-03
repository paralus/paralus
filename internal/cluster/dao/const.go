package dao

const (
	deletingExpr          = "not (((conditions -> ?) @> ?::jsonb)) as deleting"
	conditionStatusQ      = "(conditions -> ?) @> ?::jsonb"
	notConditionStatusQ   = "not ((conditions -> ?) @> ?::jsonb)"
	conditionLastUpdatedQ = "(conditions #>> '{?, lastUpdated}')::timestamp with time zone < ?::timestamp with time zone"
)
