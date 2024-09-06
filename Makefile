.PHONY: test
test:
	go test `go list ./... | grep -v pubsubx/kafkax`

test_kafka:
	go test -v ./pubsubx/kafkax/...

test_kafka_short:
	go test ./pubsubx/kafkax/... -short

test_kafka_race:
	go test ./pubsubx/kafkax/... -short -race

.PHONY: lint
lint:
	golangci-lint run

generate-triton:
	./scripts/generate-triton.sh


