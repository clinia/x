.PHONY: test
test:
	go test -p 1 ./...

.PHONY: lint
lint:
	golangci-lint run