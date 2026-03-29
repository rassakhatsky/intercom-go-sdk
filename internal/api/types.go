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

// SegmentRef represents a segment in sub-resource responses.
// It is shared across contacts and companies services.
type SegmentRef struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	Name       string `json:"name,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	UpdatedAt  int64  `json:"updated_at,omitempty"`
	PersonType string `json:"person_type,omitempty"`
}

// SegmentListResult is the response for listing segments for a contact or company.
// It is shared across contacts and companies services.
type SegmentListResult struct {
	Type string       `json:"type"`
	Data []SegmentRef `json:"data"`
}

// Tag represents an Intercom tag (full representation).
// It is shared across tags and contacts services.
type Tag struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	AppliedAt *int64    `json:"applied_at,omitempty"`
	AppliedBy *AdminRef `json:"applied_by,omitempty"`
}

// TagList represents a list of tags.
// It is shared across tags and contacts services.
type TagList struct {
	Type string `json:"type"`
	Data []Tag  `json:"data"`
}

// SubscriptionType represents a subscription type in Intercom.
// It is shared across messaging and contacts services.
type SubscriptionType struct {
	Type               string        `json:"type"`
	ID                 string        `json:"id"`
	State              string        `json:"state,omitempty"`
	ConsentType        string        `json:"consent_type,omitempty"`
	DefaultTranslation *Translation  `json:"default_translation,omitempty"`
	Translations       []Translation `json:"translations,omitempty"`
	ContentTypes       []string      `json:"content_types,omitempty"`
}

// Translation represents a localised version of a subscription type.
// It is shared across messaging and contacts services.
type Translation struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Locale      string `json:"locale"`
}

// ArticleContent represents translated content for a single locale.
// It is shared across articles and help center services.
type ArticleContent struct {
	Type        string `json:"type,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Body        string `json:"body,omitempty"`
	AuthorID    int    `json:"author_id,omitempty"`
	State       string `json:"state,omitempty"`
	CreatedAt   int64  `json:"created_at,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
	URL         string `json:"url,omitempty"`
}

// ArticleTranslatedContent holds translations for an article keyed by locale.
// It is shared across articles and help center services.
type ArticleTranslatedContent struct {
	Type string          `json:"type,omitempty"`
	AR   *ArticleContent `json:"ar,omitempty"`
	BG   *ArticleContent `json:"bg,omitempty"`
	BS   *ArticleContent `json:"bs,omitempty"`
	CA   *ArticleContent `json:"ca,omitempty"`
	CS   *ArticleContent `json:"cs,omitempty"`
	DA   *ArticleContent `json:"da,omitempty"`
	DE   *ArticleContent `json:"de,omitempty"`
	EL   *ArticleContent `json:"el,omitempty"`
	EN   *ArticleContent `json:"en,omitempty"`
	ES   *ArticleContent `json:"es,omitempty"`
	ET   *ArticleContent `json:"et,omitempty"`
	FI   *ArticleContent `json:"fi,omitempty"`
	FR   *ArticleContent `json:"fr,omitempty"`
	HE   *ArticleContent `json:"he,omitempty"`
	HR   *ArticleContent `json:"hr,omitempty"`
	HU   *ArticleContent `json:"hu,omitempty"`
	ID   *ArticleContent `json:"id,omitempty"`
	IT   *ArticleContent `json:"it,omitempty"`
	JA   *ArticleContent `json:"ja,omitempty"`
	KO   *ArticleContent `json:"ko,omitempty"`
	LT   *ArticleContent `json:"lt,omitempty"`
	LV   *ArticleContent `json:"lv,omitempty"`
	MN   *ArticleContent `json:"mn,omitempty"`
	NB   *ArticleContent `json:"nb,omitempty"`
	NL   *ArticleContent `json:"nl,omitempty"`
	PL   *ArticleContent `json:"pl,omitempty"`
	PT   *ArticleContent `json:"pt,omitempty"`
	PtBR *ArticleContent `json:"pt-BR,omitempty"`
	RO   *ArticleContent `json:"ro,omitempty"`
	RU   *ArticleContent `json:"ru,omitempty"`
	SL   *ArticleContent `json:"sl,omitempty"`
	SR   *ArticleContent `json:"sr,omitempty"`
	SV   *ArticleContent `json:"sv,omitempty"`
	TR   *ArticleContent `json:"tr,omitempty"`
	VI   *ArticleContent `json:"vi,omitempty"`
	ZhCN *ArticleContent `json:"zh-CN,omitempty"`
	ZhTW *ArticleContent `json:"zh-TW,omitempty"`
}
