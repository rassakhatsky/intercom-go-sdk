package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

func ExampleDecode() {
	// Simulate a successful API response with a JSON body.
	result := &api.Result{
		StatusCode: 200,
		Body:       []byte(`{"id":"42","name":"Alice"}`),
	}

	type User struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	user, err := api.Decode[User](result)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s Name=%s\n", user.ID, user.Name)
	// Output: ID=42 Name=Alice
}

func ExampleDecode_emptyBody() {
	// A 204 No Content response with no body returns a zero-value struct.
	result := &api.Result{
		StatusCode: 204,
		Body:       nil,
	}

	type Resp struct{ OK bool }

	resp, err := api.Decode[Resp](result)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("OK=%v\n", resp.OK)
	// Output: OK=false
}

func ExampleSingleFilterOf() {
	// Build a filter that matches contacts with role "user".
	f := api.SingleFilterOf("role", api.OpEquals, "user")

	data, _ := json.Marshal(f)
	fmt.Println(string(data))
	// Output: {"field":"role","operator":"=","value":"user"}
}

func ExampleAnd() {
	// Combine two filters with AND — both must match.
	f := api.And(
		api.SingleFilterOf("role", api.OpEquals, "user"),
		api.SingleFilterOf("name", api.OpStarts, "A"),
	)

	data, _ := json.Marshal(f)
	fmt.Println(string(data))
	// Output: {"operator":"AND","value":[{"field":"role","operator":"=","value":"user"},{"field":"name","operator":"^","value":"A"}]}
}

func ExampleOr() {
	// Combine two filters with OR — either can match.
	f := api.Or(
		api.SingleFilterOf("email", api.OpContains, "@acme.com"),
		api.SingleFilterOf("email", api.OpContains, "@example.com"),
	)

	data, _ := json.Marshal(f)
	fmt.Println(string(data))
	// Output: {"operator":"OR","value":[{"field":"email","operator":"~","value":"@acme.com"},{"field":"email","operator":"~","value":"@example.com"}]}
}

func ExampleNewIter() {
	// Create a mock page fetcher that returns two pages.
	fetcher := func(_ context.Context, opts *api.ListOptions) (*api.PagedResult[string], error) {
		switch opts.StartingAfter {
		case "":
			return &api.PagedResult[string]{
				Data: []string{"a", "b"},
				Pages: api.CursorPages{
					Next: &api.StartingAfterPage{StartingAfter: "cursor-2"},
				},
			}, nil
		case "cursor-2":
			return &api.PagedResult[string]{
				Data:  []string{"c"},
				Pages: api.CursorPages{},
			}, nil
		default:
			return nil, fmt.Errorf("unexpected cursor")
		}
	}

	iter := api.NewIter(context.Background(), nil, fetcher)

	// Idiomatic iteration: Next/Current/Err loop.
	for iter.Next() {
		fmt.Println(iter.Current())
	}
	if err := iter.Err(); err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// a
	// b
	// c
}

func ExampleIter_Collect() {
	fetcher := func(_ context.Context, opts *api.ListOptions) (*api.PagedResult[string], error) {
		if opts.StartingAfter == "" {
			return &api.PagedResult[string]{
				Data:  []string{"x", "y", "z"},
				Pages: api.CursorPages{},
			}, nil
		}
		return nil, fmt.Errorf("unexpected cursor")
	}

	iter := api.NewIter(context.Background(), nil, fetcher)
	items, err := iter.Collect()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(items)
	// Output: [x y z]
}

func ExampleIsNotFound() {
	// Construct an ErrorResponse that represents a 404.
	err := &api.ErrorResponse{
		StatusCode: http.StatusNotFound,
		Errors: []api.ErrorDetail{
			{Code: "not_found", Message: "Contact not found"},
		},
	}

	fmt.Println(api.IsNotFound(err))
	fmt.Println(api.IsRateLimited(err))
	// Output:
	// true
	// false
}

func ExampleIsRateLimited() {
	err := &api.ErrorResponse{
		StatusCode: http.StatusTooManyRequests,
		Errors: []api.ErrorDetail{
			{Code: api.ErrRateLimitExceeded, Message: "Rate limit exceeded"},
		},
	}

	fmt.Println(api.IsRateLimited(err))
	// Output: true
}
