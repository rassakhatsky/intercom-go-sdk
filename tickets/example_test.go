package tickets_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/tickets"
)

func ExampleService_Create() {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"ticket","id":"t1","ticket_id":"TICK-001","category":"Customer"}`)
	})

	ticket, err := svc.Create(context.Background(), &tickets.CreateRequest{
		TicketTypeID: "type-1",
		Contacts:     []tickets.ContactRef{{Email: "alice@example.com"}},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s TicketID=%s\n", ticket.ID, ticket.TicketID)
	// Output: ID=t1 TicketID=TICK-001
}

func ExampleService_Get() {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/t1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"ticket","id":"t1","ticket_id":"TICK-001","open":true}`)
	})

	ticket, err := svc.Get(context.Background(), "t1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s Open=%v\n", ticket.ID, ticket.Open)
	// Output: ID=t1 Open=true
}

func ExampleService_Update() {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/t1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"ticket","id":"t1","ticket_id":"TICK-001","ticket_state_id":"resolved"}`)
	})

	ticket, err := svc.Update(context.Background(), "t1", &tickets.UpdateRequest{
		TicketStateID: "resolved",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Updated ticket %s\n", ticket.ID)
	// Output: Updated ticket t1
}
