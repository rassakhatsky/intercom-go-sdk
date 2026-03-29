package intercom_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	intercom "github.com/rassakhatsky/intercom-go-sdk"
)

// ExampleContactsService_GetRaw demonstrates using a Raw method to access
// HTTP metadata alongside the typed response data.
func ExampleContactsService_GetRaw() {
	client := intercom.NewClient("your-bearer-token")
	ctx := context.Background()

	result, err := client.Contacts().GetRaw(ctx, "contact-id")
	if err != nil {
		log.Fatal(err) // only returned for transport/network errors
	}

	// Check for API errors (4xx/5xx) — no Go error is returned for these
	if result.Error != nil {
		fmt.Printf("API error %d: %s\n", result.StatusCode, result.Error.Error())
		return
	}

	// Decode the typed response data using the Parse function
	contact, err := intercom.ParseContactGetResult(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Contact:", contact.Name)

	// Access HTTP metadata for debugging or advanced logic
	fmt.Println("Status:", result.StatusCode)
	fmt.Println("Rate-Limit-Remaining:", result.Header.Get("X-RateLimit-Remaining"))

	// Access the raw JSON body
	fmt.Println("Raw body length:", len(result.Body))
}

// ExampleContactsService_SearchRaw demonstrates using a Raw search method
// with access to rate-limit headers for custom retry logic.
func ExampleContactsService_SearchRaw() {
	client := intercom.NewClient("your-bearer-token")
	ctx := context.Background()

	result, err := client.Contacts().SearchRaw(ctx, &intercom.SearchRequest{
		Query: intercom.SingleFilterOf("email", "=", "alice@example.com"),
	})
	if err != nil {
		log.Fatal(err)
	}

	if result.Error != nil {
		fmt.Printf("API error: %s\n", result.Error.Error())
		return
	}

	page, err := intercom.ParseContactSearchResult(result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d contacts\n", len(page.Data))
	fmt.Printf("Rate limit remaining: %s\n", result.Header.Get("X-RateLimit-Remaining"))
}

// ExampleIsRateLimited demonstrates handling rate-limited responses
// with parsed rate limit metadata.
func ExampleIsRateLimited() {
	client := intercom.NewClient("your-bearer-token")
	ctx := context.Background()

	_, err := client.Contacts().Get(ctx, "contact-id")
	if intercom.IsRateLimited(err) {
		var apiErr *intercom.ErrorResponse
		if errors.As(err, &apiErr) && apiErr.RateLimit != nil {
			fmt.Printf("Rate limited. Limit: %d, Remaining: %d\n",
				apiErr.RateLimit.Limit, apiErr.RateLimit.Remaining)
			fmt.Printf("Retry after: %v\n", apiErr.RateLimit.RetryAfter)
			fmt.Printf("Window resets at: %v\n", apiErr.RateLimit.Reset)
			time.Sleep(apiErr.RateLimit.RetryAfter)
		}
	}
}

// ExampleErrorResponse_HasErrorCode demonstrates matching specific
// Intercom API error codes returned in error responses.
func ExampleErrorResponse_HasErrorCode() {
	client := intercom.NewClient("your-bearer-token")
	ctx := context.Background()

	_, err := client.Contacts().Get(ctx, "contact-id")
	if err != nil {
		var apiErr *intercom.ErrorResponse
		if errors.As(err, &apiErr) {
			switch {
			case apiErr.HasErrorCode(intercom.ErrParameterInvalid):
				fmt.Println("Invalid parameter:", apiErr.Error())
			case apiErr.HasErrorCode(intercom.ErrTokenUnauthorized):
				fmt.Println("Token is unauthorized, check your API key")
			case apiErr.HasErrorCode(intercom.ErrRateLimitExceeded):
				fmt.Println("Rate limit exceeded, back off and retry")
			default:
				fmt.Println("API error:", apiErr.Error())
			}
		} else {
			log.Fatal(err) // transport error
		}
	}
}

// ExampleIsServerError demonstrates checking for server-side errors
// to implement retry logic.
func ExampleIsServerError() {
	client := intercom.NewClient("your-bearer-token")
	ctx := context.Background()

	_, err := client.Contacts().Get(ctx, "contact-id")
	if intercom.IsServerError(err) {
		fmt.Println("Server error, retrying...")
		// implement retry with backoff
	} else if intercom.IsNotFound(err) {
		fmt.Println("Contact not found")
	} else if err != nil {
		log.Fatal(err)
	}
}
