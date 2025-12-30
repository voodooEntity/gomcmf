# gomcmf

A small, self-contained static content/site builder written in Go. It provides a minimal project scaffold, a simple page model with ordered filenames, a lightweight HTML template, and a custom Markdown-like converter tuned for producing useful page HTML.

## Features
- Single binary CLI; no external runtime required
- Project scaffolding (init): pages, resources, output, and default templates
- Page creation (create) with incrementing numeric sequence (for example, `001-hello-world.md`)
- Site building (build) using a main HTML template and a simple content pipeline
- Custom Markdown-like converter: headings, links, images, lists, blockquotes, inline bold/italic, fenced code blocks
- Simple navigation generation based on page groups (by directory)

## Installation
Build from source with Go 1.20+ (recommended):

```
git clone https://github.com/voodooEntity/gomcmf.git
cd gomcmf
go build -o gomcmf ./cmd/cli
```

You can then run `./gomcmf` from your project directory.

## Quick start
1) Initialize a new project in an empty directory:

```
./gomcmf -command init
```

This creates:
- `index.md`, `404.md`, `main.html`, `config.json`
- `pages/`, `resources/`, `output/`

2) Create a page:

```
./gomcmf -command create -name "My First Post" -type md
```

3) Build the site:

```
./gomcmf -command build -target ./public
```

## Commands and flags

```
gomcmf -command <init|create|build> [flags]
gomcmf help
```

- init
  - Scaffolds default files and directories.

- create
  - Flags:
    - `-name string` (required): Page title used to form the filename
    - `-type string` (optional, default `md`): One of `md`, `html`, `link`
    - `-sequence int` (optional): Manually set sequence; otherwise next free is used

- build
  - Flags:
    - `-target string` (optional, default `./`): Target directory to build into

Global flags:
- `-verbose` Enable verbose logging
- `-help`    Show flag help from Go's `flag` package

You can also run `gomcmf help` to print a concise usage guide.

## Configuration (config.json)
These keys are used during build. All paths are relative to your current working directory unless absolute.

- `base`            Base URL for links, for example `/` or `https://example.com/`
- `title`           Site title
- `pagesPath`       Source pages directory (for example, `pages`)
- `resourcesPath`   Static assets directory (for example, `resources`)
- `buildPath`       Output directory (for example, `output`)
- `mainFile`        Main HTML template file (for example, `main.html`)
- `indexFile`       Index markdown file (for example, `index.md`)
- `404File`         Not-Found markdown file (for example, `404.md`)

## Content authoring (Markdown-like)
The converter is intentionally minimal and tailored for page HTML. Supported elements:

- Headings: `# H1`, `## H2`, ... (`###### H6`)
- Links: `[text](url)` and `[text](url "title")`
- Images: `![alt](src)` and `![alt](src "title")`
- Bold: `**strong**` or `__strong__`
- Italic: `*em*` or `_em_`
- Unordered lists: lines starting with `- `
- Blockquotes: lines starting with `> `
- Fenced code blocks: triple backticks ``` with optional language, for example ```go

Notes:
- Inline formatting (bold/italic) is applied to text, not inside HTML tags or attributes. This prevents links from breaking when URLs contain underscores.
- Pages can be of type `md`, `html`, or `link`. The `link` type is treated as a navigation entry and not rendered to its own HTML file.

## Filenames and ordering
Pages are stored with a numeric sequence prefix to determine order, for example:

```
pages/
  001-hello-world.md
  010-another-post.md
```

`gomcmf create` will pick the next available sequence automatically unless `-sequence` is provided.

## Templates
- `main.html` is the base template. Page content and other blocks are injected by the build step.
- The default template included with `init` is a simple starter; you can customize it to your needs.

## Project layout
After `init`:

```
.
├── config.json
├── main.html
├── index.md
├── 404.md
├── pages/
├── resources/
└── output/          # or a custom buildPath
```

## Build output
`buildPath` (default `output/`) will contain the generated site, including copied resources and rendered pages (`.html`).

## Exit codes and errors
At present, some errors cause the process to exit immediately. Non-zero exit codes on failure are recommended, but parts of the current code may still exit 0 on error.

## License
See [LICENSE](LICENSE).
