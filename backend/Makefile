.PHONY: build generate test test-unit test-integration test-e2e financial-gate

VERSION ?= $(shell ./scripts/resolve-version.sh)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

generate:
	go generate ./ent
	go generate ./cmd/server

test:
	go test ./...
	golangci-lint run ./...

test-unit:
	go test -tags=unit ./...

test-integration:
	go test -tags=integration ./...

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	go test -tags=e2e -v -timeout=300s ./internal/integration/...

# DATABASE_URL must point at an explicitly selected PostgreSQL database.
# Add FINANCIAL_GATE_ARGS="-run-scenarios" only for a new, empty local database.
financial-gate:
	go run ./cmd/financialgate $(FINANCIAL_GATE_ARGS)
