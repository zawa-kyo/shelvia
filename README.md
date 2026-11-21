# Shelvia

Shelvia is a CLI tool for managing your reading log locally. It stores each finished book as an individual TOML file, allowing you to validate, list, and search your collection directly from the terminal (unread books and books in progress are out of scope).

Your reading data is never trapped inside an application database. Instead, it remains plain text that you can open in any text editor, track with Git, and inspect with standard command-line tools like `grep` and `find`. On top of that, Shelvia enables rich, structured querying across ratings, completion dates, genres, publishers, and imprints.

## Why Shelvia?

You can track your reading in a spreadsheet or a hosted book-tracking service. Both are convenient, but over time you may want to keep complete ownership of your data, avoid dependence on a proprietary service or UI, and query your history freely.

On the other hand, plain unstructured notes (such as raw Markdown files) are easy for humans to read, but answering structured questions becomes tedious:

- Which books did I rate 90 or higher?
- Which technical books did I read in 2024?
- Which books came from a specific publisher or imprint?

Shelvia bridges this gap, combining the simplicity and durability of plain text with the structured query power of a database.

## Features

- One TOML file per book
- Store and version your reading data directly in your own Git repository
- Initialize a shelf with `config.toml` and `example.toml` via `shelvia init`
- Scaffold new book files from templates with `shelvia add`
- Validate required fields, rating ranges, dates, and controlled vocabularies (genres, publishers, editions, imprints)
- Flexible filtering using SQL-like expressions
- Plain TOML as the single source of truth; SQLite used purely as an ephemeral in-memory query view

## Installation

Installation instructions will be provided prior to the initial release.

## Quick Start

Create a directory for your reading data. Shelvia calls this directory the `shelf root`.

```bash
shelvia init ./my-shelf
```

`init` initializes the directory as a `shelf root` and generates `config.toml` and `example.toml` (creating the directory if it does not already exist).
`example.toml` is a valid sample book included in validation. Both files provide a minimal working baseline with comments explaining how to customize them.

```text
my-shelf/
  config.toml
  example.toml
```

Set the `shelf root` environment variable:

```bash
export SHELVIA_DIR=./my-shelf
```

Generate a template for a new book:

```bash
shelvia add "Some Book" --date 2024-01-01
```

`add` generates `Some Book.toml` with the specified title prefilled. Open the file to enter your book details:

```toml
title = "Some Book"
author = ["Some Author"]
rating = 90
read_date = 2024-01-01

genre = ["Novel"]
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"

[thoughts]
summary = "A short note."
body = """
Longer thoughts can live here.
"""
```

Add your allowed values to `config.toml`. This configuration file prevents naming inconsistencies across genres, publishers, editions, and imprints.

```toml
# List the values you want to use in this shelf.
kind = "shelvia-config"

[values]
# Add the genres you use.
genres = [
  "Novel",
]

# Add the publishers you use.
publishers = [
  "Example Publisher",
]

# Add imprint names if you use them. Leave this as [] if not needed yet.
imprints = [
  "Example Paperback",
]

# Add the editions you use.
# Set imprint_required = true when that edition always needs an imprint.
[[values.editions]]
name = "Paperback"
imprint_required = true

[[values.editions]]
name = "Hardcover"
imprint_required = false
```

Validate your shelf data:

```bash
shelvia validate
```

List your books:

```bash
shelvia list
```

`list` sorts books by `read_date` descending (newest first), resolving ties by `title` ascending. The displayed columns are `read_date`, `rating`, `title`, `author`, `genre`, and `publisher`.

Filter books with search conditions:

```bash
shelvia search 'rating >= 90'
shelvia search 'genre = "Novel" and publisher = "Example Publisher"'
shelvia search 'rating >= 90 order by rating desc'
```

`search` accepts filter conditions targeting `title`, `author`, `rating`, `read_date`, `genre`, `publisher`, `edition`, `imprint`, `series`, and `translator`. Append `order by <column> [asc|desc]` to change the sort order (defaults to the same sort order and columns as `list`).

Inspect a single book by providing its `title` (not its filename):

```bash
shelvia show "Some Book"
```

Find the path of a book's TOML file by its `title`:

```bash
shelvia path "Some Book"
```

