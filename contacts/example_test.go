package contacts_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/api"
	"github.com/rassakhatsky/intercom-go-sdk/contacts"
)

func ExampleService_Create() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"contact","id":"c1","role":"user","email":"alice@example.com","name":"Alice"}`)
	})

	contact, err := svc.Create(context.Background(), &contacts.CreateRequest{
		Role:  "user",
		Email: "alice@example.com",
		Name:  "Alice",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s Email=%s Name=%s\n", contact.ID, contact.Email, contact.Name)
	// Output: ID=c1 Email=alice@example.com Name=Alice
}

func ExampleService_Get() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/c1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"contact","id":"c1","email":"alice@example.com","name":"Alice"}`)
	})

	contact, err := svc.Get(context.Background(), "c1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s Name=%s\n", contact.ID, contact.Name)
	// Output: ID=c1 Name=Alice
}

func ExampleService_ListAll() {
	svc, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{"type":"list","data":[{"type":"contact","id":"c1","name":"Alice"}],"pages":{"next":{"starting_after":"cursor-2"}}}`)
		} else {
			fmt.Fprint(w, `{"type":"list","data":[{"type":"contact","id":"c2","name":"Bob"}],"pages":{}}`)
		}
	})

	iter := svc.ListAll(context.Background(), nil)
	for iter.Next() {
		c := iter.Current()
		fmt.Println(c.Name)
	}
	if err := iter.Err(); err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// Alice
	// Bob
}

func ExampleService_Search() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"contact","id":"c1","email":"alice@acme.com","name":"Alice"}],"total_count":1,"pages":{}}`)
	})

	filter := api.SingleFilterOf("email", api.OpContains, "@acme.com")
	result, err := svc.Search(context.Background(), &api.SearchRequest{
		Query: filter,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, c := range result.Data {
		fmt.Printf("ID=%s Email=%s\n", c.ID, c.Email)
	}
	// Output: ID=c1 Email=alice@acme.com
}
