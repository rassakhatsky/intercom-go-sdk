package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	defaultBaseURL = "https://api.intercom.io/"
	apiVersion     = "2.15"
	userAgent      = "intercom-go-sdk/0.1.0"
)

// Client manages communication with the Intercom API.
type Client struct {
	token      string
	baseURL    string
	httpClient *http.Client
	logger     Logger

	common service // reuse a single struct for all services

	AIContent           *AIContentService
	Admins              *AdminsService
	Articles            *ArticlesService
	AwayStatusReasons   *AwayStatusReasonsService
	Brands              *BrandsService
	Calls               *CallsService
	Contacts            *ContactsService
	Companies           *CompaniesService
	Conversations       *ConversationsService
	CustomChannelEvents *CustomChannelEventsService
	CustomObjects       *CustomObjectsService
	DataAttributes      *DataAttributesService
	DataEvents          *DataEventsService
	DataExport          *DataExportService
	Emails              *EmailsService
	ExportReporting     *ExportReportingService
	FinVoice            *FinVoiceService
	HelpCenter          *HelpCenterService
	InternalArticles    *InternalArticlesService
	IPAllowlist         *IPAllowlistService
	Jobs                *JobsService
	Messages            *MessagesService
	News                *NewsService
	Notes               *NotesService
	PhoneCallRedirects  *PhoneCallRedirectsService
	Segments            *SegmentsService
	SubscriptionTypes   *SubscriptionTypesService
	Tags                *TagsService
	Teams               *TeamsService
	Tickets             *TicketsService
	TicketStates        *TicketStatesService
	TicketTypes         *TicketTypesService
	Visitors            *VisitorsService
}

// service is the base type for all Intercom API services.
type service struct {
	client *Client
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client for API requests.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithLogger sets a custom logger.
func WithLogger(l Logger) ClientOption {
	return func(c *Client) {
		c.logger = l
	}
}

// WithBaseURL sets a custom base URL for API requests.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) {
		if !strings.HasSuffix(u, "/") {
			u += "/"
		}
		c.baseURL = u
	}
}

// NewClient creates a new Intercom API client.
func NewClient(token string, opts ...ClientOption) *Client {
	c := &Client{
		token:      token,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
		logger:     slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	c.common.client = c
	c.initialize()
	return c
}

// initialize wires all services to share the common client.
func (c *Client) initialize() {
	c.AIContent = (*AIContentService)(&c.common)
	c.Admins = (*AdminsService)(&c.common)
	c.Articles = (*ArticlesService)(&c.common)
	c.AwayStatusReasons = (*AwayStatusReasonsService)(&c.common)
	c.Brands = (*BrandsService)(&c.common)
	c.Calls = (*CallsService)(&c.common)
	c.Contacts = (*ContactsService)(&c.common)
	c.Companies = (*CompaniesService)(&c.common)
	c.Conversations = (*ConversationsService)(&c.common)
	c.CustomChannelEvents = (*CustomChannelEventsService)(&c.common)
	c.CustomObjects = (*CustomObjectsService)(&c.common)
	c.DataAttributes = (*DataAttributesService)(&c.common)
	c.DataEvents = (*DataEventsService)(&c.common)
	c.DataExport = (*DataExportService)(&c.common)
	c.Emails = (*EmailsService)(&c.common)
	c.ExportReporting = (*ExportReportingService)(&c.common)
	c.FinVoice = (*FinVoiceService)(&c.common)
	c.HelpCenter = (*HelpCenterService)(&c.common)
	c.InternalArticles = (*InternalArticlesService)(&c.common)
	c.IPAllowlist = (*IPAllowlistService)(&c.common)
	c.Jobs = (*JobsService)(&c.common)
	c.Messages = (*MessagesService)(&c.common)
	c.News = (*NewsService)(&c.common)
	c.Notes = (*NotesService)(&c.common)
	c.PhoneCallRedirects = (*PhoneCallRedirectsService)(&c.common)
	c.Segments = (*SegmentsService)(&c.common)
	c.SubscriptionTypes = (*SubscriptionTypesService)(&c.common)
	c.Tags = (*TagsService)(&c.common)
	c.Teams = (*TeamsService)(&c.common)
	c.Tickets = (*TicketsService)(&c.common)
	c.TicketStates = (*TicketStatesService)(&c.common)
	c.TicketTypes = (*TicketTypesService)(&c.common)
	c.Visitors = (*VisitorsService)(&c.common)
}

// NewRequest creates an API request. A relative URL path can be provided in
// urlStr, in which case it is resolved relative to the BaseURL of the Client.
// If body is non-nil, it is JSON-encoded and included as the request body.
func (c *Client) NewRequest(method, urlStr string, body any) (*http.Request, error) {
	u, err := url.Parse(c.baseURL + urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Intercom-Version", apiVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

// Response wraps a Result to provide additional API-specific data.
type Response struct {
	*Result
}

// DoRaw executes an HTTP request and returns a Result with raw HTTP metadata
// and body. API errors (4xx/5xx) populate Result.Error instead of returning a
// Go error; only transport/IO failures return a Go error.
func (c *Client) DoRaw(ctx context.Context, req *http.Request) (*Result, error) {
	req = req.WithContext(ctx)
	c.logger.Debug("http request", "method", req.Method, "url", req.URL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return buildResult(resp, body), nil
}

// Do sends an API request and returns the API response. The JSON response
// body is decoded into v if v is non-nil. API errors (4xx/5xx) are returned
// as *ErrorResponse errors, preserving compatibility with IsNotFound,
// IsRateLimited, and IsUnauthorized.
func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*Response, error) {
	result, err := c.DoRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	response := &Response{Result: result}

	if result.Error != nil {
		return response, resultError(result)
	}

	if v != nil && result.StatusCode != http.StatusNoContent && len(result.Body) > 0 {
		if err := json.Unmarshal(result.Body, v); err != nil {
			return response, err
		}
	}

	return response, nil
}

// addQueryOptions encodes struct fields tagged with `url:"name,omitempty"`
// as query parameters and appends them to the given path.
func addQueryOptions(path string, opts any) (string, error) {
	if opts == nil {
		return path, nil
	}

	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return path, nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return path, fmt.Errorf("opts must be a struct, got %s", v.Kind())
	}

	params := url.Values{}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("url")
		if tag == "" || tag == "-" {
			continue
		}

		name, opts := parseTag(tag)
		fv := v.Field(i)

		if opts == "omitempty" && isZero(fv) {
			continue
		}

		val := fv
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}
		params.Set(name, fmt.Sprintf("%v", val.Interface()))
	}

	if len(params) == 0 {
		return path, nil
	}

	u, err := url.Parse(path)
	if err != nil {
		return path, err
	}
	q := u.Query()
	for k, vs := range params {
		for _, v := range vs {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func parseTag(tag string) (string, string) {
	parts := strings.SplitN(tag, ",", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return v.IsNil()
	default:
		return false
	}
}
