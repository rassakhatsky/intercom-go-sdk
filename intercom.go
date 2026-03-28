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
	"strings"

	"github.com/rassakhatsky/intercom-go-sdk/admins"
	"github.com/rassakhatsky/intercom-go-sdk/export"
	"github.com/rassakhatsky/intercom-go-sdk/messaging"
	"github.com/rassakhatsky/intercom-go-sdk/news"
	"github.com/rassakhatsky/intercom-go-sdk/segments"
	"github.com/rassakhatsky/intercom-go-sdk/settings"
	"github.com/rassakhatsky/intercom-go-sdk/tags"
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

	admins              *admins.Service
	AIContent           *AIContentService
	Articles            *ArticlesService
	awayStatusReasons   *admins.AwayStatusReasonsService
	brands              *settings.BrandsService
	Calls               *CallsService
	Contacts            *ContactsService
	Companies           *CompaniesService
	Conversations       *ConversationsService
	customChannelEvents *settings.ChannelEventsService
	CustomObjects       *CustomObjectsService
	DataAttributes      *DataAttributesService
	DataEvents          *DataEventsService
	dataExport          *export.DataService
	emails              *messaging.EmailsService
	exportReporting     *export.ReportingService
	FinVoice            *FinVoiceService
	HelpCenter          *HelpCenterService
	InternalArticles    *InternalArticlesService
	ipAllowlist         *settings.IPAllowlistService
	jobs                *settings.JobsService
	messages            *messaging.MessagesService
	news                *news.Service
	notes               *settings.NotesService
	PhoneCallRedirects  *PhoneCallRedirectsService
	segments            *segments.Service
	subscriptionTypes   *messaging.SubscriptionsService
	tags                *tags.Service
	teams               *admins.TeamsService
	Tickets             *TicketsService
	TicketStates        *TicketStatesService
	TicketTypes         *TicketTypesService
	Visitors            *VisitorsService
	Workflows           *WorkflowsService
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
	c.admins = admins.NewService(c)
	c.AIContent = (*AIContentService)(&c.common)
	c.Articles = (*ArticlesService)(&c.common)
	c.awayStatusReasons = admins.NewAwayStatusReasonsService(c)
	c.brands = settings.NewBrandsService(c)
	c.Calls = (*CallsService)(&c.common)
	c.Contacts = (*ContactsService)(&c.common)
	c.Companies = (*CompaniesService)(&c.common)
	c.Conversations = (*ConversationsService)(&c.common)
	c.customChannelEvents = settings.NewChannelEventsService(c)
	c.CustomObjects = (*CustomObjectsService)(&c.common)
	c.DataAttributes = (*DataAttributesService)(&c.common)
	c.DataEvents = (*DataEventsService)(&c.common)
	c.dataExport = export.NewDataService(c)
	c.emails = messaging.NewEmailsService(c)
	c.exportReporting = export.NewReportingService(c)
	c.FinVoice = (*FinVoiceService)(&c.common)
	c.HelpCenter = (*HelpCenterService)(&c.common)
	c.InternalArticles = (*InternalArticlesService)(&c.common)
	c.ipAllowlist = settings.NewIPAllowlistService(c)
	c.jobs = settings.NewJobsService(c)
	c.messages = messaging.NewMessagesService(c)
	c.news = news.NewService(c)
	c.notes = settings.NewNotesService(c)
	c.PhoneCallRedirects = (*PhoneCallRedirectsService)(&c.common)
	c.segments = segments.NewService(c)
	c.subscriptionTypes = messaging.NewSubscriptionsService(c)
	c.tags = tags.NewService(c)
	c.teams = admins.NewTeamsService(c)
	c.Tickets = (*TicketsService)(&c.common)
	c.TicketStates = (*TicketStatesService)(&c.common)
	c.TicketTypes = (*TicketTypesService)(&c.common)
	c.Visitors = (*VisitorsService)(&c.common)
	c.Workflows = (*WorkflowsService)(&c.common)
}

// Segments returns the segments service.
func (c *Client) Segments() *segments.Service {
	return c.segments
}

// Tags returns the tags service.
func (c *Client) Tags() *tags.Service {
	return c.tags
}

// ExportReporting returns the export reporting service.
func (c *Client) ExportReporting() *export.ReportingService {
	return c.exportReporting
}

// DataExport returns the data export service.
func (c *Client) DataExport() *export.DataService {
	return c.dataExport
}

// News returns the news service.
func (c *Client) News() *news.Service {
	return c.news
}

// Messages returns the messages service.
func (c *Client) Messages() *messaging.MessagesService {
	return c.messages
}

// Emails returns the emails service.
func (c *Client) Emails() *messaging.EmailsService {
	return c.emails
}

// SubscriptionTypes returns the subscription types service.
func (c *Client) SubscriptionTypes() *messaging.SubscriptionsService {
	return c.subscriptionTypes
}

// Brands returns the brands service.
func (c *Client) Brands() *settings.BrandsService {
	return c.brands
}

// IPAllowlist returns the IP allowlist service.
func (c *Client) IPAllowlist() *settings.IPAllowlistService {
	return c.ipAllowlist
}

// CustomChannelEvents returns the custom channel events service.
func (c *Client) CustomChannelEvents() *settings.ChannelEventsService {
	return c.customChannelEvents
}

// Jobs returns the jobs service.
func (c *Client) Jobs() *settings.JobsService {
	return c.jobs
}

// Notes returns the notes service.
func (c *Client) Notes() *settings.NotesService {
	return c.notes
}

// Admins returns the admins service.
func (c *Client) Admins() *admins.Service {
	return c.admins
}

// Teams returns the teams service.
func (c *Client) Teams() *admins.TeamsService {
	return c.teams
}

// AwayStatusReasons returns the away status reasons service.
func (c *Client) AwayStatusReasons() *admins.AwayStatusReasonsService {
	return c.awayStatusReasons
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

// DoDownload executes an HTTP request and streams the response body to w.
// It is intended for binary downloads (e.g., CSV/gzip exports). On 4xx/5xx
// responses, it reads the error body and returns an *ErrorResponse.
func (c *Client) DoDownload(ctx context.Context, req *http.Request, w io.Writer) error {
	req = req.WithContext(ctx)
	c.logger.Debug("http request", "method", req.Method, "url", req.URL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP %d: failed to read error body: %w", resp.StatusCode, readErr)
		}
		result := buildResult(resp, body)
		return resultError(result)
	}

	_, err = io.Copy(w, resp.Body)
	return err
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
