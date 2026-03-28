package intercom_test

import (
	"context"
	"fmt"
	"log"

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
