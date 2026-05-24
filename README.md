# Shelvia

Shelvia is a CLI tool for managing your reading log on your own machine. It stores each book as a TOML file and lets you validate, list, and search your records from the command line.

Your reading data is not locked inside an application database. It stays as plain text that you can read, edit, track with Git, and search with tools like `grep` and `find`. Shelvia adds structured validation and querying on top, so fields like rating, read date, genre, publisher, and imprint remain useful later.

## Why Shelvia?

You can keep a reading log in a spreadsheet or a hosted book-tracking service. Both are useful, but over time you may want the data to stay in your own hands, avoid depending on a specific service or UI, and search it in your own way.

Free-form notes, such as Markdown, solve the ownership problem but make structured questions harder to answer:

- Which books did I rate 90 or higher?
- Which technical books did I read in 2024?
- Which books did I read from a specific publisher or imprint?

Shelvia is built to keep the convenience of plain text while adding database-like search when you need it.

## Features

- Store one book as one TOML file.
- Keep reading data in your own repository.
- Create the initial directory and vocab files with `init`.
- Create a book file template with `new`.
- Validate required fields, ratings, dates, genres, publishers, and other controlled values.
- Search with SQL-like conditions.
- Treat TOML as the source of truth and use SQLite only as a temporary search view.

## Installation

Installation will be added before the first release.

## Quick Start

Create a directory for your reading data. Shelvia calls this directory the `shelf root`.

```bash
shelvia init ./my-shelf
```

`init` creates `books/`, `vocab/`, and empty vocab files.

```text
my-shelf/
  books/
    2024/
      Some Book.toml
  vocab/
    genres.toml
    publishers.toml
    editions.toml
    imprints.toml
```

Set the `shelf root` in an environment variable.

```bash
export SHELVIA_DIR=./my-shelf
```

Create a book file template.

```bash
shelvia new "Some Book" --read-date 2024-01-01
```

`new` creates `books/2024/Some Book.toml`. Open the generated file and fill in the book details.

```toml
title = "Some Book"
author = "Some Author"
rating = 90
read_date = 2024-01-01

genre = "Novel"
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"

[thoughts]
summary = "A short note."
body = """
Longer thoughts can live here.
"""
```

Add allowed values to the vocab files under `vocab/`. Vocab files define allowed values for genres, publishers, editions, imprints, and similar fields.

```toml
# vocab/genres.toml
values = [
  "Novel",
]
```

```toml
# vocab/publishers.toml
values = [
  "Example Publisher",
]
```

```toml
# vocab/editions.toml
[[values]]
name = "Paperback"
imprint_required = true
```

```toml
# vocab/imprints.toml
values = [
  "Example Paperback",
]
```

Validate the data.

```bash
shelvia validate
```

List your books.

```bash
shelvia list
```

Search with conditions.

```bash
shelvia query --where 'rating >= 90'
shelvia query --where 'genre = "Novel" and publisher = "Example Publisher"'
```

## Data Format

Book files are written in TOML. Required fields are:

- `title`
- `author`
- `rating`
- `read_date`
- `genre`
- `publisher`

Optional fields are:

- `edition`
- `imprint`
- `series`
- `translator`
- `thoughts.summary`
- `thoughts.body`

`rating` is an integer from `0` to `100`. `read_date` is written as a TOML date. `genre`, `publisher`, `edition`, and `imprint` are checked against vocab files.

Optional fields can be omitted. Empty strings in optional fields are treated the same as omitted fields.

## Vocab

Vocab files prevent inconsistent names for genres, publishers, editions, and imprints. Shelvia reads these files from `vocab/` under the `shelf root`:

- `genres.toml`
- `publishers.toml`
- `editions.toml`
- `imprints.toml`

Genres are user-managed vocab values too. If a vocab type is not used yet, create the file anyway and leave `values` empty.

```toml
# vocab/genres.toml
values = [
  "Novel",
  "Essay",
  "Technical",
  "Business",
]
```

```toml
# vocab/publishers.toml
values = [
  "Example Publisher",
  "Another Publisher",
]
```

```toml
# vocab/editions.toml
[[values]]
name = "Paperback"
imprint_required = true

[[values]]
name = "Hardcover"
imprint_required = false
```

```toml
# vocab/imprints.toml
values = [
  "Example Paperback",
]
```

When an edition has `imprint_required = true`, omitting the imprint is an error. If an imprint is present, an edition must also be present.

## `shelf root`

Shelvia receives one `shelf root` and reads `books/` and `vocab/` under it.

```text
my-shelf/
  books/
    2024/
      Some Book.toml
  vocab/
    genres.toml
    publishers.toml
    editions.toml
    imprints.toml
```

Under `books/`, Shelvia recursively loads `.toml` files. Under `vocab/`, it reads `genres.toml`, `publishers.toml`, `editions.toml`, and `imprints.toml`.

Shelvia determines the `shelf root` in this order:

1. The directory passed to `--shelf`
2. The `SHELVIA_DIR` environment variable

When `--shelf` is passed, it takes precedence over `SHELVIA_DIR`. If neither is set, Shelvia exits with an error.

```bash
export SHELVIA_DIR=~/reading-log
shelvia validate
shelvia --shelf ~/other-reading-log validate
```

## Validation

`validate` only checks that the whole `shelf root` can be loaded. It reads book files and vocab, then validates required fields, types, ratings, dates, vocab references, and the relationship between editions and imprints. It does not create, update, or convert files.

```bash
shelvia validate
```

On success, Shelvia prints the number of loaded files.

```text
Validated 1 book, 4 vocab files.
```

On failure, Shelvia prints the file path, location, and reason.

```text
books/2024/Some Book.toml:5: unknown genre "Novel"
```

If `validate` succeeds, the same shelf can be loaded by `list` and `query`. It is intended as a pre-commit check after adding or editing reading data.
