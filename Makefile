.PHONY: test test-race fix

test:
	go test -cover ./...

test-race:
	go test -race -cover ./...

fix:
	go vet ./...
	gofmt -l .
