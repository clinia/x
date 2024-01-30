.PHONY: test
test:
	go test `go list ./... | grep -v pubsubx/kafkax`

test-kafka:
	go test -v ./pubsubx/kafkax/...

.PHONY: lint
lint:
	golangci-lint run