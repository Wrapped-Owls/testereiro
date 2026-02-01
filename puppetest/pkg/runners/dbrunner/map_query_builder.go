package dbrunner

import (
	"fmt"
	"maps"
	"strings"

	"github.com/wrapped-owls/testereiro/puppetest/internal/stgctx"
)

// MapQueryBuilder builds a query from a map of filters.
type MapQueryBuilder struct {
	table       string
	fields      []string
	lateFilters []filterFromContext
}

// NewMapQuery creates a new MapQueryBuilder with the given table and initial filters.
func NewMapQuery(table string, filters map[string]any) *MapQueryBuilder {
	qb := &MapQueryBuilder{
		table: table,
	}
	if len(filters) > 0 {
		qb.AddFilter(func(_ stgctx.RunnerContext) (map[string]any, error) {
			return filters, nil
		})
	}
	return qb
}

func (b *MapQueryBuilder) AddFilter(filter filterFromContext) {
	b.lateFilters = append(b.lateFilters, filter)
}

func (b *MapQueryBuilder) AddSelectFields(fields ...string) {
	b.fields = append(b.fields, fields...)
}

func (b *MapQueryBuilder) Build(ctx stgctx.RunnerContext) (string, []any, error) {
	filters := make(map[string]any)
	for _, filterResolver := range b.lateFilters {
		newFilter, err := filterResolver(ctx)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build late filters: %w", err)
		}
		maps.Copy(filters, newFilter)
	}

	var (
		where = make([]string, 0, len(filters))
		args  = make([]any, 0, len(filters))
	)
	for filterKey, filterValue := range filters {
		where = append(where, fmt.Sprintf("%s = ?", filterKey))
		args = append(args, filterValue)
	}

	selectionFields := "*"
	if len(b.fields) > 0 {
		selectionFields = strings.Join(b.fields, ",")
	}
	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		selectionFields, b.table,
	)
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	return query, args, nil
}
