GO ?= go
SHELVIA ?= $(GO) run ./cmd/shelvia
DEBUG_SHELF ?= fixtures/shelves/valid/minimal
WHERE ?= rating >= 90

.PHONY: debug-validate debug-list debug-query debug-query-novel debug

debug: debug-validate debug-list debug-query

debug-validate:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) validate

debug-list:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) list

debug-query:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) query --where '$(WHERE)'

debug-query-novel:
	@SHELVIA_DIR="$(DEBUG_SHELF)" $(SHELVIA) query --where 'genre = "Novel"'
