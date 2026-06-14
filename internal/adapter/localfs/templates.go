package localfs

const configTemplate = `# List the values you want to use in this shelf.
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
`

const exampleTemplate = `# example.toml is real shelf data and is included in validation.
title = "Example Book"
author = "Example Author"
rating = 80
read_date = 2024-01-01

genre = "Novel"
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"
# Optional fields can be omitted. Empty strings are treated as unset.
series = ""
translator = ""

[thoughts]
summary = "A short note."
body = """
Longer thoughts can live here.
"""
`

const bookTemplate = `title = "%s"
author = ""
rating = 0
read_date = %s

# Use the values listed in config.toml.
genre = ""
edition = ""
imprint = ""
publisher = ""
# Optional fields can stay empty.
series = ""
translator = ""

[thoughts]
summary = ""
body = """
"""
`
