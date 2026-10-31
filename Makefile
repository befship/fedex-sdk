.PHONY: generate fetch-specs test lint fmt tidy clean

# Preferred: POSTMAN_DIR=... or SPEC_LOCAL=...
generate:
	bash generate.sh

fetch-specs:
	FORCE_FETCH=1 bash generate.sh

test:
	go test ./... -count=1

lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed"; exit 1; }
	golangci-lint run ./...

fmt:
	gofmt -w fedex examples
	find internal/generated -name '*.go' -print0 2>/dev/null | xargs -0 gofmt -w

tidy:
	go mod tidy

clean:
	rm -rf .cache/postman .cache/specs
