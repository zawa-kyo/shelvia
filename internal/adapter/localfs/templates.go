package localfs

const configTemplate = `kind = "shelvia-config"

[values]
genres = [
  "Novel",
]

publishers = [
  "Example Publisher",
]

imprints = [
  "Example Paperback",
]

[[values.editions]]
name = "Paperback"
imprint_required = true

[[values.editions]]
name = "Hardcover"
imprint_required = false
`

const exampleTemplate = `title = "Example Book"
author = "Example Author"
rating = 80
read_date = 2024-01-01

genre = "Novel"
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"
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

genre = ""
edition = ""
imprint = ""
publisher = ""
series = ""
translator = ""

[thoughts]
summary = ""
body = """
"""
`
