# intercom-go-sdk

Go SDK for the [Intercom REST API](https://developers.intercom.com/docs/references/rest-api/api.intercom.io/) v2.15.

- Minimal dependencies (stdlib + [`google/go-querystring`](https://github.com/google/go-querystring))
- 167 endpoints across 34 services organized in 16 domain-scoped sub-packages
- Generic auto-pagination iterator
- Search filter builder

## Requirements

Go 1.26.1 or later.

## Installation

```
go get github.com/rassakhatsky/intercom-go-sdk
```

## Quick Start

```go
import (
    "context"
    "fmt"

    intercom "github.com/rassakhatsky/intercom-go-sdk"
)

client := intercom.NewClient("your-bearer-token")
ctx := context.Background()

contact, err := client.Contacts().Get(ctx, "contact-id")
```

### Client Options

```go
client := intercom.NewClient("token",
    intercom.WithHTTPClient(customHTTPClient),
    intercom.WithLogger(myLogger),
    intercom.WithBaseURL("https://custom.url/"),
)
```

## CRUD

```go
import "github.com/rassakhatsky/intercom-go-sdk/contacts"

// Create
contact, err := client.Contacts().Create(ctx, &contacts.CreateRequest{
    Role:  "user",
    Email: "alice@example.com",
    Name:  "Alice",
})

// Update
contact, err = client.Contacts().Update(ctx, contact.ID, &contacts.UpdateRequest{
    Name: "Alice Smith",
})

// Delete
deleted, err := client.Contacts().Delete(ctx, contact.ID)
```

## Pagination

```go
// Auto-paginate with iterator
iter := client.Contacts().ListAll(ctx, &intercom.ListOptions{PerPage: 50})
for iter.Next() {
    fmt.Println(iter.Current().Email)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}

// Or collect all at once
contacts, err := client.Contacts().ListAll(ctx, &intercom.ListOptions{PerPage: 50}).Collect()

// Callback-based iteration
err := client.Contacts().ListAll(ctx, nil).ForEach(func(c contacts.Contact) error {
    fmt.Println(c.Name)
    return nil // return non-nil error to stop early
})
```

Companies use scroll-based pagination via `client.Companies().Scroll(ctx, scrollParam)`.

## Search

```go
// Single filter
result, err := client.Contacts().Search(ctx, &intercom.SearchRequest{
    Query: intercom.SingleFilterOf("email", "=", "alice@example.com"),
})

// Compound filter
result, err = client.Contacts().Search(ctx, &intercom.SearchRequest{
    Query: intercom.And(
        intercom.SingleFilterOf("role", "=", "user"),
        intercom.SingleFilterOf("email", "~", "example.com"),
    ),
})
```

## Error Handling

```go
contact, err := client.Contacts().Get(ctx, "nonexistent")
if err != nil {
    if intercom.IsNotFound(err) {
        fmt.Println("Contact not found")
    } else if intercom.IsRateLimited(err) {
        fmt.Println("Rate limited")
    } else if intercom.IsServerError(err) {
        fmt.Println("Server error, retry with backoff")
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

Additional status helpers: `IsUnauthorized` (401), `IsBadRequest` (400), `IsForbidden` (403), `IsConflict` (409), `IsUnprocessableEntity` (422).

### Rate Limit Details

Rate-limited (429) responses include parsed rate limit metadata:

```go
if intercom.IsRateLimited(err) {
    var apiErr *intercom.ErrorResponse
    if errors.As(err, &apiErr) && apiErr.RateLimit != nil {
        fmt.Printf("Retry after %v\n", apiErr.RateLimit.RetryAfter)
        time.Sleep(apiErr.RateLimit.RetryAfter)
    }
}
```

### Error Code Matching

Match specific Intercom API error codes using `HasErrorCode` and predefined `ErrorCode` constants:

```go
var apiErr *intercom.ErrorResponse
if errors.As(err, &apiErr) {
    if apiErr.HasErrorCode(intercom.ErrParameterInvalid) {
        fmt.Println("Invalid parameter")
    }
}
```

## Raw Methods

Every method has a `*Raw` companion returning `(*Result, error)` with raw response bytes and HTTP metadata. API errors (4xx/5xx) populate `Result.Error` instead of returning a Go error.

```go
import "github.com/rassakhatsky/intercom-go-sdk/contacts"

result, err := client.Contacts().GetRaw(ctx, "contact-id")
if err != nil {
    log.Fatal(err) // transport error only
}
if result.Error != nil {
    fmt.Println("API error:", result.Error.Error())
    return
}
contact, err := contacts.ParseGetResult(result)

// Or use the generic Decode helper
contact, err = intercom.Decode[contacts.Contact](result)
```

## Sub-Packages

Services are organized into domain-scoped sub-packages. Each service is accessed via an accessor method on the client (e.g., `client.Contacts()`, `client.Tags()`).

| Package | Accessor Methods | Description |
|---------|-----------------|-------------|
| `contacts` | `Contacts()`, `Visitors()` | Contacts (users/leads) and visitors |
| `conversations` | `Conversations()` | Conversations, replies, parts |
| `companies` | `Companies()` | Companies (with scroll pagination) |
| `tickets` | `Tickets()`, `TicketTypes()`, `TicketStates()` | Tickets and ticket configuration |
| `articles` | `Articles()`, `InternalArticles()` | Help center and internal articles |
| `helpcenter` | `HelpCenter()` | Help center collections and centers |
| `news` | `News()` | News items and newsfeeds |
| `ai` | `AIContent()`, `FinVoice()` | AI content and Fin Voice |
| `calls` | `Calls()`, `PhoneCallRedirects()` | Calls and phone call redirects |
| `tags` | `Tags()` | Tags |
| `admins` | `Admins()`, `Teams()`, `AwayStatusReasons()` | Admins, teams, away status |
| `segments` | `Segments()` | Segments |
| `data` | `DataEvents()`, `DataAttributes()`, `CustomObjects()` | Data events, attributes, custom objects |
| `messaging` | `Messages()`, `Emails()`, `SubscriptionTypes()` | Messages, emails, subscriptions |
| `export` | `DataExport()`, `ExportReporting()` | Data and reporting exports |
| `settings` | `Brands()`, `IPAllowlist()`, `CustomChannelEvents()`, `Jobs()`, `Notes()` | Account settings |
| `workflows` | `Workflows()` | Workflow export |

Core types (`Result`, `Iter[T]`, `Filter`, `ErrorResponse`, `ListOptions`, etc.) are re-exported as type aliases in the root `intercom` package.

### Importing Sub-Package Types

When you need request/response types from a specific service, import the sub-package directly:

```go
import (
    intercom "github.com/rassakhatsky/intercom-go-sdk"
    "github.com/rassakhatsky/intercom-go-sdk/contacts"
    "github.com/rassakhatsky/intercom-go-sdk/tags"
)

// Use sub-package types for requests
contact, err := client.Contacts().Create(ctx, &contacts.CreateRequest{
    Role:  "user",
    Email: "alice@example.com",
})

// Use root aliases for shared types
iter := client.Tags().ListAll(ctx, &intercom.ListOptions{PerPage: 25})
```

## Documentation

- [Intercom API Reference (v2.15)](https://developers.intercom.com/docs/references/rest-api/api.intercom.io/)

## License

MIT
