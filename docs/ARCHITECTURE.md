# Architecture

This document describes how Uncluster processes HTML, public URLs, and local ZIP captures. It reflects the packages under `internal/` and the three entry points in this repository.

## Entry points

| Entry | Role |
| --- | --- |
| `main.go` | Fiber API and the static browser UI in `dist/`. The server listens on `PORT`, or port `3000` if `PORT` is unset. Request bodies are capped at 50 MiB. CORS is currently open to all origins. |
| `cmd/uncluster` | Multi-mode CLI for local files: `format`, `analyze`, `split`, `nodejs`, `nodejs-ejs`, and `bundle`. |
| `cmd/uncluster-split` | Narrow CLI that runs `extractor.Extract` on one HTML file and writes the split tree, optionally with `split-manifest.json`. |

The primary CLI does not fetch URLs. URL capture is exposed by the API routes `/api/scrape`, `/api/scrape-nodejs`, and `/api/scrape-nodejs-ejs`, and by the browser UI.

## Pipelines

### Raw HTML

Used by `format`, `analyze`, `split`, `nodejs`, `nodejs-ejs`, and the corresponding `/api/format`, `/api/analyze`, `/api/export`, `/api/export-nodejs`, and `/api/export-nodejs-ejs` handlers.

```text
HTML
  -> html.Parse
  -> htmlutil.InlineCollector (inline <style> and executable <script> bodies)
  -> extractor finds literal absolute http(s) stylesheet and script URLs
  -> fetcher downloads those URLs through safehttp
  -> extractor rewrites tags to local paths
  -> formatter.Format
  -> zipper, nodejs.GenerateProject, or nodejs.GenerateEJSProject
```

`extractor.Extract` is the shared step for split and project generation. Format and analyze do not go through the extractor: format calls `formatter.Format` directly, and analyze calls `analyzer.AnalyzeComponents`.

For raw HTML, extraction downloads only literal absolute HTTP(S) stylesheet and script URLs. It does not invent a base URL for relative images or other assets. Use URL scraping or bundle mode when those files must be localized. Data scripts such as JSON-LD stay in the document; executable inline scripts are extracted.

### Public URL

```text
URL
  -> safehttp.ValidateURL
  -> scraper.ScrapeURL fetches the page with a guarded client
  -> HTML and CSS references are discovered and rewritten to local paths
  -> fetcher downloads CSS, JS, and binary assets through safehttp
  -> htmlutil.InlineCollector extracts remaining inline style and script
  -> formatter.Format
  -> extractor.ExtractedContent
  -> zipper or nodejs generators
```

`scraper` records a capture manifest (`uncluster-capture.json` in API exports) with per-asset status: `localized`, `retained-external`, or `failed`. `complete` is true only when every attempted localization succeeded. Intentionally retained resources, such as an iframe or a Google Fonts stylesheet, may still require network access.

### Local HTML or ZIP bundle

```text
HTML or ZIP
  -> bundle.ProcessWithOptions
  -> ZIP entries are extracted under a temporary root; paths outside that root are rejected
  -> extracted content is capped at 512 MiB
  -> an index.html is chosen by source-name match, depth, size, and path
  -> local assets are discovered recursively, including nested CSS url() and @import references
  -> resolution stays inside the source root
  -> output: index.html (selected source, unchanged), unzip/ (split output and assets), ejs/ (Express + EJS project)
```

Bundle mode does not currently create a TSX project. Use `-to nodejs` on an HTML source for React output.

## Package responsibilities

| Package | Responsibility |
| --- | --- |
| `internal/htmlutil` | DOM and attribute helpers, plus `InlineCollector`, which moves inline `<style>` and `<script>` bodies into numbered files and leaves a reference element in the tree. |
| `internal/formatter` | DOM-based HTML rendering with indentation and whitespace rules. |
| `internal/extractor` | Coordinates inline collection, remote CSS/JS fetches, link rewriting, and capture metadata. `ExtractedContent` is the common intermediate value passed to zippers and project generators. |
| `internal/safehttp` | Accepts only `http`/`https` URLs with a host; dials through a hook that blocks loopback, private, link-local, multicast, unspecified, and reserved addresses; revalidates each redirect (at most 10); and caps each response body at 25 MiB. |
| `internal/fetcher` | Downloads remote text and binary resources with `safehttp`, and assigns stable, collision-resistant filenames. |
| `internal/scraper` | Fetches a public page, localizes HTML and CSS assets, and returns `extractor.ExtractedContent`. |
| `internal/converter` | Deterministic HTML-fragment-to-TSX translation (`ConvertSectionToTSX`). Attribute names, boolean attributes, inline styles, and character escaping are handled with fixed tables and rules. |
| `internal/nodejs` | Builds a runnable React/Vite/TypeScript project or an Express/EJS project around extracted resources. TSX markup comes from `converter`. Component boundaries and names are shared between the two project types. ZIP packing for generated projects lives here as well. |
| `internal/bundle` | Loads a local HTML file or ZIP, selects an index, rewrites local asset paths, and writes the bundle layout. |
| `internal/analyzer` | Heuristic suggestions for repeated semantic UI patterns, emitted as JSON. This path is independent of TSX conversion. |
| `internal/zipper` | Packs split HTML, inline and external resources, local assets, and an optional capture manifest into a ZIP for API export routes. |

## API surface

`setupRoutes` in `main.go` mounts:

| Method | Route | Result |
| --- | --- | --- |
| `GET` | `/api/health` | Service status and VCS revision |
| `POST` | `/api/format` | Formatted HTML |
| `POST` | `/api/analyze` | Component suggestion JSON |
| `POST` | `/api/export` | Split resource ZIP |
| `POST` | `/api/export-nodejs` | React/Vite/TypeScript project ZIP |
| `POST` | `/api/export-nodejs-ejs` | Express/EJS project ZIP |
| `POST` | `/api/bundle-zip` | Bundle ZIP from an uploaded ZIP |
| `POST` | `/api/scrape` | Localized split ZIP from a URL |
| `POST` | `/api/scrape-nodejs` | Localized React project ZIP from a URL |
| `POST` | `/api/scrape-nodejs-ejs` | Localized EJS project ZIP from a URL |

Static files from `dist/` are served at `/`.

## Safety notes

Remote fetches are constrained by `safehttp` as described above. ZIP extraction in `bundle` rejects paths that escape the temporary root and caps expanded size at 512 MiB. These controls reduce accidental local-file or SSRF exposure; they do not authorize copying a third-party site.

The development server enables CORS for all origins. Restrict that before exposing the API on an untrusted network.

## Related documentation

- [Usage guide](usage.md) for CLI flags, API examples, and output layouts
- [Contributing](../CONTRIBUTING.md) for build and test commands
- [Security policy](../SECURITY.md) for private vulnerability reporting
