GO ?= go
SHELVIA ?= $(GO) run ./cmd/shelvia
DEBUG_SHELF ?= fixtures/shelves/valid/minimal
WHERE ?= rating >= 90

.PHONY: test test-v test-pretty test-cover vet debug-validate debug-list debug-query debug-query-novel debug

test:
	@$(GO) test ./...

test-v:
	@$(GO) test -v ./...

test-cover:
	@$(GO) test ./... -cover

vet:
	@$(GO) vet ./...

debug: debug-validate debug-list debug-query

debug-validate:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) validate

debug-list:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) list

debug-query:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) query --where '$(WHERE)'

debug-query-novel:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) query --where 'genre = "Novel"'
