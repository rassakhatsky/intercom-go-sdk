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
	aiPkg "github.com/rassakhatsky/intercom-go-sdk/ai"
	"github.com/rassakhatsky/intercom-go-sdk/api"
	"github.com/rassakhatsky/intercom-go-sdk/articles"
	"github.com/rassakhatsky/intercom-go-sdk/calls"
	"github.com/rassakhatsky/intercom-go-sdk/companies"
	"github.com/rassakhatsky/intercom-go-sdk/contacts"
	"github.com/rassakhatsky/intercom-go-sdk/conversations"
	"github.com/rassakhatsky/intercom-go-sdk/data"
	"github.com/rassakhatsky/intercom-go-sdk/export"
	"github.com/rassakhatsky/intercom-go-sdk/helpcenter"
	"github.com/rassakhatsky/intercom-go-sdk/messaging"
	"github.com/rassakhatsky/intercom-go-sdk/news"
	"github.com/rassakhatsky/intercom-go-sdk/segments"
	"github.com/rassakhatsky/intercom-go-sdk/settings"
	"github.com/rassakhatsky/intercom-go-sdk/tags"
	"github.com/rassakhatsky/intercom-go-sdk/tickets"
	"github.com/rassakhatsky/intercom-go-sdk/workflows"
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
	logger     api.Logger

	admins              *admins.Service
	aiContent           *aiPkg.ContentService
	articles            *articles.Service
	awayStatusReasons   *admins.AwayStatusReasonsService
	brands              *settings.BrandsService
	calls               *calls.Service
	companies           *companies.Service
	contacts            *contacts.Service
	conversations       *conversations.Service
	customChannelEvents *settings.ChannelEventsService
	customObjects       *data.ObjectsService
	dataAttributes      *data.AttributesService
	dataEvents          *data.EventsService
	dataExport          *export.DataService
	emails              *messaging.EmailsService
	exportReporting     *export.ReportingService
	finVoice            *aiPkg.VoiceService
	helpCenter          *helpcenter.Service
	internalArticles    *articles.InternalService
	ipAllowlist         *settings.IPAllowlistService
	jobs                *settings.JobsService
	messages            *messaging.MessagesService
	news                *news.Service
	notes               *settings.NotesService
	phoneCallRedirects  *calls.RedirectsService
	segments            *segments.Service
	subscriptionTypes   *messaging.SubscriptionsService
	tags                *tags.Service
	teams               *admins.TeamsService
	tickets             *tickets.Service
	ticketStates        *tickets.StatesService
	ticketTypes         *tickets.TypesService
	visitors            *contacts.VisitorsService
	workflows           *workflows.Service
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
func WithLogger(l api.Logger) ClientOption {
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
// It panics if token is empty or contains only whitespace.
func NewClient(token string, opts ...ClientOption) *Client {
	if strings.TrimSpace(token) == "" {
		panic("intercom: NewClient requires a non-empty token")
	}
	c := &Client{
		token:      token,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
		logger:     slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	c.initialize()
	return c
}

// initialize wires all services to share the common client.
func (c *Client) initialize() {
	c.admins = admins.NewService(c)
	c.aiContent = aiPkg.NewContentService(c)
	c.articles = articles.NewService(c)
	c.awayStatusReasons = admins.NewAwayStatusReasonsService(c)
	c.brands = settings.NewBrandsService(c)
	c.calls = calls.NewService(c)
	c.contacts = contacts.NewService(c)
	c.companies = companies.NewService(c)
	c.conversations = conversations.NewService(c)
	c.customChannelEvents = settings.NewChannelEventsService(c)
	c.customObjects = data.NewObjectsService(c)
	c.dataAttributes = data.NewAttributesService(c)
	c.dataEvents = data.NewEventsService(c)
	c.dataExport = export.NewDataService(c)
	c.emails = messaging.NewEmailsService(c)
	c.exportReporting = export.NewReportingService(c)
	c.finVoice = aiPkg.NewVoiceService(c)
	c.helpCenter = helpcenter.NewService(c)
	c.internalArticles = articles.NewInternalService(c)
	c.ipAllowlist = settings.NewIPAllowlistService(c)
	c.jobs = settings.NewJobsService(c)
	c.messages = messaging.NewMessagesService(c)
	c.news = news.NewService(c)
	c.notes = settings.NewNotesService(c)
	c.phoneCallRedirects = calls.NewRedirectsService(c)
	c.segments = segments.NewService(c)
	c.subscriptionTypes = messaging.NewSubscriptionsService(c)
	c.tags = tags.NewService(c)
	c.teams = admins.NewTeamsService(c)
	c.tickets = tickets.NewService(c)
	c.ticketStates = tickets.NewStatesService(c)
	c.ticketTypes = tickets.NewTypesService(c)
	c.visitors = contacts.NewVisitorsService(c)
	c.workflows = workflows.NewService(c)
}

// AIContent returns the AI content service.
func (c *Client) AIContent() *aiPkg.ContentService {
	return c.aiContent
}

// FinVoice returns the Fin Voice service.
func (c *Client) FinVoice() *aiPkg.VoiceService {
	return c.finVoice
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

// DataEvents returns the data events service.
func (c *Client) DataEvents() *data.EventsService {
	return c.dataEvents
}

// DataAttributes returns the data attributes service.
func (c *Client) DataAttributes() *data.AttributesService {
	return c.dataAttributes
}

// CustomObjects returns the custom objects service.
func (c *Client) CustomObjects() *data.ObjectsService {
	return c.customObjects
}

// Companies returns the companies service.
func (c *Client) Companies() *companies.Service {
	return c.companies
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

// Calls returns the calls service.
func (c *Client) Calls() *calls.Service {
	return c.calls
}

// PhoneCallRedirects returns the phone call redirects service.
func (c *Client) PhoneCallRedirects() *calls.RedirectsService {
	return c.phoneCallRedirects
}

// HelpCenter returns the help center service.
func (c *Client) HelpCenter() *helpcenter.Service {
	return c.helpCenter
}

// Articles returns the articles service.
func (c *Client) Articles() *articles.Service {
	return c.articles
}

// InternalArticles returns the internal articles service.
func (c *Client) InternalArticles() *articles.InternalService {
	return c.internalArticles
}

// Contacts returns the contacts service.
func (c *Client) Contacts() *contacts.Service {
	return c.contacts
}

// Visitors returns the visitors service.
func (c *Client) Visitors() *contacts.VisitorsService {
	return c.visitors
}

// Conversations returns the conversations service.
func (c *Client) Conversations() *conversations.Service {
	return c.conversations
}

// Tickets returns the tickets service.
func (c *Client) Tickets() *tickets.Service {
	return c.tickets
}

// TicketTypes returns the ticket types service.
func (c *Client) TicketTypes() *tickets.TypesService {
	return c.ticketTypes
}

// TicketStates returns the ticket states service.
func (c *Client) TicketStates() *tickets.StatesService {
	return c.ticketStates
}

// Workflows returns the workflows service.
func (c *Client) Workflows() *workflows.Service {
	return c.workflows
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
func (c *Client) DoRaw(ctx context.Context, req *http.Request) (*api.Result, error) {
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

	return api.BuildResult(resp, body), nil
}

// DoRawNoRedirect executes an HTTP request like DoRaw, but does not follow
// HTTP redirects. This is used when the API returns a redirect (e.g., 302)
// and the caller needs the Location header rather than the redirect target.
func (c *Client) DoRawNoRedirect(ctx context.Context, req *http.Request) (*api.Result, error) {
	noRedirectClient := &http.Client{
		Transport: c.httpClient.Transport,
		Timeout:   c.httpClient.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req = req.WithContext(ctx)
	c.logger.Debug("http request", "method", req.Method, "url", req.URL)

	resp, err := noRedirectClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return api.BuildResult(resp, body), nil
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
		result := api.BuildResult(resp, body)
		return api.ResultError(result)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

// Do sends an API request and returns the API response. The JSON response
// body is decoded into v if v is non-nil. API errors (4xx/5xx) are returned
// as *ErrorResponse errors suitable for inspection with status predicates
// (IsNotFound, IsBadRequest, etc.).
func (c *Client) Do(ctx context.Context, req *http.Request, v any) (*api.Response, error) {
	result, err := c.DoRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	response := &api.Response{Result: result}

	if result.Error != nil {
		return response, api.ResultError(result)
	}

	if v != nil && result.StatusCode != http.StatusNoContent && len(result.Body) > 0 {
		if err := json.Unmarshal(result.Body, v); err != nil {
			return response, err
		}
	}

	return response, nil
}
