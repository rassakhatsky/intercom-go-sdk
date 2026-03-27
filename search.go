package intercom

// SearchRequest represents a request body for Intercom's search endpoints.
type SearchRequest struct {
	Query      *Filter           `json:"query"`
	Pagination *SearchPagination `json:"pagination,omitempty"`
}

// SearchPagination controls pagination in search requests.
type SearchPagination struct {
	PerPage       int    `json:"per_page,omitempty"`
	StartingAfter string `json:"starting_after,omitempty"`
}

// Filter represents either a single field filter or a compound (AND/OR) filter.
// Use SingleFilterOf, And, and Or to construct filters.
type Filter struct {
	// Single filter fields
	Field    string `json:"field,omitempty"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

// SingleFilterOf creates a filter that matches a single field.
// The value can be a string, int, or slice for IN/NIN operators.
func SingleFilterOf(field, operator string, value any) *Filter {
	return &Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// And creates a compound filter that requires all sub-filters to match.
func And(filters ...*Filter) *Filter {
	return &Filter{
		Operator: "AND",
		Value:    filters,
	}
}

// Or creates a compound filter that requires any sub-filter to match.
func Or(filters ...*Filter) *Filter {
	return &Filter{
		Operator: "OR",
		Value:    filters,
	}
}
