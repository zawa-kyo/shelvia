GO ?= go
GOTESTSUM ?= gotestsum
SHELVIA ?= $(GO) run ./cmd/shelvia
DEBUG_SHELF ?= fixtures/shelves/valid/minimal
WHERE ?= rating >= 90

.PHONY: test test-v test-pretty test-cover vet debug-validate debug-list debug-query debug-query-novel debug

test:
	@$(GO) test ./...

test-v:
	@$(GO) test -v ./...

test-pretty:
	@command -v $(GOTESTSUM) >/dev/null 2>&1 || { echo "gotestsum is required. Install: go install gotest.tools/gotestsum@latest"; exit 1; }
	@$(GOTESTSUM) --format testname -- ./...

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
