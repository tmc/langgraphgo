# This file contains convenience targets for the project.
# It is not intended to be used as a build system.
# See the README for more information.

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: lint-deps
	golangci-lint run --color=always --sort-results ./...

.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix --skip-dirs=./exp ./...

.PHONY: test-race
test-race:
	go run test -race ./...

.PHONY: test-cover
test-cover:
	go run test -cover ./...

.PHONY: lint-deps
lint-deps:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo >&2 "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.1; \
	}

