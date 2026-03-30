package api

import (
	"encoding/json"
	"testing"
)

func TestOperatorConstants(t *testing.T) {
	tests := []struct {
		name string
		op   Operator
		want string
	}{
		{"Equals", OpEquals, "="},
		{"NotEquals", OpNotEquals, "!="},
		{"GreaterThan", OpGreaterThan, ">"},
		{"LessThan", OpLessThan, "<"},
		{"Contains", OpContains, "~"},
		{"NotContains", OpNotContains, "!~"},
		{"In", OpIn, "IN"},
		{"NotIn", OpNotIn, "NIN"},
		{"Starts", OpStarts, "^"},
		{"Ends", OpEnds, "$"},
		{"AND", OpAND, "AND"},
		{"OR", OpOR, "OR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.op) != tt.want {
				t.Errorf("Op%s = %q, want %q", tt.name, tt.op, tt.want)
			}
		})
	}
}

func TestOperator_IsDefinedType(t *testing.T) {
	// Verify Operator is a defined type (not an alias) by checking
	// that string(op) round-trips correctly.
	op := OpEquals
	s := string(op)
	if s != "=" {
		t.Errorf("string(OpEquals) = %q, want %q", s, "=")
	}
	roundTripped := Operator(s)
	if roundTripped != OpEquals {
		t.Errorf("Operator(%q) = %q, want %q", s, roundTripped, OpEquals)
	}
}

func TestOperatorConstants_InFilter(t *testing.T) {
	f := SingleFilterOf("role", OpIn, []string{"admin", "user"})

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["operator"] != "IN" {
		t.Errorf("operator = %v, want IN", got["operator"])
	}
	values, ok := got["value"].([]any)
	if !ok {
		t.Fatalf("value should be array, got %T", got["value"])
	}
	if len(values) != 2 {
		t.Errorf("value length = %d, want 2", len(values))
	}
}

func TestSingleFilterOf(t *testing.T) {
	f := SingleFilterOf("created_at", OpGreaterThan, "1306054154")

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["field"] != "created_at" {
		t.Errorf("field = %v, want created_at", got["field"])
	}
	if got["operator"] != ">" {
		t.Errorf("operator = %v, want >", got["operator"])
	}
	if got["value"] != "1306054154" {
		t.Errorf("value = %v, want 1306054154", got["value"])
	}

	// Should NOT have "operator" at the compound level
	if _, ok := got["value"].([]any); ok {
		t.Error("single filter value should not be an array")
	}
}

func TestSingleFilterOf_IntValue(t *testing.T) {
	f := SingleFilterOf("created_at", OpGreaterThan, 1306054154)

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// JSON numbers decode as float64
	if got["value"] != float64(1306054154) {
		t.Errorf("value = %v (%T), want 1306054154", got["value"], got["value"])
	}
}

func TestAnd(t *testing.T) {
	f := And(
		SingleFilterOf("created_at", OpGreaterThan, "1306054154"),
		SingleFilterOf("created_at", OpLessThan, "1609459200"),
	)

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["operator"] != "AND" {
		t.Errorf("operator = %v, want AND", got["operator"])
	}

	value, ok := got["value"].([]any)
	if !ok {
		t.Fatalf("value should be an array, got %T", got["value"])
	}
	if len(value) != 2 {
		t.Fatalf("value length = %d, want 2", len(value))
	}

	first := value[0].(map[string]any)
	if first["field"] != "created_at" {
		t.Errorf("first filter field = %v, want created_at", first["field"])
	}
}

func TestOr(t *testing.T) {
	f := Or(
		SingleFilterOf("email", OpEquals, "alice@example.com"),
		SingleFilterOf("email", OpEquals, "bob@example.com"),
	)

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["operator"] != "OR" {
		t.Errorf("operator = %v, want OR", got["operator"])
	}

	value, ok := got["value"].([]any)
	if !ok {
		t.Fatalf("value should be an array, got %T", got["value"])
	}
	if len(value) != 2 {
		t.Fatalf("value length = %d, want 2", len(value))
	}
}

func TestNestedAndOr(t *testing.T) {
	f := And(
		Or(
			SingleFilterOf("email", OpEquals, "alice@example.com"),
			SingleFilterOf("email", OpEquals, "bob@example.com"),
		),
		SingleFilterOf("created_at", OpGreaterThan, "1306054154"),
	)

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if got["operator"] != "AND" {
		t.Errorf("operator = %v, want AND", got["operator"])
	}

	value := got["value"].([]any)
	if len(value) != 2 {
		t.Fatalf("value length = %d, want 2", len(value))
	}

	// First element should be a nested OR compound filter
	nested := value[0].(map[string]any)
	if nested["operator"] != "OR" {
		t.Errorf("nested operator = %v, want OR", nested["operator"])
	}
	nestedValue := nested["value"].([]any)
	if len(nestedValue) != 2 {
		t.Errorf("nested value length = %d, want 2", len(nestedValue))
	}
}

func TestSearchRequest_MarshalJSON(t *testing.T) {
	sr := &SearchRequest{
		Query: And(
			SingleFilterOf("created_at", OpGreaterThan, "1306054154"),
		),
		Pagination: &SearchPagination{
			PerPage:       5,
			StartingAfter: "abc123",
		},
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Check query is present
	query, ok := got["query"].(map[string]any)
	if !ok {
		t.Fatalf("query should be a map, got %T", got["query"])
	}
	if query["operator"] != "AND" {
		t.Errorf("query.operator = %v, want AND", query["operator"])
	}

	// Check pagination is present
	pagination, ok := got["pagination"].(map[string]any)
	if !ok {
		t.Fatalf("pagination should be a map, got %T", got["pagination"])
	}
	if pagination["per_page"] != float64(5) {
		t.Errorf("pagination.per_page = %v, want 5", pagination["per_page"])
	}
	if pagination["starting_after"] != "abc123" {
		t.Errorf("pagination.starting_after = %v, want abc123", pagination["starting_after"])
	}
}

func TestSearchRequest_NoPagination(t *testing.T) {
	sr := &SearchRequest{
		Query: SingleFilterOf("email", OpEquals, "test@example.com"),
	}

	data, err := json.Marshal(sr)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := got["pagination"]; ok {
		t.Error("pagination should be omitted when nil")
	}

	query := got["query"].(map[string]any)
	if query["field"] != "email" {
		t.Errorf("query.field = %v, want email", query["field"])
	}
}
