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
//	    Query: intercom.SingleFilterOf("email", intercom.OpEquals, "alice@example.com"),
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
// Helper functions check for common HTTP status codes:
//
//	intercom.IsNotFound(err)            // 404
//	intercom.IsRateLimited(err)         // 429
//	intercom.IsUnauthorized(err)        // 401
//	intercom.IsBadRequest(err)          // 400
//	intercom.IsForbidden(err)           // 403
//	intercom.IsConflict(err)            // 409
//	intercom.IsUnprocessableEntity(err) // 422
//	intercom.IsServerError(err)         // 5xx
//
// For rate-limited responses (429), the ErrorResponse includes parsed
// rate limit headers:
//
//	if intercom.IsRateLimited(err) {
//	    var apiErr *intercom.ErrorResponse
//	    if errors.As(err, &apiErr) && apiErr.RateLimit != nil {
//	        fmt.Printf("Retry after %v\n", apiErr.RateLimit.RetryAfter)
//	        time.Sleep(apiErr.RateLimit.RetryAfter)
//	    }
//	}
//
// Intercom API error codes can be matched using HasErrorCode and the
// predefined ErrorCode constants:
//
//	var apiErr *intercom.ErrorResponse
//	if errors.As(err, &apiErr) {
//	    if apiErr.HasErrorCode(intercom.ErrParameterInvalid) {
//	        // handle invalid parameter
//	    }
//	}
//
// Core types (Result, Iter, Filter, ErrorResponse, ListOptions) live in the
// public api/ package. Sub-package-specific types (request/response structs,
// domain models) are imported from their respective packages.
package intercom