If multiple books share the same `title`, both `show` and `path` exit with an error and list all matching file paths.

Export your list or search results to CSV format. To export all fields for all books, use `export`:

```bash
shelvia list --format csv > books.csv
shelvia search 'rating >= 90' --format csv > selected.csv
shelvia export > all-books.csv
```

CSV exports include headers. Omitted optional fields are rendered as blank cells (an empty `translator` array renders as `[]`). The `export` command outputs all book data, including reader thoughts (`thoughts`), while excluding config files and TOML comments.

## Data Format

The initial release tracks a single completion date per book record (re-read histories are not supported).

Book records are written in TOML. The required fields are:

- `title`: Book title
- `author`: Author(s) (array of one or more non-empty strings)
- `rating`: Integer from `0` to `100`
- `read_date`: Completion date (TOML date format)
- `genre`: Genre(s) (array of one or more non-empty strings, validated against `config.toml`)
- `publisher`: Publisher (validated against `config.toml`)

Optional fields are:

- `edition`: Format or edition (validated against `config.toml`)
- `imprint`: Imprint (validated against `config.toml`)
- `series`: Series name
- `translator`: Translator(s) (array of non-empty strings)
- `thoughts.summary`: Summary of your impressions
- `thoughts.body`: Extended thoughts or notes

`author` and `genre` require arrays containing at least one string. `translator` is an optional array of strings. Array elements cannot be blank (single-string formats are supported for backward compatibility). Search filters match against individual array elements. In CSV output, arrays are serialized as JSON arrays within a single cell.

Optional fields can be omitted entirely; specifying an empty string is treated the same as omitting the field.

## `config.toml`

`config.toml` acts as a controlled vocabulary to ensure consistency across genres, publishers, editions, and imprints. Shelvia loads these definitions from the `config.toml` located directly at the `shelf root`.

The `config.toml` generated by `init` is a minimal valid configuration that satisfies `example.toml`. You can expand it with your own allowed values. Lists for unused fields can remain empty (`[]`).

```toml
# List the values you want to use in this shelf.
kind = "shelvia-config"

[values]
# Add the genres you use.
genres = [
  "Novel",
]

# Add the publishers you use.
publishers = [
  "Example Publisher",
]

# Add imprint names if you use them. Leave this as [] if not needed yet.
imprints = [
  "Example Paperback",
]

# Add the editions you use.
# Set imprint_required = true when that edition always needs an imprint.
[[values.editions]]
name = "Paperback"
imprint_required = true

[[values.editions]]
name = "Hardcover"
imprint_required = false
```

When an edition specifies `imprint_required = true`, books using that edition must also specify an `imprint`. Furthermore, specifying an `imprint` requires an `edition` to be present.

## `shelf root`

Shelvia discovers books by recursively reading `.toml` files under a specified `shelf root`. `config.toml` located directly at the `shelf root` is treated exclusively as the configuration file. Any `config.toml` located in a subdirectory is considered misplaced and triggers an error.

```text
my-shelf/
  config.toml
  example.toml
  Some Book.toml
```

Shelvia resolves the `shelf root` using the following precedence:

1. Explicit command arguments (such as `init PATH`) or the `--shelf` option
2. The `SHELVIA_DIR` environment variable

Explicit command-line arguments always take precedence over `SHELVIA_DIR`. If neither is provided, Shelvia exits with an error.

```bash
export SHELVIA_DIR=~/reading-log
shelvia validate
shelvia --shelf ~/other-reading-log validate
```

## Validation

`shelvia validate` checks that the entire `shelf root` can be loaded cleanly. It reads all book files and `config.toml`, verifying TOML syntax, required fields, data types, rating boundaries, date formats, configuration references, and edition/imprint rules. It does not create, modify, or convert any files.

```bash
shelvia validate
```

If validation succeeds, Shelvia reports the total count of verified files:

```text
Validated 1 book, 1 config file.
```

If validation fails, Shelvia outputs the file path, location, and reason for the error:

```text
Some Book.toml:5: unknown genre "Novel"
```

A shelf that passes `validate` is guaranteed to load properly in `list` and `search`. Use `validate` as a pre-commit check after creating or editing book records.
