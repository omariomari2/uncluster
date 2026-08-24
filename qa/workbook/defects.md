# Defects

> Generated from [`qa/source/quality-model.json`](../source/quality-model.json)
> by `go run ./qa/tools/qualitydoc`. Do not edit by hand.
> [Back to the coverage summary](README.md).

## Index

| Defect ID | Severity | Status | Feature | Title | Test case | Last tested |
|---|---|---|---|---|---|---|
| [DEF-001](#def-001) | Medium | Fixed | [FEAT-UI-016](features.md#feat-ui-016) | Scrape mode exposes a Format HTML action that cannot use the scraped URL | TC-FEAT-UI-016-009 | 2026-08-23 |
| [DEF-002](#def-002) | Medium | Fixed | [FEAT-UI-003](features.md#feat-ui-003) | Dropdown triggers are skipped by keyboard navigation | TC-FEAT-UI-003-009 | 2026-08-23 |
| [DEF-003](#def-003) | Medium | Fixed | [FEAT-UI-010](features.md#feat-ui-010) | Toast feedback is not announced to assistive technology | TC-FEAT-UI-010-009 | 2026-08-23 |
| [DEF-004](#def-004) | Medium | Fixed | [FEAT-UI-017](features.md#feat-ui-017) | Action links wrap native buttons and create duplicate tab stops | TC-FEAT-UI-017-009 | 2026-08-23 |
| [DEF-005](#def-005) | Medium | Fixed | [FEAT-UI-017](features.md#feat-ui-017) | Scrape button causes horizontal overflow at 320px | TC-FEAT-UI-017-010 | 2026-08-23 |
| [DEF-006](#def-006) | Critical | Fixed | [FEAT-CORE-005](features.md#feat-core-005) | Inline style conversion emits invalid JSX object syntax | TC-FEAT-CORE-005-007 | 2026-08-23 |
| [DEF-007](#def-007) | High | Fixed | [FEAT-CORE-005](features.md#feat-core-005) | Presence-only boolean HTML attributes become false | TC-FEAT-CORE-005-008 | 2026-08-23 |
| [DEF-008](#def-008) | High | Fixed | [FEAT-CORE-005](features.md#feat-core-005) | JSX text and attribute serialization does not escape JSX syntax | TC-FEAT-CORE-005-009 | 2026-08-23 |
| [DEF-009](#def-009) | High | Fixed | [FEAT-CORE-005](features.md#feat-core-005) | JSX conversion removes spaces around inline elements | TC-FEAT-CORE-005-010 | 2026-08-23 |
| [DEF-010](#def-010) | High | Fixed | [FEAT-CORE-004](features.md#feat-core-004) | Scraped srcset assets are downloaded but srcset is never rewritten | TC-FEAT-CORE-004-008 | 2026-08-23 |
| [DEF-011](#def-011) | High | Fixed | [FEAT-CORE-004](features.md#feat-core-004) | Scraper extracts JSON-LD and other data scripts as executable JavaScript | TC-FEAT-CORE-004-009 | 2026-08-23 |
| [DEF-012](#def-012) | High | Fixed | [FEAT-CORE-004](features.md#feat-core-004) | Plain scraped ZIP rewrites resources to absolute-root paths | TC-FEAT-CORE-004-007 | 2026-08-23 |
| [DEF-013](#def-013) | Medium | Fixed | [FEAT-UI-002](features.md#feat-ui-002) | localStorage exception can stop UI initialization | TC-FEAT-UI-002-009 | 2026-08-23 |
| [DEF-014](#def-014) | Low | Fixed | [FEAT-UI-016](features.md#feat-ui-016) | Blank scrape download names do not use server defaults | TC-FEAT-UI-016-010 | 2026-08-23 |
| [DEF-015](#def-015) | High | Fixed | [FEAT-CORE-004](features.md#feat-core-004) | Downloaded CSS dependencies are not rewritten to local assets | TC-FEAT-CORE-004-011 | 2026-08-23 |
| [DEF-016](#def-016) | High | Fixed | [FEAT-CORE-008](features.md#feat-core-008) | Generated TSX writes JavaScript files without loading them | TC-FEAT-CORE-008-007 | 2026-08-23 |
| [DEF-017](#def-017) | High | Fixed | [FEAT-CORE-008](features.md#feat-core-008) | TSX section generation can omit page content and collapse duplicates | TC-FEAT-CORE-008-008 | 2026-08-23 |
| [DEF-018](#def-018) | Medium | Fixed | [FEAT-CORE-003](features.md#feat-core-003) | Binary fetch creates a new idle transport per asset | TC-FEAT-CORE-003-007 | 2026-08-23 |
| [DEF-019](#def-019) | Medium | Fixed | [FEAT-UI-004](features.md#feat-ui-004) | Technical API documentation is incomplete and stale | TC-FEAT-UI-004-009 | 2026-08-23 |
| [DEF-020](#def-020) | High | Fixed | [FEAT-CORE-008](features.md#feat-core-008) | Generated project stylesheets retain standalone archive asset paths | TC-FEAT-CORE-008-009 | 2026-08-23 |

## Records

<a id="def-001"></a>

### DEF-001 — Scrape mode exposes a Format HTML action that cannot use the scraped URL

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-016](features.md#feat-ui-016) URL scrape workflows &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Switch to Scrape URL mode.
2. Enter a nonblank URL and click Scrape.
3. Click the visible Format HTML action.

**Expected result**

Every action revealed for scrape mode operates on the entered URL or is hidden.

**Actual result**

Before fix: Format HTML was visible and clicking it reported Please upload an HTML file first. After fix: computed visibility=hidden, opacity=0, pointer-events=none.

**Root cause hypothesis**

The shared action visibility helper was used for upload and scrape modes while the format handler has no scrape branch.

**Root cause**

showActionButtons added button-visible to the Format wrapper for both modes.

**Fix**

Toggle Format visibility only when scrapeMode is false.

**Verification**

qa/browser/ui-regression.html and direct Chrome smoke passed.

**Regression test:** `TC-FEAT-UI-016-009`

---

<a id="def-002"></a>

### DEF-002 — Dropdown triggers are skipped by keyboard navigation

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-003](features.md#feat-ui-003) How To Use dropdown &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Load the main screen.
2. Use Tab navigation.
3. Attempt to focus and operate How To Use or Downloads.

**Expected result**

Both dropdowns are focusable and activate with native keyboard controls.

**Actual result**

Before fix both triggers were P elements with tabIndex -1; after fix both are BUTTON elements with tabIndex 0.

**Root cause hypothesis**

Pointer click handlers were attached to paragraph elements without keyboard semantics.

**Root cause**

Noninteractive paragraph elements were used as dropdown triggers.

**Fix**

Use type=button triggers and reset their visual button chrome.

**Verification**

Focused Chrome regression passed.

**Regression test:** `TC-FEAT-UI-003-009`

---

<a id="def-003"></a>

### DEF-003 — Toast feedback is not announced to assistive technology

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-010](features.md#feat-ui-010) Feedback and reset &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Trigger any informational or error toast.
2. Inspect the created toast semantics.
3. Observe the absence of role or aria-live.

**Expected result**

Dynamic feedback is exposed through a live region.

**Actual result**

Before fix role and aria-live were null; after fix informational toasts use status/polite and errors use alert/assertive.

**Root cause hypothesis**

showToast created a visual div only.

**Root cause**

No accessible live-region attributes were assigned at toast creation.

**Fix**

Assign role, aria-live, and aria-atomic based on toast type.

**Verification**

Focused Chrome regression passed.

**Regression test:** `TC-FEAT-UI-010-009`

---

<a id="def-004"></a>

### DEF-004 — Action links wrap native buttons and create duplicate tab stops

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-017](features.md#feat-ui-017) Responsive and accessible use &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Load the main screen.
2. Inspect the action wrappers or traverse them with Tab.
3. Count anchor elements that directly contain buttons.

**Expected result**

Each action contributes exactly one valid interactive control.

**Actual result**

Before fix seven links directly wrapped buttons; after fix the nested interactive count is 0.

**Root cause hypothesis**

Decorative positioning wrappers were implemented as anchors even though the inner button owns the action.

**Root cause**

Invalid nested interactive markup created two tab stops per visible action.

**Fix**

Use noninteractive div wrappers while preserving the native buttons.

**Verification**

Focused Chrome regression and direct DOM check passed.

**Regression test:** `TC-FEAT-UI-017-009`

---

<a id="def-005"></a>

### DEF-005 — Scrape button causes horizontal overflow at 320px

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-017](features.md#feat-ui-017) Responsive and accessible use &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Render the main screen at 320px width.
2. Switch to scrape mode.
3. Compare document scrollWidth with innerWidth and inspect the Scrape button bounds.

**Expected result**

The scrape controls fit without horizontal overflow.

**Actual result**

Before fix the inner Scrape button remained 160px wide and produced scrollWidth 344 at innerWidth 320; after fix both widths are 320.

**Root cause hypothesis**

The global button selector had higher specificity than the 100px Scrape button rule.

**Root cause**

A 160px child button overflowed its 100px wrapper by 60px.

**Fix**

Match the global selector specificity for the 100px Scrape button override.

**Verification**

Focused 320px Chrome regression passed.

**Regression test:** `TC-FEAT-UI-017-010`

---

<a id="def-006"></a>

### DEF-006 — Inline style conversion emits invalid JSX object syntax

**Severity:** Critical &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-005](features.md#feat-core-005) JSX conversion &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Create the minimal fixture described by the static evidence.
2. Exercise the mapped public feature.
3. Compare output with the feature contract.

**Expected result**

Feature contract remains correct and complete.

**Actual result**

Before fix the Go test observed style={color: 'red', fontSize: '12px'}; after fix it observed the required nested object expression and passed.

**Root cause hypothesis**

convertStyleToObject returned only the JSX expression braces and omitted the JavaScript object braces.

**Root cause**

The style serializer emitted one brace pair instead of two.

**Fix**

Wrap the joined declarations in double braces.

**Verification**

Converter WebAssembly suite passed.

**Regression test:** `TC-FEAT-CORE-005-007`

---

<a id="def-007"></a>

### DEF-007 — Presence-only boolean HTML attributes become false

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-005](features.md#feat-core-005) JSX conversion &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Create the minimal fixture described by the static evidence.
2. Exercise the mapped public feature.
3. Compare output with the feature contract.

**Expected result**

Feature contract remains correct and complete.

**Actual result**

Before fix the Go test observed disabled={false} and checked={false}; after fix both were true and the test passed.

**Root cause hypothesis**

The converter interpreted the parsed string value instead of HTML boolean presence.

**Root cause**

x/net/html represents a presence-only boolean with an empty value.

**Fix**

Emit true whenever checked, disabled, or selected is present.

**Verification**

Converter WebAssembly suite passed.

**Regression test:** `TC-FEAT-CORE-005-008`

---

<a id="def-008"></a>

### DEF-008 — JSX text and attribute serialization does not escape JSX syntax

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-005](features.md#feat-core-005) JSX conversion &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Create the minimal fixture described by the static evidence.
2. Exercise the mapped public feature.
3. Compare output with the feature contract.

**Expected result**

Feature contract remains correct and complete.

**Actual result**

Before fix decoded quotes, ampersands, and braces were written raw; after fix both attribute and text assertions passed with entity encoding.

**Root cause hypothesis**

The HTML parser decoded entities, but the JSX serializer did not encode for its output context.

**Root cause**

Generic attribute and text branches wrote parser values directly.

**Fix**

Add separate attribute/text escapers and apply them only at serialization boundaries.

**Verification**

Converter WebAssembly suite passed.

**Regression test:** `TC-FEAT-CORE-005-009`

---

<a id="def-009"></a>

### DEF-009 — JSX conversion removes spaces around inline elements

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-005](features.md#feat-core-005) JSX conversion &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Convert <p>Hello <strong>world</strong> again</p>.
2. Inspect the generated paragraph.
3. Compare the rendered word boundaries.

**Expected result**

The generated JSX preserves spaces before and after the strong element.

**Actual result**

Before fix the Go test observed <p>Hello<strong>world</strong>again</p>; after fix it observed the expected spaced output.

**Root cause hypothesis**

Every text node was independently passed through strings.TrimSpace.

**Root cause**

Independent trimming removed semantic boundary whitespace between sibling text and inline element nodes.

**Fix**

Use the existing inline whitespace normalizer that collapses runs while retaining leading/trailing boundaries.

**Verification**

Converter WebAssembly suite passed.

**Regression test:** `TC-FEAT-CORE-005-010`

---

<a id="def-010"></a>

### DEF-010 — Scraped srcset assets are downloaded but srcset is never rewritten

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-004](features.md#feat-core-004) Scraper localization &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Create the minimal fixture described by the static evidence.
2. Exercise the mapped public feature.
3. Compare output with the feature contract.

**Expected result**

Feature contract remains correct and complete.

**Actual result**

Before fix the Go test retained both remote srcset candidates; after fix both localized paths and descriptors passed.

**Root cause hypothesis**

Asset discovery parsed srcset but HTML rewriting handled only src and poster.

**Root cause**

rewriteHTMLPaths omitted the srcset attribute.

**Fix**

Rewrite each candidate URL while preserving its descriptor.

**Verification**

Scraper WebAssembly suite passed.

**Regression test:** `TC-FEAT-CORE-004-008`

---

<a id="def-011"></a>

### DEF-011 — Scraper extracts JSON-LD and other data scripts as executable JavaScript

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-004](features.md#feat-core-004) Scraper localization &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Create the minimal fixture described by the static evidence.
2. Exercise the mapped public feature.
3. Compare output with the feature contract.

**Expected result**

Feature contract remains correct and complete.

**Actual result**

Before fix the Go test extracted both JSON-LD and JavaScript; after fix only the executable script was extracted and JSON-LD remained.

**Root cause hypothesis**

The scraper checked only for absence of src and ignored script type.

**Root cause**

extractInlineResources lacked the executable MIME allowlist.

**Fix**

Apply the same JavaScript/module type allowlist used by the extractor.

**Verification**

Scraper WebAssembly suite passed.

**Regression test:** `TC-FEAT-CORE-004-009`

---

<a id="def-012"></a>

### DEF-012 — Plain scraped ZIP rewrites resources to absolute-root paths

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-004](features.md#feat-core-004) Scraper localization &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Scrape a public page containing a relative image or stylesheet.
2. Extract the returned plain scrape ZIP.
3. Open index.html directly from the extracted directory.

**Expected result**

Localized references resolve to files within the extracted archive.

**Actual result**

Before fix the Go tests observed /external, /assets, and /inline references; after fix all standalone archive references were relative.

**Root cause hypothesis**

The scraper serialized server-root paths before the output type was known.

**Root cause**

rewriteAttr and inline replacement nodes unconditionally prepended a slash.

**Fix**

Keep scraped paths relative and let NodeJS/EJS rewrite stages apply server-specific roots.

**Verification**

TC-FEAT-CORE-004-007 and TC-FEAT-CORE-004-010 passed in the scraper WebAssembly suite.

**Regression test:** `TC-FEAT-CORE-004-007`

---

<a id="def-013"></a>

### DEF-013 — localStorage exception can stop UI initialization

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-002](features.md#feat-ui-002) Theme persistence &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Load the application in a context where localStorage writes throw SecurityError.
2. Apply a theme or allow DOMContentLoaded to initialize it.
3. Observe whether remaining initialization can continue.

**Expected result**

The theme applies in memory and the UI remains usable when storage is unavailable.

**Actual result**

Before fix applyTheme propagated SecurityError; after fix the light theme applied and all six browser regressions passed.

**Root cause hypothesis**

applyTheme wrote storage without a guarded fallback.

**Root cause**

Theme initialization directly read and wrote localStorage even though access can throw in restricted contexts.

**Fix**

Guard theme reads and writes, normalize values, and continue applying the DOM theme when persistence is unavailable.

**Verification**

Red/green browser regression passed with a forced SecurityError.

**Regression test:** `TC-FEAT-UI-002-009`

---

<a id="def-014"></a>

### DEF-014 — Blank scrape download names do not use server defaults

**Severity:** Low &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-016](features.md#feat-ui-016) URL scrape workflows &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Return a successful scrape response with Content-Disposition filename example-capture.zip.
2. Leave the relevant custom download name blank.
3. Run the ZIP scrape workflow and inspect the downloader filename.

**Expected result**

A blank custom field falls back to the server-provided filename.

**Actual result**

Before fix the downloader received extracted.zip; after fix it received example-capture.zip.

**Root cause hypothesis**

scrapeAndExport parsed serverFilename but resolved blank inputs against a hard-coded default.

**Root cause**

The server filename was only used in a dead filename || serverFilename fallback because resolveDownloadName always returns a value.

**Fix**

Use serverFilename as resolveDownloadName's fallback and pass the resolved value directly.

**Verification**

Red/green browser regression passed; all seven UI harness rows are green.

**Regression test:** `TC-FEAT-UI-016-010`

---

<a id="def-015"></a>

### DEF-015 — Downloaded CSS dependencies are not rewritten to local assets

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-004](features.md#feat-core-004) Scraper localization &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Download a stylesheet containing a relative image URL and an absolute font URL.
2. Download both dependencies into the archive assets directory.
3. Inspect the saved stylesheet content.

**Expected result**

The saved stylesheet points to the localized asset paths relative to its external/css directory.

**Actual result**

Before fix remote url() values remained in the CSS; after fix they became ../../assets paths and data URLs remained intact.

**Root cause hypothesis**

CSS dependencies entered urlToLocal after discovery, but FetchedResource.Content was never rewritten.

**Root cause**

The scraper rewrote HTML attributes only and returned downloaded stylesheet content unchanged.

**Fix**

Rewrite each successful external stylesheet after binary localization, resolving url() values against the source stylesheet and emitting portable relative paths.

**Verification**

Focused red/green test and the complete native scraper package suite passed.

**Regression test:** `TC-FEAT-CORE-004-011`

---

<a id="def-016"></a>

### DEF-016 — Generated TSX writes JavaScript files without loading them

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-008](features.md#feat-core-008) TSX scaffolder &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Generate a TSX project from HTML with inline JavaScript and a successful external script.
2. Inspect generated files and src/main.tsx.
3. Confirm whether the original scripts can execute.

**Expected result**

The project includes both script contents and loads them after React mounts.

**Actual result**

Before fix inline JavaScript was discarded and external JavaScript was unreferenced; after fix both are public files loaded sequentially.

**Root cause hypothesis**

organizeSourceFiles wrote only external JS under src and generateMainTsx produced CSS imports only.

**Root cause**

The scaffolder had no generated runtime path from ProjectConfig.JS or ExternalJS to the browser.

**Fix**

Write scripts under public/scripts and generate a post-render sequential script loader in src/main.tsx.

**Verification**

Focused WebAssembly Go regression passed with both file and loader assertions.

**Regression test:** `TC-FEAT-CORE-008-007`

---

<a id="def-017"></a>

### DEF-017 — TSX section generation can omit page content and collapse duplicates

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-008](features.md#feat-core-008) TSX scaffolder &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Generate TSX from a root containing non-section siblings around two identical semantic sections.
2. Inspect generated component files and MainComponent.
3. Compare the rendered component sequence with all direct source nodes.

**Expected result**

Every direct content node is represented in order, including repeated content.

**Actual result**

Before fix intro/outro nodes were omitted and two sections collapsed to one; after fix four distinct component instances are generated.

**Root cause hypothesis**

The TSX builder selected only nested section boundaries, then reused component names by identical rendered HTML and deduplicated them again.

**Root cause**

Section discovery was not a complete partition of the source root, and content-based naming erased duplicate occurrences.

**Fix**

Partition by direct content children of the selected root and generate a unique name for every occurrence.

**Verification**

Focused WebAssembly Go regression passed alongside the JavaScript-loading regression.

**Regression test:** `TC-FEAT-CORE-008-008`

---

<a id="def-018"></a>

### DEF-018 — Binary fetch creates a new idle transport per asset

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-003](features.md#feat-core-003) Guarded fetching &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Create two guarded HTTP clients as repeated asset fetches do.
2. Inspect their Transport identities.
3. Confirm whether idle connection pools can be reused.

**Expected result**

Guarded clients share a bounded, SSRF-protected transport while retaining their own total timeouts.

**Actual result**

Before fix every client had a distinct idle pool; after fix both clients share one transport and remain distinct instances.

**Root cause hypothesis**

safehttp.Client constructed a new dialer and http.Transport on every call.

**Root cause**

Transport lifecycle was coupled to the short-lived client wrapper instead of process-level guarded connection pooling.

**Fix**

Create one guarded transport with bounded idle settings and reuse it across per-timeout clients.

**Verification**

Focused WebAssembly Go regression and the complete safehttp suite passed.

**Regression test:** `TC-FEAT-CORE-003-007`

---

<a id="def-019"></a>

### DEF-019 — Technical API documentation is incomplete and stale

**Severity:** Medium &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-UI-004](features.md#feat-ui-004) Technical overview screen &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Open the technical API reference.
2. Compare its endpoint table and payload summary with main.go route registration and handlers.
3. Check scrape, bundle, and health entries.

**Expected result**

All 11 routes and their real request/response families are documented.

**Actual result**

Before fix only 7 routes appeared, all requests were described as html JSON, and health claimed version; after fix the browser check found all 11 current entries.

**Root cause hypothesis**

The static reference was not updated when scrape and bundle routes and revision metadata were added.

**Root cause**

Documentation duplicated route contracts manually without a matching regression check.

**Fix**

Add the four missing routes and describe html JSON, url JSON, multipart upload, ZIP responses, and health revision accurately.

**Verification**

Red/green browser documentation regression passed with all eight UI harness rows green.

**Regression test:** `TC-FEAT-UI-004-009`

---

<a id="def-020"></a>

### DEF-020 — Generated project stylesheets retain standalone archive asset paths

**Severity:** High &nbsp;·&nbsp; **Status:** Fixed &nbsp;·&nbsp; **Feature:** [FEAT-CORE-008](features.md#feat-core-008) TSX scaffolder &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Reproduction steps**

1. Localize a scraped stylesheet dependency to ../../assets/hero.png for the standalone archive.
2. Generate TSX and EJS projects from the same extracted content.
3. Inspect the stylesheet paths relative to the projects' public asset directories.

**Expected result**

Generated projects reference /assets/hero.png while the standalone archive keeps its portable relative path.

**Actual result**

Before fix both generated stylesheets retained ../../assets/hero.png, which resolves outside public assets; after fix both use /assets/hero.png.

**Root cause hypothesis**

The scraper's output-neutral relative CSS representation was copied verbatim into output formats with different asset roots.

**Root cause**

TSX and EJS builders did not adapt localized stylesheet paths to their public directory contract.

**Fix**

Translate standalone ../../assets references to /assets while writing external CSS into generated public-serving projects.

**Verification**

Focused red/green WebAssembly regression passed for both TSX and EJS outputs.

**Regression test:** `TC-FEAT-CORE-008-009`

---

