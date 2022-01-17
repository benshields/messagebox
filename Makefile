.DEFAULT_GOAL := all

GO_TEST_FLAGS ?= -v -cover
GO_PACKAGES   ?= $(shell go list ./... | grep -v vendor)

.PHONY: all
all: fmt test

.PHONY: fmt
fmt:
	@go fmt $(GO_PACKAGES)

.PHONY: test
test: vendor
	@go test $(GO_TEST_FLAGS) $(GO_PACKAGES)

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
