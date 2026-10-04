# Uncluster usage guide

Uncluster is a deterministic HTML transformation toolkit written in Go. It accepts raw HTML, local HTML/ZIP inputs, or public HTTP(S) pages; separates page resources; and generates editable React/Vite/TypeScript or Express/EJS projects.

The repository includes three entry points:

- A Fiber API and browser UI served by `main.go`.
- The primary multi-mode CLI in `cmd/uncluster`.
- A smaller extraction-only CLI in `cmd/uncluster-split`.

## Capabilities

| Operation | Input | Output |
| --- | --- | --- |
| Format | HTML | Normalized, indented HTML |
| Split | HTML | Rewritten HTML plus extracted inline/external CSS and JavaScript |
| Scrape | Public HTTP(S) URL | Localized page archive with CSS, JavaScript, images, fonts, and other referenced assets |
| React project | HTML or scraped page | Express + Vite + React 18 + TypeScript project with TSX components |
| EJS project | HTML or scraped page | Express + EJS project with reusable partials |
| Bundle | Local HTML or ZIP | Original page, split output, and an EJS project in one directory |
| Analyze | HTML | JSON suggestions for repeated semantic UI patterns |

TSX generation is rule-based and reproducible. No model or LLM participates in parsing, conversion, component selection, or project generation. The converter prioritizes DOM fidelity over inferred abstractions: it preserves nodes, attributes, source order, text, and significant whitespace while translating HTML syntax to React-valid TSX.

## Requirements

- Go 1.21 or newer.
- Node.js 18 or newer to run generated Vite projects.

## Quick Start

```bash
git clone https://github.com/omariomari2/uncluster.git
cd uncluster
go run .
```

The browser UI is available at `http://localhost:3000`. Verify the API with:

```bash
curl http://localhost:3000/api/health
```

The server uses `PORT` when set and otherwise listens on port `3000`.

## Primary CLI

Run the CLI directly:

```bash
go run ./cmd/uncluster --help
go run ./cmd/uncluster ./page.html -to split -out ./split-output
```

Or build it:

```bash
go build -o ./bin/uncluster ./cmd/uncluster
```

### Syntax

```text
uncluster <input> -to <format> [-out <dir>] [-dest <dir>]
```

| Format | Accepted input | Default output | Behavior |
| --- | --- | --- | --- |
| `format` | HTML file | Standard output | Writes formatted HTML; with `-out`, writes `<dir>/index.html` |
| `analyze` | HTML file | Standard output | Writes suggestion JSON; with `-out`, writes `<dir>/components.json` |
| `split` | HTML file | `./split-output` | Extracts inline and remote CSS/JS and writes `split-manifest.json` |
| `nodejs` | HTML file | `./nodejs-project` | Generates a React/Vite/TypeScript project |
| `nodejs-ejs` | HTML file | `./nodejs-ejs-project` | Generates an Express/EJS project |
| `bundle` | HTML or ZIP file | `./bundle-output/<site>` | Writes the original page, split resources, and an EJS project |

`-dest` is valid only for `bundle` and selects the exact final directory. Without it, bundle mode derives a site name and creates that directory under `-out`.

Examples:

```bash
uncluster ./page.html -to format
uncluster ./page.html -to analyze -out ./analysis
uncluster ./page.html -to nodejs -out ./react-site
uncluster ./page.html -to nodejs-ejs -out ./ejs-site
uncluster ./capture.zip -to bundle -out ./sites
uncluster ./capture.zip -to bundle -dest ./sites/example.com
```

The primary CLI processes local files. URL capture is currently exposed through the API and browser UI.

For raw HTML, extraction downloads only literal absolute HTTP(S) stylesheet and script URLs. It does not infer a base URL for relative images or other assets; use URL scraping or bundle mode when those files must be localized.

### Extraction-only CLI

`cmd/uncluster-split` provides a narrow interface for splitting one HTML file:

```bash
go run ./cmd/uncluster-split \
  -input ./page.html \
  -output ./split-output \
  -manifest=true
```

Set `-manifest=false` to omit `split-manifest.json`.

## HTTP API

Start the service with `go run .`. JSON endpoints accept `Content-Type: application/json`; export endpoints return `application/zip` attachments.

| Method | Route | Request | Response |
| --- | --- | --- | --- |
| `GET` | `/api/health` | None | Service status and VCS revision |
| `POST` | `/api/format` | `{"html":"..."}` | `{"success":true,"data":"..."}` |
| `POST` | `/api/analyze` | `{"html":"..."}` | Component suggestions as JSON |
| `POST` | `/api/export` | `{"html":"..."}` | Split resource ZIP |
| `POST` | `/api/export-nodejs` | `{"html":"..."}` | React/Vite/TypeScript project ZIP |
| `POST` | `/api/export-nodejs-ejs` | `{"html":"..."}` | Express/EJS project ZIP |
| `POST` | `/api/bundle-zip` | Multipart field `file` containing a ZIP | Bundle ZIP |
| `POST` | `/api/scrape` | `{"url":"https://..."}` | Localized split resource ZIP |
| `POST` | `/api/scrape-nodejs` | `{"url":"https://..."}` | Localized React project ZIP |
| `POST` | `/api/scrape-nodejs-ejs` | `{"url":"https://..."}` | Localized EJS project ZIP |

Capture a public page:

```bash
curl -sS -X POST http://localhost:3000/api/scrape \
  -H "Content-Type: application/json" \
  --data '{"url":"https://example.com"}' \
  --output example.zip
```

