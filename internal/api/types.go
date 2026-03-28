package api

// TagRef represents a tag in sub-resource responses.
// It is shared across multiple services (contacts, companies, conversations, tickets).
type TagRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}
