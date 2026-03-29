// Package intercom provides a Go client for the Intercom API v2.15.
//
// Services are organized into domain-scoped sub-packages (contacts, tickets,
// tags, etc.) and accessed via accessor methods on the Client.
//
// Usage:
//
//	client := intercom.NewClient("your-bearer-token")
//
//	// Get a contact
//	contact, err := client.Contacts().Get(ctx, "contact-id")
//
//	// List all contacts with auto-pagination
//	iter := client.Contacts().ListAll(ctx, nil)
//	for iter.Next() {
//	    fmt.Println(iter.Current().Name)
//	}
//	if err := iter.Err(); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Or collect all pages into a slice in one call
//	contacts, err := client.Contacts().ListAll(ctx, nil).Collect()
//
//	// Iterate with a callback using ForEach
//	err = client.Contacts().ListAll(ctx, nil).ForEach(func(c contacts.Contact) error {
//	    fmt.Println(c.Name)
//	    return nil
//	})
//
//	// Search contacts with filters
//	result, err := client.Contacts().Search(ctx, &intercom.SearchRequest{
//	    Query: intercom.SingleFilterOf("email", "=", "alice@example.com"),
//	})
//
// Every service method has a Raw companion that returns a *Result with
// HTTP metadata (status code, headers, raw body bytes).
// API errors (4xx/5xx) populate Result.Error instead of returning a Go error:
//
//	result, err := client.Contacts().GetRaw(ctx, "contact-id")
//	if err != nil {
//	    log.Fatal(err) // transport error only
//	}
//	if result.Error != nil {
//	    fmt.Printf("API error %d: %s\n", result.StatusCode, result.Error.Error())
//	} else {
//	    contact, _ := contacts.ParseGetResult(result)
//	    fmt.Println(contact.Name)
//	    fmt.Println(result.Header.Get("X-RateLimit-Remaining"))
//	}
//
// The client can be configured with functional options:
//
//	client := intercom.NewClient("token",
//	    intercom.WithHTTPClient(customHTTPClient),
//	    intercom.WithLogger(myLogger),
//	    intercom.WithBaseURL("https://custom.url/"),
//	)
//
// All API methods accept a context.Context as their first parameter for
// cancellation and timeout control.
//
// Error responses from the Intercom API are returned as *ErrorResponse values.
// Helper functions IsNotFound, IsRateLimited, and IsUnauthorized can be used
// to check for common error types.
//
// Core types (Result, Iter, Filter, ErrorResponse, ListOptions) are
// re-exported as type aliases in this root package so consumers do not need
// to import internal/api directly. Sub-package-specific types (request/response
// structs, domain models) are imported from their respective packages.
package intercom
