.PHONY: test
test:
	go test `go list ./... | grep -v pubsubx/kgox`

.PHONY: test_kafka
test_kafka:
	go test -v ./pubsubx/kgox/...

.PHONY: test_kafka_race
test_kafka_race:
	go test -v ./pubsubx/kgox/... -short -race

.PHONY: lint
lint:
	golangci-lint run --tests=false

.PHONY: format
format:
	gofumpt -l -w .

.PHONY: generate-triton
generate-triton:
	./scripts/generate-triton.sh

.PHONY: install
install:
	go install mvdan.cc/gofumpt@latest