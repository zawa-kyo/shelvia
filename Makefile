GO ?= go
LOCAL_BIN := $(CURDIR)/.bin
GOTESTSUM ?= $(LOCAL_BIN)/gotestsum
SHELVIA ?= $(GO) run ./cmd/shelvia
DEBUG_SHELF ?= fixtures/shelves/valid/minimal
WHERE ?= rating >= 90

.PHONY: install-tools test test-v test-pretty test-cover vet debug-validate debug-list debug-query debug-sort debug-query-novel debug

install-tools:
	@mkdir -p "$(LOCAL_BIN)"
	@GOBIN="$(LOCAL_BIN)" $(GO) install gotest.tools/gotestsum

test:
	@$(GO) test ./...

test-v:
	@$(GO) test -v ./...

test-pretty:
	@command -v "$(GOTESTSUM)" >/dev/null 2>&1 || { echo "gotestsum is required. Run: make install-tools"; exit 1; }
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

debug-sort:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) query --where 'rating >= 0 order by rating desc'

debug-query-novel:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) query --where 'genre = "Novel"'
