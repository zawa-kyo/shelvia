GO ?= go
LOCAL_BIN := $(CURDIR)/.bin
GOTESTSUM ?= $(LOCAL_BIN)/gotestsum
SHELVIA ?= $(GO) run ./cmd/shelvia
DEMO_SHELF ?= fixtures/shelves/valid/minimal
WHERE ?= rating >= 90

.PHONY: install test test-v test-pretty test-cover vet e2e demo-validate demo-list demo-query demo-sort demo-query-novel demo

install:
	@mkdir -p "$(LOCAL_BIN)"
	@GOBIN="$(LOCAL_BIN)" $(GO) install gotest.tools/gotestsum

test:
	@$(GO) test ./...

test-v:
	@$(GO) test -v ./...

test-pretty:
	@command -v "$(GOTESTSUM)" >/dev/null 2>&1 || { echo "gotestsum is required. Run: make install"; exit 1; }
	@$(GOTESTSUM) --format testname -- ./...

test-cover:
	@$(GO) test ./... -cover

vet:
	@$(GO) vet ./...

e2e:
	@$(GO) test -tags=e2e ./e2e

demo: demo-validate demo-list demo-query

demo-validate:
	@SHELVIA_DIR="$(DEMO_SHELF)" $(SHELVIA) validate

demo-list:
	@SHELVIA_DIR="$(DEMO_SHELF)" $(SHELVIA) list

demo-query:
	@SHELVIA_DIR="$(DEMO_SHELF)" $(SHELVIA) query --where '$(WHERE)'

demo-sort:
	@SHELVIA_DIR="$(DEMO_SHELF)" $(SHELVIA) query --where 'rating >= 0 order by rating desc'

demo-query-novel:
	@SHELVIA_DIR="$(DEMO_SHELF)" $(SHELVIA) query --where 'genre = "Novel"'
