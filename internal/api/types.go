package api

// TagRef represents a tag in sub-resource responses.
// It is shared across multiple services (contacts, companies, conversations, tickets).
type TagRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// AdminRef represents an admin in sub-resource responses.
// It is shared across multiple services (tags, conversations, etc.).
type AdminRef struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// ContactRef is a lightweight reference to a contact.
// It is shared across notes, contacts, and companies services.
type ContactRef struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// NoteAuthor represents the admin who authored a note.
// It is shared across notes, contacts, and companies services.
type NoteAuthor struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// Note represents an Intercom note.
// It is shared across notes, contacts, and companies services.
type Note struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	CreatedAt int64       `json:"created_at,omitempty"`
	Body      string      `json:"body,omitempty"`
	Contact   *ContactRef `json:"contact,omitempty"`
	Author    *NoteAuthor `json:"author,omitempty"`
}

// NoteListResult is the response for listing notes on a contact or company.
// It is shared across contacts and companies services.
type NoteListResult struct {
	Type       string      `json:"type"`
	Data       []Note      `json:"data"`
	TotalCount int         `json:"total_count"`
	Pages      CursorPages `json:"pages"`
}
