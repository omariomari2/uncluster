# Features

> Generated from [`qa/source/quality-model.json`](../source/quality-model.json)
> by `go run ./qa/tools/qualitydoc`. Do not edit by hand.
> [Back to the coverage summary](README.md).

## Index

| Feature ID | Feature name | Current status | Test cases | Defects | Severity | Last tested |
|---|---|---|---:|---:|---|---|
| [FEAT-API-001](#feat-api-001) | POST /api/format | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-002](#feat-api-002) | POST /api/convert | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-003](#feat-api-003) | POST /api/analyze | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-004](#feat-api-004) | POST /api/export | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-005](#feat-api-005) | POST /api/export-nodejs | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-006](#feat-api-006) | POST /api/export-nodejs-ejs | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-007](#feat-api-007) | POST /api/scrape | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-008](#feat-api-008) | POST /api/scrape-nodejs | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-009](#feat-api-009) | POST /api/scrape-nodejs-ejs | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-010](#feat-api-010) | POST /api/bundle-zip | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-API-011](#feat-api-011) | GET /api/health | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CFG-001](#feat-cfg-001) | API PORT | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CFG-002](#feat-cfg-002) | HTTP middleware and body/CORS policy | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CFG-003](#feat-cfg-003) | Safety limits | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CFG-004](#feat-cfg-004) | Build/install/deploy commands | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CFG-005](#feat-cfg-005) | Generated project runtime | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-001](#feat-cli-001) | CLI parsing and help | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-002](#feat-cli-002) | CLI format | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-003](#feat-cli-003) | CLI JSX | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-004](#feat-cli-004) | CLI analyze | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-005](#feat-cli-005) | CLI split | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-006](#feat-cli-006) | CLI TSX project | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-007](#feat-cli-007) | CLI EJS project | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-008](#feat-cli-008) | CLI bundle | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CLI-009](#feat-cli-009) | uncluster-split utility | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-001](#feat-core-001) | Formatter engine | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-002](#feat-core-002) | Resource extraction | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-003](#feat-core-003) | Guarded fetching | Partially Tested - Transport Reuse Fixed | 7 | 1 | Medium fixed | 2026-08-23 |
| [FEAT-CORE-004](#feat-core-004) | Scraper localization | Partially Tested - Four Defects Fixed | 11 | 4 | High fixed | 2026-08-23 |
| [FEAT-CORE-005](#feat-core-005) | JSX conversion | Partially Tested - Defects Fixed | 10 | 4 | Critical fixed | 2026-08-23 |
| [FEAT-CORE-006](#feat-core-006) | Section/list TSX conversion | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-007](#feat-core-007) | Component analyzer | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-008](#feat-core-008) | TSX scaffolder | Partially Tested - Three Defects Fixed | 9 | 3 | High fixed | 2026-08-23 |
| [FEAT-CORE-009](#feat-core-009) | EJS scaffolder | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-010](#feat-core-010) | ZIP generation | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-CORE-011](#feat-core-011) | Local bundle processor | Discovered - Comprehensive Test Suite Designed | 6 | 0 | None confirmed | — |
| [FEAT-UI-001](#feat-ui-001) | Main tool screen | Partially Tested - Smoke Passed | 8 | 0 | None confirmed | 2026-08-23 |
| [FEAT-UI-002](#feat-ui-002) | Theme persistence | Partially Tested - Storage Failure Fixed | 9 | 1 | Medium fixed | 2026-08-23 |
| [FEAT-UI-003](#feat-ui-003) | How To Use dropdown | Partially Tested - Defect Fixed | 9 | 1 | Medium fixed | 2026-08-23 |
| [FEAT-UI-004](#feat-ui-004) | Technical overview screen | Partially Tested - API Reference Fixed | 9 | 1 | Medium fixed | 2026-08-23 |
| [FEAT-UI-005](#feat-ui-005) | Download settings | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-006](#feat-ui-006) | Save As fallback | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-007](#feat-ui-007) | Input mode switching | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-008](#feat-ui-008) | HTML file upload | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-009](#feat-ui-009) | ZIP file upload | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-010](#feat-ui-010) | Feedback and reset | Partially Tested - Defect Fixed | 9 | 1 | Medium fixed | 2026-08-23 |
| [FEAT-UI-011](#feat-ui-011) | Format download workflow | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-012](#feat-ui-012) | Extracted ZIP workflow | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-013](#feat-ui-013) | TSX project workflow | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-014](#feat-ui-014) | EJS project workflow | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-015](#feat-ui-015) | ZIP bundle workflow | Discovered - Comprehensive Test Suite Designed | 8 | 0 | None confirmed | — |
| [FEAT-UI-016](#feat-ui-016) | URL scrape workflows | Partially Tested - Two Defects Fixed | 10 | 2 | Medium and Low fixed | 2026-08-23 |
| [FEAT-UI-017](#feat-ui-017) | Responsive and accessible use | Partially Tested - Defects Fixed | 10 | 2 | Medium fixed | 2026-08-23 |

## Browser UI (`FEAT-UI-*`)

<a id="feat-ui-001"></a>

### FEAT-UI-001 — Main tool screen

> As a browser user, I can open the root URL and access Uncluster.

**Expected behaviour**

Fiber serves dist/index.html plus CSS/JS; upload mode is active and processing actions are initially hidden.

**Edge cases**

Missing assets; JavaScript disabled; GSAP unavailable; unknown path.

**Validation rules**

No request input; actions stay hidden until input state exists.

**Dependencies**

Fiber static middleware; dist assets.

**Assumptions**

Server starts from a directory containing ./dist.

**Notes**

Desktop page identity, meaningful render, initial action state, and console health passed in Chrome; remaining category cases are pending.

**Status:** Partially Tested - Smoke Passed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (8):** [TC-FEAT-UI-001-001 … TC-FEAT-UI-001-008](test-cases.md#feat-ui-001)

<a id="feat-ui-002"></a>

### FEAT-UI-002 — Theme persistence

> As a user, I can switch dark/light themes and retain the choice.

**Expected behaviour**

data-theme, button label, and localStorage theme are updated on both screens.

**Edge cases**

Storage unavailable; invalid stored value; missing button.

**Validation rules**

UI emits dark or light.

**Dependencies**

DOM; localStorage; script.js.

**Assumptions**

localStorage normally permits writes.

**Notes**

Browser regression confirms theme application continues when localStorage throws; broader theme cases remain designed but unexecuted.

**Status:** Partially Tested - Storage Failure Fixed &nbsp;·&nbsp; **Severity:** Medium fixed &nbsp;·&nbsp; **Defect count:** 1 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (9):** [TC-FEAT-UI-002-001 … TC-FEAT-UI-002-009](test-cases.md#feat-ui-002)

**Defects:** [DEF-013](defects.md#def-013) (Medium, Fixed)

<a id="feat-ui-003"></a>

### FEAT-UI-003 — How To Use dropdown

> As a new user, I can reveal contextual instructions.

**Expected behaviour**

Click toggles the dropdown and outside click closes it.

**Edge cases**

Keyboard navigation; click inside; multiple dropdowns.

**Validation rules**

No data validation.

**Dependencies**

Dropdown DOM/CSS/listeners.

**Assumptions**

Pointer interaction is available.

**Notes**

Chrome red/green regression confirms native button triggers are keyboard-focusable; remaining category cases are pending.

**Status:** Partially Tested - Defect Fixed &nbsp;·&nbsp; **Severity:** Medium fixed &nbsp;·&nbsp; **Defect count:** 1 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (9):** [TC-FEAT-UI-003-001 … TC-FEAT-UI-003-009](test-cases.md#feat-ui-003)

**Defects:** [DEF-002](defects.md#def-002) (Medium, Fixed)

<a id="feat-ui-004"></a>

### FEAT-UI-004 — Technical overview screen

> As a technical user, I can read feature, pipeline, project, and API documentation.

**Expected behaviour**

/how-it-works.html renders documentation, theme/back controls, animations, and source link.

**Edge cases**

Direct-only navigation; stale content; GSAP failure; small viewport.

**Validation rules**

No input validation.

**Dependencies**

Static page; style/script; GSAP CDN.

**Assumptions**

Direct static URL is discoverable externally.

**Notes**

Browser regression confirms all 11 registered routes, payload families, and the health revision response are documented.

**Status:** Partially Tested - API Reference Fixed &nbsp;·&nbsp; **Severity:** Medium fixed &nbsp;·&nbsp; **Defect count:** 1 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (9):** [TC-FEAT-UI-004-001 … TC-FEAT-UI-004-009](test-cases.md#feat-ui-004)

**Defects:** [DEF-019](defects.md#def-019) (Medium, Fixed)

<a id="feat-ui-005"></a>

### FEAT-UI-005 — Download settings

> As a user, I can persist five filenames and picker preference.

**Expected behaviour**

downloadSettings localStorage restores fields and picker setting; malformed JSON falls back.

**Edge cases**

Blank names; malformed JSON; storage error; unsupported picker.

**Validation rules**

Invalid filename characters are replaced at download time.

**Dependencies**

Downloads dropdown; localStorage.

**Assumptions**

Extra stored keys are ignored.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-005-001 … TC-FEAT-UI-005-008](test-cases.md#feat-ui-005)

<a id="feat-ui-006"></a>

### FEAT-UI-006 — Save As fallback

> As a user, I can choose a save location or receive a browser download.

**Expected behaviour**

Supported picker writes Blob; cancel is non-error; other picker failures use anchor download.

**Edge cases**

AbortError; permission denial; unsupported API; URL revocation timing.

**Validation rules**

Required extension is appended case-insensitively.

**Dependencies**

File System Access API; Blob/Object URL.

**Assumptions**

Picker requires a compatible secure browser context.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-006-001 … TC-FEAT-UI-006-008](test-cases.md#feat-ui-006)

<a id="feat-ui-007"></a>

### FEAT-UI-007 — Input mode switching

> As a user, I can switch between upload and scrape modes.

**Expected behaviour**

Mode classes/visibility change; uploaded state/actions reset.

**Edge cases**

Switch with loaded file; rapid toggles; URL text persists until reset.

**Validation rules**

Mode originates from upload or scrape button.

**Dependencies**

DOM; uploadedHTML/uploadedZip/scrapeMode.

**Assumptions**

Clearing loaded state on switch is intentional.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-007-001 … TC-FEAT-UI-007-008](test-cases.md#feat-ui-007)

<a id="feat-ui-008"></a>

### FEAT-UI-008 — HTML file upload

> As a user, I can load .html/.htm and unlock four actions.

**Expected behaviour**

Non-ZIP selected file is read as text; HTML state set; actions shown; Upload becomes Finish Operation.

**Edge cases**

Empty/large file; read error; uppercase extension; non-HTML content.

**Validation rules**

accept is advisory; only .zip suffix changes branch.

**Dependencies**

Browser File API; action-state functions.

**Assumptions**

Server performs semantic processing later.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-008-001 … TC-FEAT-UI-008-008](test-cases.md#feat-ui-008)

<a id="feat-ui-009"></a>

### FEAT-UI-009 — ZIP file upload

> As a user, I can load a ZIP and unlock bundle.

**Expected behaviour**

Case-insensitive .zip suffix stores the File and shows only bundle action.

**Edge cases**

Corrupt/empty/large/renamed archive.

**Validation rules**

Client checks suffix; server validates archive.

**Dependencies**

File API; bundle workflow.

**Assumptions**

One selected ZIP is held in memory.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-009-001 … TC-FEAT-UI-009-008](test-cases.md#feat-ui-009)

<a id="feat-ui-010"></a>

### FEAT-UI-010 — Feedback and reset

> As a user, I see loading/toast feedback and can reset operations.

**Expected behaviour**

Buttons disable and gain ...; toast types then fades; Finish clears state/actions.

**Edge cases**

Concurrent clicks; long toast; missing button; scrape has no Finish control.

**Validation rules**

Client state must exist before action.

**Dependencies**

DOM; timers; global state.

**Assumptions**

Reset is driven by upload button.

**Notes**

Chrome red/green regression confirms toast role/live-region semantics; remaining feedback/reset cases are pending.

**Status:** Partially Tested - Defect Fixed &nbsp;·&nbsp; **Severity:** Medium fixed &nbsp;·&nbsp; **Defect count:** 1 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (9):** [TC-FEAT-UI-010-001 … TC-FEAT-UI-010-009](test-cases.md#feat-ui-010)

**Defects:** [DEF-003](defects.md#def-003) (Medium, Fixed)

<a id="feat-ui-011"></a>

### FEAT-UI-011 — Format download workflow

> As a user, I can format uploaded HTML and save it.

**Expected behaviour**

POST /api/format returns data used for one .html download and toast.

**Edge cases**

No upload; API/network error; cancel.

**Validation rules**

Client and server require nonblank HTML.

**Dependencies**

API format; download subsystem.

**Assumptions**

Successful response has success/data.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-011-001 … TC-FEAT-UI-011-008](test-cases.md#feat-ui-011)

<a id="feat-ui-012"></a>

### FEAT-UI-012 — Extracted ZIP workflow

> As a user, I can export separated HTML/CSS/JS.

**Expected behaviour**

POST /api/export blob downloads under configured .zip name.

**Edge cases**

No upload; partial fetch; server error; cancel.

**Validation rules**

Client requires uploadedHTML.

**Dependencies**

API export; downloader.

**Assumptions**

Failed external resources may be omitted.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-012-001 … TC-FEAT-UI-012-008](test-cases.md#feat-ui-012)

<a id="feat-ui-013"></a>

### FEAT-UI-013 — TSX project workflow

> As a user, I can download a generated TSX project.

**Expected behaviour**

POST /api/export-nodejs blob uses server/custom filename and downloads.

**Edge cases**

No upload; missing header; invalid project; cancel.

**Validation rules**

Client requires uploadedHTML.

**Dependencies**

TSX API; downloader.

**Assumptions**

Server attachment name is quoted.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-013-001 … TC-FEAT-UI-013-008](test-cases.md#feat-ui-013)

<a id="feat-ui-014"></a>

### FEAT-UI-014 — EJS project workflow

> As a user, I can download a generated EJS project.

**Expected behaviour**

POST /api/export-nodejs-ejs blob uses server/custom filename and downloads.

**Edge cases**

No upload; no qualifying partials; missing header; cancel.

**Validation rules**

Client requires uploadedHTML.

**Dependencies**

EJS API; downloader.

**Assumptions**

Small pages may remain in index.ejs.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-014-001 … TC-FEAT-UI-014-008](test-cases.md#feat-ui-014)

<a id="feat-ui-015"></a>

### FEAT-UI-015 — ZIP bundle workflow

> As a user, I can bundle a website ZIP into source/split/EJS outputs.

**Expected behaviour**

Multipart field file is sent to /api/bundle-zip and response downloads.

**Edge cases**

Missing ZIP; bad suffix/content; no index; zip slip; expansion limit.

**Validation rules**

Client needs uploadedZip; server needs .zip filename and valid archive.

**Dependencies**

Bundle API; downloader.

**Assumptions**

Result includes a derived site folder.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (8):** [TC-FEAT-UI-015-001 … TC-FEAT-UI-015-008](test-cases.md#feat-ui-015)

<a id="feat-ui-016"></a>

### FEAT-UI-016 — URL scrape workflows

> As a user, I can scrape a URL into ZIP, TSX, or EJS.

**Expected behaviour**

Nonblank URL reveals actions; three buttons call scrape endpoints and download ZIPs.

**Edge cases**

Blank/invalid/private URL; redirect; asset failure; blank custom name.

**Validation rules**

Client checks nonblank; server restricts http/https and public IPs.

**Dependencies**

Three scrape APIs; downloader.

**Assumptions**

Remote JavaScript is downloaded, not executed.

**Notes**

Browser red/green regressions confirm upload-only Format HTML is hidden in scrape mode and blank custom names use the server-provided scrape filename.

**Status:** Partially Tested - Two Defects Fixed &nbsp;·&nbsp; **Severity:** Medium and Low fixed &nbsp;·&nbsp; **Defect count:** 2 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (10):** [TC-FEAT-UI-016-001 … TC-FEAT-UI-016-010](test-cases.md#feat-ui-016)

**Defects:** [DEF-001](defects.md#def-001) (Medium, Fixed), [DEF-014](defects.md#def-014) (Low, Fixed)

<a id="feat-ui-017"></a>

### FEAT-UI-017 — Responsive and accessible use

> As a mobile, keyboard, or assistive user, I can operate the tool.

**Expected behaviour**

CSS has <=1000px layout rules and native inputs/buttons; current dropdown/toast semantics determine accessibility.

**Edge cases**

320px; zoom; keyboard dropdown; screen reader toast; reduced motion.

**Validation rules**

URL type is not programmatically checked with checkValidity.

**Dependencies**

HTML semantics; CSS; GSAP.

**Assumptions**

No ARIA live region exists.

**Notes**

Chrome red/green regressions confirm no nested link/button controls and no horizontal overflow at a 320px iframe viewport; remaining accessibility/responsive cases are pending.

**Status:** Partially Tested - Defects Fixed &nbsp;·&nbsp; **Severity:** Medium fixed &nbsp;·&nbsp; **Defect count:** 2 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (10):** [TC-FEAT-UI-017-001 … TC-FEAT-UI-017-010](test-cases.md#feat-ui-017)

**Defects:** [DEF-004](defects.md#def-004) (Medium, Fixed), [DEF-005](defects.md#def-005) (Medium, Fixed)

## HTTP API (`FEAT-API-*`)

<a id="feat-api-001"></a>

### FEAT-API-001 — POST /api/format

> As an API client, I can receive normalized HTML.

**Expected behaviour**

Malformed/blank requests return 400; valid html returns success/data; processing errors return 500.

**Edge cases**

Malformed JSON; missing/whitespace html; recoverable malformed markup; body limit.

**Validation rules**

JSON parse and strings.TrimSpace(html) nonempty.

**Dependencies**

Fiber; formatter.

**Assumptions**

HTML parser repairs most malformed input.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-001-001 … TC-FEAT-API-001-006](test-cases.md#feat-api-001)

<a id="feat-api-002"></a>

### FEAT-API-002 — POST /api/convert

> As an API client, I can receive a React component string.

**Expected behaviour**

Valid html is passed to ConvertToJSX with empty resource inputs and returned in data.

**Edge cases**

Quotes/braces; boolean/style/event attrs; malformed request.

**Validation rules**

JSON and nonblank html required.

**Dependencies**

Fiber; converter.

**Assumptions**

Output is not compiled server-side.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-002-001 … TC-FEAT-API-002-006](test-cases.md#feat-api-002)

<a id="feat-api-003"></a>

### FEAT-API-003 — POST /api/analyze

> As an API client, I can receive repeated-component suggestions.

**Expected behaviour**

Qualifying patterns return success/suggestions; none may omit suggestions via omitempty.

**Edge cases**

2 vs 3 repeats; structural tags; token boundaries; map order.

**Validation rules**

Count >=3, nonstructural tag, obvious normalized token.

**Dependencies**

Fiber; analyzer.

**Assumptions**

Ordering follows map iteration.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-003-001 … TC-FEAT-API-003-006](test-cases.md#feat-api-003)

<a id="feat-api-004"></a>

### FEAT-API-004 — POST /api/export

> As an API client, I can receive extracted.zip.

**Expected behaviour**

Validated HTML is extracted and archived with ZIP headers.

**Edge cases**

Empty resources; fetch failure; duplicates; large body; path errors.

**Validation rules**

Nonblank html; only successful nonempty resources archived.

**Dependencies**

Extractor; fetcher; zipper.

**Assumptions**

Archive is built in memory.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-004-001 … TC-FEAT-API-004-006](test-cases.md#feat-api-004)

<a id="feat-api-005"></a>

### FEAT-API-005 — POST /api/export-nodejs

> As an API client, I can receive a TSX project ZIP.

**Expected behaviour**

Extract/rewrite/generate/archive uses project-<Unix seconds> root.

**Edge cases**

Same-second requests; fallback conversion; external failures; size.

**Validation rules**

Nonblank html.

**Dependencies**

Fiber; extractor; TSX scaffolder; project zipper.

**Assumptions**

The archive is generated in memory and same-second requests can share the timestamp-derived project name.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-005-001 … TC-FEAT-API-005-006](test-cases.md#feat-api-005)

<a id="feat-api-006"></a>

### FEAT-API-006 — POST /api/export-nodejs-ejs

> As an API client, I can receive an Express/EJS project ZIP.

**Expected behaviour**

A valid nonblank HTML request is extracted, rewritten for EJS, scaffolded, and returned as an application/zip attachment rooted under a timestamped project name.

**Edge cases**

Malformed JSON; blank HTML; no qualifying partials; external fetch failures; large generated archive.

**Validation rules**

JSON body must parse and trimmed html must be nonempty.

**Dependencies**

Fiber; extractor; EJS scaffolder; project zipper.

**Assumptions**

The archive is generated in memory and timestamp naming is sufficiently unique for normal use.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-006-001 … TC-FEAT-API-006-006](test-cases.md#feat-api-006)

<a id="feat-api-007"></a>

### FEAT-API-007 — POST /api/scrape

> As an API client, I can scrape a public web URL into a localized resource ZIP.

**Expected behaviour**

A valid public http/https URL is fetched through safe HTTP controls, its referenced resources are localized, inline resources extracted, and an application/zip attachment is returned.

**Edge cases**

Malformed JSON; blank or malformed URL; unsupported scheme; private/reserved IP; redirect to private IP; missing assets; oversized responses.

**Validation rules**

JSON url is required; only public http/https fetches within redirect, timeout, and response-size limits are allowed.

**Dependencies**

Fiber; scraper; safehttp; fetcher; resource zipper.

**Assumptions**

Remote scripts are downloaded as data and are not executed by the server.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-007-001 … TC-FEAT-API-007-006](test-cases.md#feat-api-007)

<a id="feat-api-008"></a>

### FEAT-API-008 — POST /api/scrape-nodejs

> As an API client, I can scrape a public URL into a generated TSX project ZIP.

**Expected behaviour**

The guarded scrape result is rewritten for NodeJS, converted into a TSX scaffold, and returned as a timestamped application/zip attachment.

**Edge cases**

All scrape failures; partial resource fetches; pages without semantic sections; large projects; timestamp collision.

**Validation rules**

URL must parse as a public http/https target and all guarded-fetch limits apply.

**Dependencies**

Fiber; scraper; safehttp; TSX scaffolder; project zipper.

**Assumptions**

A partially fetched page may still generate when the main document succeeds.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-008-001 … TC-FEAT-API-008-006](test-cases.md#feat-api-008)

<a id="feat-api-009"></a>

### FEAT-API-009 — POST /api/scrape-nodejs-ejs

> As an API client, I can scrape a public URL into a generated EJS project ZIP.

**Expected behaviour**

The guarded scrape result is rewritten for EJS, scaffolded, and returned as a timestamped application/zip attachment.

**Edge cases**

All scrape failures; no qualifying partials; partial assets; oversized projects; timestamp collision.

**Validation rules**

URL must parse as a public http/https target and all guarded-fetch limits apply.

**Dependencies**

Fiber; scraper; safehttp; EJS scaffolder; project zipper.

**Assumptions**

Small pages may remain wholly in index.ejs.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-009-001 … TC-FEAT-API-009-006](test-cases.md#feat-api-009)

<a id="feat-api-010"></a>

### FEAT-API-010 — POST /api/bundle-zip

> As an API client, I can upload an existing website ZIP and receive source, split, and EJS bundle outputs.

**Expected behaviour**

A multipart file field with a .zip filename is safely extracted, processed into bundle artifacts, and returned as an application/zip attachment.

**Edge cases**

Missing multipart file; non-.zip name; corrupt archive; no HTML index; path traversal; symlink-like paths; expansion over 512 MiB; duplicate names.

**Validation rules**

Multipart field file is required; filename must end .zip case-insensitively; archive must be valid, contained, and within expansion limits.

**Dependencies**

Fiber multipart handling; bundle processor; project zipper; temporary directory.

**Assumptions**

Temporary working data is removed after the request.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-010-001 … TC-FEAT-API-010-006](test-cases.md#feat-api-010)

<a id="feat-api-011"></a>

### FEAT-API-011 — GET /api/health

> As an operator, I can query service health and build identity.

**Expected behaviour**

GET returns JSON containing status healthy, service uncluster-api, and revision derived from build information or development fallback.

**Edge cases**

Missing VCS build settings; dirty build metadata; unsupported method; middleware errors.

**Validation rules**

No request body or authentication is required.

**Dependencies**

Fiber; runtime/debug build info.

**Assumptions**

A healthy response reflects process reachability, not downstream network availability.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-API-011-001 … TC-FEAT-API-011-006](test-cases.md#feat-api-011)

## Command line (`FEAT-CLI-*`)

<a id="feat-cli-001"></a>

### FEAT-CLI-001 — CLI parsing and help

> As a CLI user, I can select an input, output mode, destination, and help.

**Expected behaviour**

The CLI parses one input path plus -to, -out, -dest, and help flags, validates the requested mode, prints usage for help/errors, and dispatches the matching workflow.

**Edge cases**

No arguments; help anywhere; flag missing a value; unknown flag; unknown mode; multiple positional inputs; flags after input.

**Validation rules**

An existing input is required for work; -to must be one of split, nodejs, nodejs-ejs, format, jsx, analyze, or bundle; bundle destination is optional.

**Dependencies**

os.Args parser; filesystem; mode handlers.

**Assumptions**

Additional positional arguments are ignored by the current parser.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-001-001 … TC-FEAT-CLI-001-006](test-cases.md#feat-cli-001)

<a id="feat-cli-002"></a>

### FEAT-CLI-002 — CLI format

> As a CLI user, I can format an HTML file to stdout or a file.

**Expected behaviour**

The input file is parsed and normalized; without -out the formatted HTML is printed, while -out writes formatted.html beneath the target.

**Edge cases**

Unreadable input; malformed HTML repaired by parser; empty file; missing output directory; unwritable output.

**Validation rules**

Input path must be readable; -out changes output from stdout to a fixed filename.

**Dependencies**

Formatter; filesystem; CLI parser.

**Assumptions**

The HTML parser accepts and repairs many malformed documents.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-002-001 … TC-FEAT-CLI-002-006](test-cases.md#feat-cli-002)

<a id="feat-cli-003"></a>

### FEAT-CLI-003 — CLI JSX

> As a CLI user, I can convert HTML to a React component on stdout or disk.

**Expected behaviour**

The input is converted with ConvertToJSX; stdout is used without -out and MainComponent.tsx is written with -out.

**Edge cases**

Quotes/braces; inline styles; boolean/event attributes; malformed HTML; unwritable destination.

**Validation rules**

Input must be readable; converter output is not compiled by the CLI.

**Dependencies**

Converter; filesystem; CLI parser.

**Assumptions**

Resource arrays are empty in this direct mode.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-003-001 … TC-FEAT-CLI-003-006](test-cases.md#feat-cli-003)

<a id="feat-cli-004"></a>

### FEAT-CLI-004 — CLI analyze

> As a CLI user, I can inspect repeated component suggestions.

**Expected behaviour**

The analyzer returns JSON-formatted suggestions to stdout or writes component-suggestions.json under -out.

**Edge cases**

No suggestions; exactly two versus three patterns; structural tags; nondeterministic map order; write failure.

**Validation rules**

Input must be readable; analyzer requires qualifying repeated nonstructural patterns.

**Dependencies**

Analyzer; JSON encoder; filesystem.

**Assumptions**

Suggestion ordering follows current map traversal unless explicitly sorted.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-004-001 … TC-FEAT-CLI-004-006](test-cases.md#feat-cli-004)

<a id="feat-cli-005"></a>

### FEAT-CLI-005 — CLI split

> As a CLI user, I can split HTML into index, CSS, JavaScript, and a manifest.

**Expected behaviour**

The extractor writes split resources beneath the output directory and includes a manifest describing produced files.

**Edge cases**

No inline resources; external fetch failure; duplicate resource names; nonexistent or unwritable output.

**Validation rules**

-out is required for split; input must be readable; only eligible script types are extracted.

**Dependencies**

Extractor; fetcher; filesystem.

**Assumptions**

Failed external resources may remain referenced or be omitted according to extractor behavior.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-005-001 … TC-FEAT-CLI-005-006](test-cases.md#feat-cli-005)

<a id="feat-cli-006"></a>

### FEAT-CLI-006 — CLI TSX project

> As a CLI user, I can scaffold a TSX project directory from HTML.

**Expected behaviour**

Extraction, NodeJS rewriting, component generation, resources, configuration, server, and README files are written under the chosen output directory.

**Edge cases**

Missing -out; invalid project directory basename; no semantic sections; external fetch failure; existing files.

**Validation rules**

-out is required; input must be readable; filesystem paths must be creatable.

**Dependencies**

Extractor; TSX scaffolder; filesystem.

**Assumptions**

The output directory basename becomes the project name.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-006-001 … TC-FEAT-CLI-006-006](test-cases.md#feat-cli-006)

<a id="feat-cli-007"></a>

### FEAT-CLI-007 — CLI EJS project

> As a CLI user, I can scaffold an Express/EJS project directory from HTML.

**Expected behaviour**

Extraction, EJS rewriting, partial detection, resources, package, server, and README are written beneath -out.

**Edge cases**

Missing -out; no qualifying partials; existing files; invalid project basename; permission failure.

**Validation rules**

-out is required; input must be readable; output paths must be creatable.

**Dependencies**

Extractor; EJS scaffolder; filesystem.

**Assumptions**

Small semantic regions may remain inline rather than becoming partials.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-007-001 … TC-FEAT-CLI-007-006](test-cases.md#feat-cli-007)

<a id="feat-cli-008"></a>

### FEAT-CLI-008 — CLI bundle

> As a CLI user, I can bundle an HTML file or website ZIP into derived outputs.

**Expected behaviour**

The bundle processor accepts HTML or ZIP input, chooses an output destination, and writes source, split, EJS, and ZIP artifacts.

**Edge cases**

Unsupported extension; corrupt ZIP; no index; traversal; expansion limit; existing stale destination data.

**Validation rules**

Input must be .html/.htm/.zip and readable; destination must remain contained and writable.

**Dependencies**

Bundle processor; filesystem; archive/zip.

**Assumptions**

When -dest is omitted the current CLI derives a destination from the input.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-008-001 … TC-FEAT-CLI-008-006](test-cases.md#feat-cli-008)

<a id="feat-cli-009"></a>

### FEAT-CLI-009 — uncluster-split utility

> As a legacy utility user, I can split one HTML file with explicit flags and optional manifest.

**Expected behaviour**

-input and -output are required; resources are extracted and written, and split-manifest.json is written by default unless -manifest=false.

**Edge cases**

Missing flags; unreadable input; external fetch failure; disabled manifest; unwritable output.

**Validation rules**

Go flag parsing requires nonblank -input and -output; manifest is boolean and defaults true.

**Dependencies**

Extractor; fetcher; flag package; filesystem.

**Assumptions**

This utility remains separately buildable from the main CLI.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CLI-009-001 … TC-FEAT-CLI-009-006](test-cases.md#feat-cli-009)

## Transformation core (`FEAT-CORE-*`)

<a id="feat-core-001"></a>

### FEAT-CORE-001 — Formatter engine

> As a caller, I can normalize HTML into deterministic indented markup.

**Expected behaviour**

The parser renders document, element, text, comment, and doctype nodes; preserves raw script/style/pre/textarea content; escapes text/attributes; and self-closes void elements.

**Edge cases**

Empty or fragment input; malformed DOM; whitespace-only text; mixed inline/block children; raw content containing closing syntax.

**Validation rules**

Input must be parsable by x/net/html; indentation uses tabs and void-element rules are fixed.

**Dependencies**

golang.org/x/net/html; strings builder.

**Assumptions**

Parser repair is part of observed behavior.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-001-001 … TC-FEAT-CORE-001-006](test-cases.md#feat-core-001)

<a id="feat-core-002"></a>

### FEAT-CORE-002 — Resource extraction

> As a caller, I can separate eligible inline and external CSS/JavaScript from HTML.

**Expected behaviour**

Inline style and executable script blocks are removed into resource lists; eligible external http/https styles/scripts are fetched and successful references rewritten; NodeJS/EJS rewrite helpers adjust paths.

**Edge cases**

JSON-LD; empty blocks; exact rel matching; uppercase schemes; Google Fonts; duplicate URLs; failed or empty fetches.

**Validation rules**

Only allowed JavaScript MIME types are extracted; external styles require exact stylesheet relation and lowercase http/https URL; successful nonempty fetches are retained.

**Dependencies**

HTML parser; fetcher; path rewrite helpers.

**Assumptions**

External fetch failure does not necessarily abort the overall extraction.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-002-001 … TC-FEAT-CORE-002-006](test-cases.md#feat-core-002)

<a id="feat-core-003"></a>

### FEAT-CORE-003 — Guarded fetching

> As a caller, I can fetch remote text or binary assets with SSRF and size controls.

**Expected behaviour**

Only http/https requests to resolved public IPs are dialed; redirects are revalidated; timeouts, user-agent, redirect count, and a 25 MiB response limit are enforced.

**Edge cases**

DNS rebinding; mixed public/private answers; IPv4-mapped IPv6; redirect loops; slow headers/body; exact and over-limit bodies; unsupported scheme.

**Validation rules**

Scheme must be http or https; loopback/private/unspecified/link-local/multicast and listed special ranges are blocked; at most 10 redirects; response at most 25 MiB.

**Dependencies**

net/http transport; net.Resolver; safehttp policy; fetcher wrappers.

**Assumptions**

DNS and network availability are external dependencies.

**Notes**

WebAssembly Go regression confirms independent client timeouts share the guarded bounded transport and connection pool.

**Status:** Partially Tested - Transport Reuse Fixed &nbsp;·&nbsp; **Severity:** Medium fixed &nbsp;·&nbsp; **Defect count:** 1 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (7):** [TC-FEAT-CORE-003-001 … TC-FEAT-CORE-003-007](test-cases.md#feat-core-003)

**Defects:** [DEF-018](defects.md#def-018) (Medium, Fixed)

<a id="feat-core-004"></a>

### FEAT-CORE-004 — Scraper localization

> As a caller, I can mirror a public page and its referenced assets into local resource paths.

**Expected behaviour**

The main HTML is guarded-fetched, resource URLs are discovered, CSS dependencies followed, assets fetched, HTML references rewritten, and inline CSS/JavaScript extracted into generated resources.

**Edge cases**

srcset; CSS url() nesting; inline style URLs; JSON-LD; collisions; query strings; missing content type; failed assets; absolute-root paths.

**Validation rules**

Main URL and resource requests use guarded fetching; only discovered successful resources enter output; filenames are sanitized and collision-adjusted.

**Dependencies**

safehttp; fetcher; extractor-like parsing; URL resolution.

**Assumptions**

Map iteration can affect fetch order and collision suffix assignment.

**Notes**

Five focused Go scraper regressions cover portable HTML paths, srcset, data-script preservation, inline resources, and downloaded CSS dependency rewriting.

**Status:** Partially Tested - Four Defects Fixed &nbsp;·&nbsp; **Severity:** High fixed &nbsp;·&nbsp; **Defect count:** 4 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (11):** [TC-FEAT-CORE-004-001 … TC-FEAT-CORE-004-011](test-cases.md#feat-core-004)

**Defects:** [DEF-010](defects.md#def-010) (High, Fixed), [DEF-011](defects.md#def-011) (High, Fixed), [DEF-012](defects.md#def-012) (High, Fixed), [DEF-015](defects.md#def-015) (High, Fixed)

<a id="feat-core-005"></a>

### FEAT-CORE-005 — JSX conversion

> As a caller, I can transform HTML into a React MainComponent source string.

**Expected behaviour**

Document wrappers and resource tags are omitted; attributes/events/styles are mapped; text/comments/elements are serialized; void tags self-close; and MainComponent is exported.

**Edge cases**

JSX braces/quotes; entity decoding; presence-only booleans; namespaced attributes; whitespace around inline elements; invalid CSS declarations.

**Validation rules**

HTML parses successfully; known attribute/event mappings and style conversion rules apply.

**Dependencies**

x/net/html; formatter helpers; string serialization.

**Assumptions**

The output is returned as text and is not compiler-validated.

**Notes**

Four real Go converter regressions executed red/green in Chrome via a js/wasm test binary; remaining category cases are pending.

**Status:** Partially Tested - Defects Fixed &nbsp;·&nbsp; **Severity:** Critical fixed &nbsp;·&nbsp; **Defect count:** 4 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (10):** [TC-FEAT-CORE-005-001 … TC-FEAT-CORE-005-010](test-cases.md#feat-core-005)

**Defects:** [DEF-006](defects.md#def-006) (Critical, Fixed), [DEF-007](defects.md#def-007) (High, Fixed), [DEF-008](defects.md#def-008) (High, Fixed), [DEF-009](defects.md#def-009) (High, Fixed)

<a id="feat-core-006"></a>

### FEAT-CORE-006 — Section/list TSX conversion

> As a generator, I can convert semantic sections and repeated list-like siblings into reusable TSX.

**Expected behaviour**

Repeated same-tag siblings with varying extractable fields can produce interfaces, data arrays, and map rendering; other nodes become indented TSX with TODO handlers.

**Edge cases**

Two versus three items; nested lists; mixed tags; identical values; missing text/image/link fields; event attributes.

**Validation rules**

List conversion requires the implemented repeated sibling pattern and extractable varying fields.

**Dependencies**

x/net/html; JSX attribute and serialization helpers.

**Assumptions**

Heuristics intentionally favor recognizable card/list structures.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-006-001 … TC-FEAT-CORE-006-006](test-cases.md#feat-core-006)

<a id="feat-core-007"></a>

### FEAT-CORE-007 — Component analyzer

> As a caller, I can discover repeated DOM patterns and receive starter component suggestions.

**Expected behaviour**

Elements are grouped by tag/class/id patterns, attribute and child frequencies calculated, structural tags excluded, a count of at least three required, names inferred from obvious tokens, and starter JSX emitted.

**Edge cases**

Exactly two/three matches; class-order differences; token boundaries; nested duplicates; map-order variation; empty documents.

**Validation rules**

Candidate count is at least three; structural containers are blocked; naming requires a recognized normalized token.

**Dependencies**

x/net/html; pattern maps and token heuristics.

**Assumptions**

Ordering currently follows Go map iteration.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-007-001 … TC-FEAT-CORE-007-006](test-cases.md#feat-core-007)

<a id="feat-core-008"></a>

### FEAT-CORE-008 — TSX scaffolder

> As a caller, I can generate a runnable Express/Vite/TypeScript project from extracted page data.

**Expected behaviour**

Package/config/server/README/source files are generated; semantic sections up to the implemented depth become components; CSS and successful external resources are written; fallback HTML is used when needed.

**Edge cases**

No sections; duplicate section names/content; content outside semantic boundaries; inline/external JavaScript; invalid project name; template metacharacters; file errors.

**Validation rules**

All generated paths must remain under the project root; templates receive parsed page/resource data.

**Dependencies**

Converter; text/template; filesystem abstraction; extracted resources.

**Assumptions**

The scaffolder does not run npm or compile the generated project.

**Notes**

WebAssembly Go regressions confirm generated projects load original JavaScript, preserve direct root content plus repeated sections, and adapt localized stylesheet asset paths for public serving in both TSX and EJS outputs.

**Status:** Partially Tested - Three Defects Fixed &nbsp;·&nbsp; **Severity:** High fixed &nbsp;·&nbsp; **Defect count:** 3 &nbsp;·&nbsp; **Last tested:** 2026-08-23

**Test cases (9):** [TC-FEAT-CORE-008-001 … TC-FEAT-CORE-008-009](test-cases.md#feat-core-008)

**Defects:** [DEF-016](defects.md#def-016) (High, Fixed), [DEF-017](defects.md#def-017) (High, Fixed), [DEF-020](defects.md#def-020) (High, Fixed)

<a id="feat-core-009"></a>

### FEAT-CORE-009 — EJS scaffolder

> As a caller, I can generate a runnable Express/EJS project from page data.

**Expected behaviour**

Server/package/README/index are generated; sufficiently large multiline semantic candidates may become partials; original document structure is retained with include markers; resources are written beneath public.

**Edge cases**

Candidates below 500 bytes or 15 newlines; duplicate names; nested candidates; no partials; resource path collisions; template/file errors.

**Validation rules**

Only implemented semantic boundary tags qualify and must meet both size and newline thresholds.

**Dependencies**

x/net/html; text/template; filesystem abstraction; extracted resources.

**Assumptions**

The generated server exposes the root render plus public static assets.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-009-001 … TC-FEAT-CORE-009-006](test-cases.md#feat-core-009)

<a id="feat-core-010"></a>

### FEAT-CORE-010 — ZIP generation

> As a caller, I can package resource or project files into an in-memory ZIP.

**Expected behaviour**

Resource ZIP creation propagates errors; project ZIP roots entries beneath the project name, logs/skips individual failed entries, and errors when no entry is written.

**Edge cases**

Empty files; duplicate names; slash normalization; large aggregate output; one versus all writer failures.

**Validation rules**

Archive entry names derive from supplied safe relative paths; at least one project entry must be written.

**Dependencies**

archive/zip; bytes buffers; resource/project abstractions.

**Assumptions**

Archives are built wholly in memory.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-010-001 … TC-FEAT-CORE-010-006](test-cases.md#feat-core-010)

<a id="feat-core-011"></a>

### FEAT-CORE-011 — Local bundle processor

> As a caller, I can turn local HTML or a website ZIP into source, split, EJS, and ZIP deliverables.

**Expected behaviour**

Input is loaded or safely extracted within 512 MiB; the best index is selected; referenced HTML, srcset, style, and recursive CSS assets are rewritten and copied; generated outputs are written under a contained destination.

**Edge cases**

ZIP traversal; no index; multiple indexes; nested CSS cycles; unused assets; stale destination files; symlinks; expansion boundary; filename collisions.

**Validation rules**

All extracted and written paths must remain within their roots; aggregate uncompressed bytes cannot exceed 512 MiB; only supported HTML/ZIP inputs proceed.

**Dependencies**

archive/zip; extractor; EJS scaffolder; filesystem; URL/path rewriting.

**Assumptions**

Only selected output subdirectories are reset by current implementation.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CORE-011-001 … TC-FEAT-CORE-011-006](test-cases.md#feat-core-011)

## Configuration and deployment (`FEAT-CFG-*`)

<a id="feat-cfg-001"></a>

### FEAT-CFG-001 — API PORT

> As an operator, I can choose the listening port through the environment.

**Expected behaviour**

The root server reads PORT verbatim and defaults to 3000 only when it is empty.

**Edge cases**

Unset, empty, nonnumeric, privileged, occupied, or whitespace port values.

**Validation rules**

No explicit validation is performed before app.Listen.

**Dependencies**

Operating system environment; Fiber listener.

**Assumptions**

Deployment supplies a valid port when overriding the default.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CFG-001-001 … TC-FEAT-CFG-001-006](test-cases.md#feat-cfg-001)

<a id="feat-cfg-002"></a>

### FEAT-CFG-002 — HTTP middleware and body/CORS policy

> As an API consumer, I receive global logging, panic recovery, body-size enforcement, and CORS behavior.

**Expected behaviour**

Fiber applies a 50 MiB body limit, logger, recoverer, and CORS allowing any origin with GET/POST/PUT/DELETE/OPTIONS and listed headers; the custom error handler emits JSON errors.

**Edge cases**

Preflight; disallowed method/header; body exactly/over limit; panic; framework versus handler error shapes.

**Validation rules**

Global middleware executes in registration order; BodyLimit is 50 MiB.

**Dependencies**

Fiber middleware and error handler.

**Assumptions**

AllowOrigins '*' is intentional current behavior and there is no authentication middleware.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CFG-002-001 … TC-FEAT-CFG-002-006](test-cases.md#feat-cfg-002)

<a id="feat-cfg-003"></a>

### FEAT-CFG-003 — Safety limits

> As an operator, I can rely on bounded remote fetching and archive extraction.

**Expected behaviour**

Remote fetches use 10 redirects, 10-second dial and 30-second client timeouts, 25 MiB bodies, and public-IP restrictions; bundle extraction is contained and capped at 512 MiB.

**Edge cases**

Exact limit; one byte over; compressed bombs; redirect chain; reserved CIDRs; slow streams; path normalization variants.

**Validation rules**

Each coded size, timeout, redirect, IP, and containment check applies at its boundary.

**Dependencies**

safehttp; bundle processor; Fiber BodyLimit.

**Assumptions**

Limits are independent and may reject at different layers.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CFG-003-001 … TC-FEAT-CFG-003-006](test-cases.md#feat-cfg-003)

<a id="feat-cfg-004"></a>

### FEAT-CFG-004 — Build/install/deploy commands

> As a maintainer, I can build, install, clean, and deploy the server and CLIs through repository configuration.

**Expected behaviour**

Make targets build CLI/server, install/uninstall binaries, and clean outputs; Nixpacks installs Go, builds the root server to out, and starts ./out.

**Edge cases**

Missing Go; permission denied install path; stale binaries; cross-platform shell differences; failed dependency download.

**Validation rules**

Commands assume their documented tools and writable destinations; Nixpacks start requires a successful root build.

**Dependencies**

Make; Go toolchain; Nixpacks; filesystem permissions.

**Assumptions**

Make commands are Unix-oriented while the current validation host is Windows.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CFG-004-001 … TC-FEAT-CFG-004-006](test-cases.md#feat-cfg-004)

<a id="feat-cfg-005"></a>

### FEAT-CFG-005 — Generated project runtime

> As a recipient of generated output, I can install and run the TSX or EJS project.

**Expected behaviour**

TSX output uses Vite with src plus an Express server for dist; EJS output serves public and renders index; both generated servers read PORT and default to 8080.

**Edge cases**

Missing npm install; build failure; port collision; missing generated assets; production versus development command differences.

**Validation rules**

Generated package scripts and file topology must match their templates; PORT is used verbatim or defaults 8080.

**Dependencies**

Node.js/npm; Express; Vite/TypeScript or EJS; generated files.

**Assumptions**

Node runtime execution is outside the Go generator and requires Node installation.

**Notes**

Derived from current implementation; comprehensive suite designed; execution pending.

**Status:** Discovered - Comprehensive Test Suite Designed &nbsp;·&nbsp; **Severity:** None confirmed &nbsp;·&nbsp; **Defect count:** 0 &nbsp;·&nbsp; **Last tested:** —

**Test cases (6):** [TC-FEAT-CFG-005-001 … TC-FEAT-CFG-005-006](test-cases.md#feat-cfg-005)