Process an existing capture ZIP as a bundle:

```bash
curl -sS -X POST http://localhost:3000/api/bundle-zip \
  -F "file=@example.zip" \
  --output bundle.zip
```

## Output Layouts

### Split output

The exact files depend on the source document and which downloads succeed:

```text
split-output/
|-- index.html
|-- inline/
|   |-- style-1.css
|   `-- script-1.js
|-- external/
|   |-- css/
|   `-- js/
|-- assets/                  # Scraped or bundled images, fonts, and other files
|-- split-manifest.json      # Local CLI output
`-- uncluster-capture.json   # API export archives
```

`index.html` is rewritten to reference the emitted resource paths. Executable inline scripts are extracted; data scripts such as JSON-LD remain in the document.

### React/Vite/TypeScript project

```text
project/
|-- package.json
|-- vite.config.js
|-- server.js
|-- tsconfig.json
|-- public/
|   |-- assets/
|   `-- scripts/
`-- src/
    |-- index.html
    |-- App.tsx
    |-- main.tsx
    |-- components/
    |   |-- MainComponent.tsx
    |   `-- <Section>.tsx
    `-- styles/
```

Run a generated React project with:

```bash
npm install
npm run dev
```

Vite listens on port `8080`. `npm run build` creates `dist/`, and `npm start` serves the production build through Express.

### Express/EJS project

```text
project/
|-- package.json
|-- server.js
|-- public/
|   |-- inline/
|   |-- external/
|   `-- assets/
`-- views/
    |-- index.ejs
    `-- partials/
```

Run a generated EJS project with `npm install && npm start`. The generated server listens on `PORT` or port `8080`.

### Bundle output

```text
<site>/
|-- index.html       # Selected source page, unchanged
|-- unzip/           # Rewritten split output and local assets
`-- ejs/             # Runnable Express/EJS project
```

For ZIP inputs, bundle mode scans for usable `index.html` files and chooses one deterministically by source-name match, depth, size, and path. Bundle mode does not currently create a TSX project; run `-to nodejs` against an HTML source when React output is required.

## Capture Manifest

API export and scrape archives include `uncluster-capture.json`:

```json
{
  "version": 1,
  "source_url": "https://example.com",
  "complete": true,
  "assets": [
    {
      "url": "https://example.com/site.css",
      "path": "external/css/site.css",
      "type": "css",
      "status": "localized"
    }
  ]
}
```

Asset status values are:

- `localized`: downloaded and written to the export.
- `retained-external`: intentionally left remote, such as an iframe or supported hosted stylesheet.
- `failed`: localization failed; the manifest sets `complete` to `false`.

`complete: true` means that every attempted localization succeeded. It does not guarantee that the result is fully offline because intentionally retained resources may still require network access.

## Processing Architecture

### Raw HTML

```text
HTML
  -> extractor.Extract
  -> htmlutil.InlineCollector + fetcher
  -> formatter.Format
  -> zipper or nodejs generator
```

### Public URL

```text
URL
  -> scraper.ScrapeURL
  -> safehttp validation and bounded fetches
  -> HTML/CSS asset discovery and path rewriting
  -> extractor.ExtractedContent
  -> split, TSX, or EJS export
```

### Local HTML or ZIP bundle

```text
HTML/ZIP
  -> bundle.ProcessWithOptions
  -> safe archive extraction and index selection
  -> recursive local-asset discovery
  -> index.html + unzip/ + ejs/
```

TSX and EJS generation share component boundary selection and naming. TSX markup translation is centralized in `internal/converter`, while `internal/nodejs` assembles runnable projects around the converted views.

## Safety and Limits

| Control | Current behavior |
| --- | --- |
| URL schemes | Only `http` and `https` are accepted |
| Network targets | Loopback, private, link-local, multicast, unspecified, and reserved destinations are blocked at connection time |
| Redirects | Every redirect is revalidated; at most 10 redirects are followed |
| Remote body size | Each fetched response is capped at 25 MiB |
| API request size | Fiber request bodies are capped at 50 MiB |
| ZIP expansion | Extracted content is capped at 512 MiB |
| ZIP paths | Entries resolving outside the temporary extraction root are rejected |
| Local assets | Resolution is confined to the source root; nested CSS references are discovered recursively |

The development server currently enables CORS for all origins. Restrict this before exposing the API on an untrusted network.

## Package Map

| Path | Responsibility |
| --- | --- |
| `main.go` | Fiber server, API handlers, health metadata, and static UI |
| `internal/htmlutil` | DOM/attribute helpers and inline resource collection |
| `internal/formatter` | DOM-based HTML rendering and whitespace handling |
| `internal/extractor` | Resource extraction, fetch coordination, link rewriting, and capture metadata |
| `internal/safehttp` | URL validation, SSRF-resistant dialing, redirects, and response limits |
| `internal/fetcher` | Remote text/binary downloads and stable resource naming |
| `internal/scraper` | Public page capture and HTML/CSS asset localization |
| `internal/converter` | Deterministic HTML-fragment-to-TSX conversion |
| `internal/nodejs` | React/Vite and EJS decomposition, templates, and ZIP generation |
| `internal/bundle` | HTML/ZIP intake, index selection, and local asset packaging |
| `internal/analyzer` | Heuristic component suggestions, independent of TSX conversion |
| `internal/zipper` | Split archive creation |


