# Uncluster agent guide

## Always-on response skill

Before returning any user-facing text for work in this repository—including progress updates, questions, errors, and final answers—read and apply `C:\Users\Bright\.codex\skills\i-have-adhd\SKILL.md`. Apply it on every turn so the next action, active state, and completed outcome remain easy to scan.

## Git safety and commit convention

- Treat editing permission as permission to change the working tree only. Create a commit only after the user explicitly asks for that commit, and push only after the user explicitly asks for that push; permission to commit does not include permission to push.
- Use Git's configured author identity because Git requires author metadata. Never override it or add `Co-authored-by`, `Signed-off-by`, AI/tool attribution, or other authorship trailers.
- Write every commit as one line: `vN.NN: <technical summary>. next: <next action>`.
- Choose the version from the most recent commit whose subject starts with `vN.NN`, then increment the two-digit suffix (`v1.01` -> `v1.02`; `v1.99` -> `v2.00`). If no versioned commit exists, start at `v1.01`.

## Purpose

Uncluster is a Go 1.21 HTML transformation tool with a Fiber API, a browser UI, and command-line entry points. It parses HTML into a DOM and can format it, extract inline and remote resources, scrape a public URL with its assets, generate React/Vite TSX or Express/EJS projects, and turn HTML or an archived site into a local bundle.

The current TSX intent is authoritative in `docs/intent/tsx-conversion.md`: conversion is deterministic and rule-based, structural fidelity comes before component decomposition, and no model belongs in the transformation path. Dated files under `docs/plans/` explain past or proposed work; confirm their status against current code and tests.

## Entry points and flows

- `main.go`: Fiber server, `/api/*` handlers, 50 MiB request limit, health endpoint, and static `dist/` UI. The default port is `3000`, overridden by `PORT`.
- `cmd/uncluster/main.go`: primary CLI with `format`, `analyze`, `split`, `nodejs`, `nodejs-ejs`, and `bundle` modes.
- `cmd/uncluster-split/main.go`: smaller extraction-only CLI that can emit `split-manifest.json`.
- Raw HTML flow: `extractor.Extract` -> shared `htmlutil.InlineCollector` plus `fetcher` -> formatted rewritten HTML -> `zipper` or a `nodejs` generator.
- URL flow: `scraper.ScrapeURL` -> guarded HTTP fetches and asset localization -> `extractor.ExtractedContent` -> the same export paths.
- Bundle flow: `bundle.ProcessWithOptions` -> load HTML or safely extract ZIP -> choose `index.html` -> localize referenced assets -> write the original page, split output under `unzip/`, and an EJS project under `ejs/`.
- TSX flow: `nodejs.GenerateProject` -> `generateTSXViews` -> shared decomposition helpers in `internal/nodejs/ejs_builder.go` -> `converter.ConvertSectionToTSX`.
- EJS flow: `nodejs.GenerateEJSProject` -> `generateEJSViews`; it shares component selection and naming with TSX.

## Package map

- `internal/htmlutil`: shared DOM/attribute helpers and inline `<style>`/executable `<script>` extraction.
- `internal/formatter`: DOM-based HTML rendering and whitespace/void/raw-text handling.
- `internal/extractor`: coordinates inline extraction, external CSS/JS downloads, link rewriting, and target-specific path rewrites.
- `internal/safehttp`, `internal/fetcher`, `internal/scraper`: SSRF-safe HTTP, bounded downloads, remote resource naming, page scraping, and HTML/CSS asset rewriting.
- `internal/bundle`: HTML/ZIP intake, index selection, archive safety, recursive local-asset discovery, and bundle directory output.
- `internal/converter`: deterministic HTML-fragment-to-TSX translation, including React attribute, form-control, escaping, and whitespace rules.
- `internal/nodejs`: React/Vite and EJS project assembly, shared decomposition/naming, templates, and project ZIP creation.
- `internal/analyzer`: heuristic suggestions for repeated semantic UI patterns; this is separate from TSX conversion.
- `internal/zipper`: ZIP output for extracted HTML and resources.

## Invariants

- TSX conversion must preserve DOM nodes, attributes, text, order, and significant whitespace. Extend `internal/converter/testdata/` and `fidelity_test.go` for markup regressions.
- Component decomposition may move existing DOM subtrees into files but must preserve surrounding and nested markup. Keep TSX and EJS selection/naming behavior aligned through the shared helpers and `internal/nodejs/decomposition_test.go`.
- Inline extraction preserves non-resource attributes, leaves data scripts such as JSON-LD in place, and uses portable relative paths.
- All remote fetches accept only HTTP(S), block non-public destinations at connection time, revalidate redirects, and cap response bodies at 25 MiB. ZIP extraction rejects paths outside its temporary root and caps expanded content at 512 MiB.
- Local asset resolution stays inside the source root, rewrites only discovered files, follows CSS references recursively, and preserves `srcset` descriptors.
- `qa/workbook/**` is generated. Edit `qa/source/quality-model.json`, then run `go run ./qa/tools/qualitydoc`.

## Build and test

```powershell
go build ./...
go test ./...
go run .
go run ./cmd/uncluster --help
```

On this Windows environment, use `powershell -File qa/run-tests.ps1` if Application Control blocks Go test binaries in the temporary directory. Add focused tests beside the changed package, then run the full suite. The opt-in real-site driver in `cmd/uncluster/zz_e2e_test.go` requires `E2E_ZIP` and `E2E_OUT`.

## Where to change things

- API routes, request validation, response headers: `main.go`
- CLI modes and filesystem output: `cmd/uncluster/main.go`
- Inline/external resource behavior: `internal/htmlutil/inline.go`, `internal/extractor`, `internal/fetcher`
- URL scraping or SSRF/download policy: `internal/scraper`, `internal/safehttp`
- TSX syntax or fidelity: `internal/converter`; add a minimal fixture when possible
- TSX/EJS component boundaries or names: `internal/nodejs/ejs_builder.go`, `tsx_builder.go`, and `decomposition_test.go`
- Generated project files: `internal/nodejs/templates.go` and `ejs_templates.go`
- ZIP/HTML bundle intake and local assets: `internal/bundle`
- Browser UI: `dist/`; keep its API calls synchronized with `setupRoutes` in `main.go`
