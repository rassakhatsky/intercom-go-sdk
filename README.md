# intercom-go-sdk

Go SDK for the [Intercom REST API](https://developers.intercom.com/docs/references/rest-api/api.intercom.io/) v2.15.

- Minimal dependencies (stdlib + [`google/go-querystring`](https://github.com/google/go-querystring))
- 167 endpoints across 34 services
- Generic auto-pagination iterator
- Search filter builder

## Installation

```
go get github.com/rassakhatsky/intercom-go-sdk
```

## Quick Start

```go
client := intercom.NewClient("your-bearer-token")
ctx := context.Background()

contact, err := client.Contacts.Get(ctx, "contact-id")
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
// Create
contact, err := client.Contacts.Create(ctx, &intercom.CreateContactRequest{
    Role:  "user",
    Email: "alice@example.com",
    Name:  "Alice",
})

// Update
contact, err = client.Contacts.Update(ctx, contact.ID, &intercom.UpdateContactRequest{
    Name: "Alice Smith",
})

// Delete
deleted, err := client.Contacts.Delete(ctx, contact.ID)
```

## Pagination

```go
// Auto-paginate with iterator
iter := client.Contacts.ListAll(ctx, &intercom.ListOptions{PerPage: 50})
for iter.Next() {
    fmt.Println(iter.Current().Email)
}
if err := iter.Err(); err != nil {
    log.Fatal(err)
}

// Or collect all at once
contacts, err := client.Contacts.ListAll(ctx, &intercom.ListOptions{PerPage: 50}).Collect()
```

Companies use scroll-based pagination via `client.Companies.Scroll(ctx, scrollParam)`.

## Search

```go
// Single filter
result, err := client.Contacts.Search(ctx, &intercom.SearchRequest{
    Query: intercom.SingleFilterOf("email", "=", "alice@example.com"),
})

// Compound filter
result, err = client.Contacts.Search(ctx, &intercom.SearchRequest{
    Query: intercom.And(
        intercom.SingleFilterOf("role", "=", "user"),
        intercom.SingleFilterOf("email", "~", "example.com"),
    ),
})
```

## Error Handling

```go
contact, err := client.Contacts.Get(ctx, "nonexistent")
if err != nil {
    if intercom.IsNotFound(err) {
        fmt.Println("Contact not found")
    } else if intercom.IsRateLimited(err) {
        fmt.Println("Rate limited")
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Raw Methods

Every method has a `*Raw` companion returning `(*Result, error)` with raw response bytes and HTTP metadata. API errors (4xx/5xx) populate `Result.Error` instead of returning a Go error.

```go
result, err := client.Contacts.GetRaw(ctx, "contact-id")
if err != nil {
    log.Fatal(err) // transport error only
}
if result.Error != nil {
    fmt.Println("API error:", result.Error.Error())
    return
}
contact, err := intercom.ParseContactGetResult(result)
```

## Available Services

| Service | Resource |
|---------|----------|
| `Admins` | Admins / teammates |
| `AIContent` | AI content import sources and external pages |
| `Articles` | Help center articles |
| `AwayStatusReasons` | Away status reasons |
| `Brands` | Brands |
| `Calls` | Calls, recordings, transcripts |
| `Companies` | Companies (with scroll pagination) |
| `Contacts` | Contacts (users and leads) |
| `Conversations` | Conversations, replies, parts |
| `CustomChannelEvents` | Custom channel events |
| `CustomObjects` | Custom object instances |
| `DataAttributes` | Data attributes |
| `DataEvents` | Data events |
| `DataExport` | Content data export jobs |
| `Emails` | Email settings |
| `ExportReporting` | Reporting data export |
| `FinVoice` | Fin Voice AI calls |
| `HelpCenter` | Help center collections and centers |
| `InternalArticles` | Internal (team-only) articles |
| `IPAllowlist` | IP allowlist settings |
| `Jobs` | Async job status |
| `Messages` | Messages (in-app, email, push) |
| `News` | News items and newsfeeds |
| `Notes` | Notes |
| `PhoneCallRedirects` | Phone call redirects |
| `Segments` | Segments |
| `SubscriptionTypes` | Subscription types |
| `Tags` | Tags |
| `Teams` | Teams |
| `Tickets` | Tickets |
| `TicketStates` | Ticket states |
| `TicketTypes` | Ticket types and attributes |
| `Visitors` | Visitors |
| `Workflows` | Workflow export |

All services are accessed via `client.<Service>` (e.g., `client.Contacts`, `client.Tickets`).

## Documentation

- [Intercom API Reference (v2.15)](https://developers.intercom.com/docs/references/rest-api/api.intercom.io/)

## License

MIT
