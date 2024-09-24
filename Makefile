.PHONY: test
test:
	go test `go list ./... | grep -v pubsubx/kgox`

test_kafka:
	go test -v ./pubsubx/kgox/...

test_kafka_race:
	go test ./pubsubx/kgox/... -short -race

.PHONY: lint
lint:
	golangci-lint run

generate-triton:
	./scripts/generate-triton.sh


