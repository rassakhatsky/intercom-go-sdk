package tags_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rassakhatsky/intercom-go-sdk/tags"
)

func ExampleService_CreateOrUpdate() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"tag","id":"t1","name":"VIP"}`)
	})

	tag, err := svc.CreateOrUpdate(context.Background(), &tags.CreateOrUpdateRequest{
		Name: "VIP",
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("ID=%s Name=%s\n", tag.ID, tag.Name)
	// Output: ID=t1 Name=VIP
}

func ExampleService_List() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"tag","id":"t1","name":"VIP"},{"type":"tag","id":"t2","name":"Beta"}]}`)
	})

	list, err := svc.List(context.Background())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for _, t := range list.Data {
		fmt.Printf("%s: %s\n", t.ID, t.Name)
	}
	// Output:
	// t1: VIP
	// t2: Beta
}

func ExampleService_TagCompany() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"tag","id":"t1","name":"Enterprise"}`)
	})

	tag, err := svc.TagCompany(context.Background(), &tags.TagCompanyRequest{
		Name:      "Enterprise",
		Companies: []tags.TagCompanyItem{{ID: "company-1"}},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Tagged with: %s\n", tag.Name)
	// Output: Tagged with: Enterprise
}

func ExampleService_UntagCompany() {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"type":"tag","id":"t1","name":"Enterprise"}`)
	})

	tag, err := svc.UntagCompany(context.Background(), &tags.UntagCompanyRequest{
		Name:      "Enterprise",
		Companies: []tags.UntagCompanyItem{{ID: "company-1", Untag: true}},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("Untagged: %s\n", tag.Name)
	// Output: Untagged: Enterprise
}
