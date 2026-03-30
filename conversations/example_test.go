package conversations_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/conversations"
)

func ExampleService_Get() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/conv-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"conversation","id":"conv-1","title":"Billing question","state":"open"}`)
	})

	conv, err := svc.Get(context.Background(), "conv-1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s Title=%s State=%s\n", conv.ID, conv.Title, conv.State)
	// Output: ID=conv-1 Title=Billing question State=open
}

func ExampleService_Reply() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/conv-1/reply", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"conversation","id":"conv-1","state":"open"}`)
	})

	conv, err := svc.Reply(context.Background(), "conv-1", &conversations.ReplyRequest{
		MessageType: "comment",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "We can help with that!",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Replied to conversation %s\n", conv.ID)
	// Output: Replied to conversation conv-1
}

func ExampleService_ListAll() {
	svc, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{"type":"list","conversations":[{"type":"conversation","id":"conv-1","title":"Help"}],"pages":{"next":{"starting_after":"cursor-2"}}}`)
		} else {
			fmt.Fprint(w, `{"type":"list","conversations":[{"type":"conversation","id":"conv-2","title":"Bug report"}],"pages":{}}`)
		}
	})

	iter := svc.ListAll(context.Background(), nil)
	for iter.Next() {
		c := iter.Current()
		fmt.Printf("%s: %s\n", c.ID, c.Title)
	}
	if err := iter.Err(); err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// conv-1: Help
	// conv-2: Bug report
}
