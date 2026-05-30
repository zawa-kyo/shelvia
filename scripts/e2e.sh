#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd "$(dirname "$0")" && pwd)"
repo_root="$(CDPATH= cd "$script_dir/.." && pwd)"
cd "$repo_root"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

bin="$tmp_dir/shelvia"
shelf="$tmp_dir/shelf"
fixture_shelf="${E2E_FIXTURE_SHELF:-fixtures/shelves/valid/minimal}"

go build -o "$bin" ./cmd/shelvia

contains() {
	output=$1
	expected=$2
	printf '%s' "$output" | grep -F "$expected" >/dev/null
}

line_number() {
	output=$1
	expected=$2
	printf '%s' "$output" | grep -n -F "$expected" | head -n 1 | cut -d: -f1
}

assert_contains() {
	output=$1
	expected=$2
	context=$3
	if ! contains "$output" "$expected"; then
		printf 'e2e failed: expected %s to contain %s\n' "$context" "$expected" >&2
		printf '%s\n' "$output" >&2
		exit 1
	fi
}

assert_order() {
	output=$1
	before=$2
	after=$3
	before_line=$(line_number "$output" "$before")
	after_line=$(line_number "$output" "$after")
	if [ -z "$before_line" ] || [ -z "$after_line" ] || [ "$before_line" -ge "$after_line" ]; then
		printf 'e2e failed: expected %s to appear before %s\n' "$before" "$after" >&2
		printf '%s\n' "$output" >&2
		exit 1
	fi
}

book_count() {
	root=$1
	count=0
	for path in "$root"/*.toml; do
		[ -e "$path" ] || continue
		[ "${path##*/}" = "config.toml" ] && continue
		count=$((count + 1))
	done
	printf '%s' "$count"
}

plural_suffix() {
	count=$1
	if [ "$count" -eq 1 ]; then
		printf ''
	else
		printf 's'
	fi
}

init_output="$("$bin" init "$shelf")"
assert_contains "$init_output" "Initialized shelf." "init output"

validate_output="$(SHELVIA_DIR="$shelf" "$bin" validate)"
assert_contains "$validate_output" "Validated 1 book, 1 config file." "validate output"

list_output="$(SHELVIA_DIR="$shelf" "$bin" list)"
assert_contains "$list_output" "Example Book" "list output"
assert_contains "$list_output" "read_date" "list output"

query_output="$(SHELVIA_DIR="$shelf" "$bin" query --where 'rating >= 80')"
assert_contains "$query_output" "Example Book" "query output"

fixture_validate_output="$(SHELVIA_DIR="$fixture_shelf" "$bin" validate)"
fixture_book_count=$(book_count "$fixture_shelf")
fixture_plural_suffix=$(plural_suffix "$fixture_book_count")
assert_contains "$fixture_validate_output" "Validated $fixture_book_count book$fixture_plural_suffix, 1 config file." "fixture validate output"

sort_output="$(SHELVIA_DIR="$fixture_shelf" "$bin" query --where 'rating >= 0 order by rating desc')"
assert_order "$sort_output" "Top Rated Book" "Some Book"
assert_order "$sort_output" "Some Book" "Example Book"
assert_order "$sort_output" "Example Book" "Mid Rated Book"
assert_order "$sort_output" "Mid Rated Book" "Low Rated Book"

if SHELVIA_DIR="$shelf" "$bin" query --where 'select * from books' >/dev/null 2>&1; then
	printf 'e2e failed: unsupported query unexpectedly succeeded\n' >&2
	exit 1
fi

printf 'e2e passed\n'
