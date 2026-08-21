# Uncluster 90+ Engineering Roadmap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (- [ ]) syntax for tracking.

**Goal:** Raise Uncluster from the reviewed 66/100 baseline to a defensible 90+ by making its conversion pipeline secure, correct, extensively tested, reproducibly deployed, benchmarked, and professionally documented.

**Architecture:** Preserve the existing Go packages for parsing, conversion, extraction, analysis, bundling, and project generation, but place them behind a single typed pipeline service shared by the HTTP server and CLI. Route all remote I/O through a hardened fetch boundary, validate generated output through target-specific build checks, and use automated quality gates for every supported workflow.

**Tech Stack:** Go, Fiber, golang.org/x/net/html, Go test/fuzz/race tooling, GitHub Actions, Docker, Railway or an equivalent container host, govulncheck, staticcheck/golangci-lint, CodeQL, benchmark and golden-fixture tooling.

**Spec:** docs/ROADMAP_TO_90.md — the approved score-driven design and implementation plan are intentionally co-located in this document.

## Global Constraints

- Preserve the existing REST API and CLI behavior unless a breaking change is explicitly versioned.
- All remote network access must deny private, loopback, link-local, multicast, and reserved address ranges after every DNS resolution and redirect.
- Every request must have explicit time, byte, redirect, asset-count, and concurrency limits.
- Generated JSX/TSX must never execute source-site event-handler strings by default.
- Generated projects must compile in clean temporary directories during integration tests.
- Tests must not depend on the public internet; use local HTTP fixtures and deterministic archives.
- The main branch must remain releasable after each merged task.
- Do not claim accuracy, security, deployment, or performance results without an automated artifact that proves the claim.
- Use conventional commits and keep each task independently reviewable.
- Add no mandatory AI or external SaaS dependency to the core transformation path.
- The target Go version must be pinned consistently in go.mod, CI, Docker, and release workflows.

---

## 1. Baseline and Target Score

The repository was reviewed at 66/100:

| Dimension | Current | Target | Evidence required for target |
|---|---:|---:|---|
| Problem and ambition | 13/15 | 14/15 | Multiple validated output targets, documented corpus, measurable conversion goals |
| Architecture and judgment | 17/25 | 22/25 | Shared pipeline, safe network boundary, typed errors, resource budgets, thin transports |
| Implementation depth | 15/20 | 18/20 | Correct escaping, deterministic transformation, robust asset rewriting, build validation |
| Testing and reliability | 4/15 | 14/15 | Broad unit/integration/security/fuzz coverage, race-clean suite, enforced coverage threshold |
| Deployment and operations | 5/10 | 9/10 | Reproducible container, automated releases, health/readiness, logs, metrics, verified deployment |
| Originality and authorship | 8/10 | 8/10 | Preserve transparent history and publish design decisions and benchmarks |
| Documentation and presentation | 4/5 | 5/5 | Accurate README, threat model, architecture, demo, examples, limitations, license |
| **Total** | **66/100** | **90/100** | All mandatory release gates below pass |

A score of 90 is not achieved by adding files. It is achieved only when the acceptance tests and release gates in this document pass on the main branch.

## 2. Definition of Done for 90+

Uncluster is 90-ready only when all of the following are true:

- A clean clone passes gofmt, go vet, static analysis, vulnerability scanning, unit tests, integration tests, fuzz smoke tests, and race detection in CI.
- At least 80% statement coverage is enforced across the core packages, with security-sensitive packages at 90% or higher.
- Remote fetches cannot reach private infrastructure through direct IPs, DNS rebinding, alternate numeric IP forms, or redirects.
- A request cannot exceed configured limits for HTML bytes, individual assets, total asset bytes, redirects, concurrent downloads, archive entries, decompressed archive size, or wall-clock time.
- Golden tests prove correct conversion for text escaping, attributes, inline styles, void elements, scripts, comments, SVG, malformed HTML, Unicode, and framework-specific edge cases.
- Generated Vite/React, Express, and EJS projects install and build in clean test environments.
- HTTP and CLI paths call the same pipeline and produce equivalent manifests for equivalent inputs.
- The public deployment is produced from the same Docker image verified in CI.
- Releases include checksums, an SBOM, provenance, changelog notes, and versioned binaries for supported platforms.
- README claims are traceable to tests, benchmark artifacts, a live demo, or documented limitations.
- A real LICENSE file and third-party attribution policy are present.
- No high or critical vulnerability is open without a documented, time-bounded exception.

---

## 3. Target Component Boundaries

The intended end-state structure is:

| Path | Responsibility |
|---|---|
| internal/pipeline | One application service coordinating parse, extract, analyze, convert, scaffold, validate, and package stages |
| internal/netguard | URL validation, DNS/IP policy, redirect enforcement, HTTP client configuration, and resource budgets |
| internal/converter | Pure DOM-to-target conversion with explicit escaping and event-handler policy |
| internal/extractor | Asset discovery, normalization, reference rewriting, and content classification |
| internal/analyzer | Deterministic structural analysis and component suggestions |
| internal/generator | Vite/React, Express, and EJS project generation |
| internal/validator | Build, manifest, path, and output-policy validation |
| internal/api | Fiber request decoding, authentication/rate-limit hooks, response encoding, and error mapping |
| internal/observability | Structured logging, request IDs, metrics, and stage timings |
| cmd/uncluster | CLI-only parsing and presentation |
| testdata/corpus | Licensed, synthetic, and adversarial fixtures with expected outputs |
| docs | Architecture, threat model, benchmark method, examples, and operational runbooks |

### Core interfaces

The implementation should converge on these stable contracts:

~~~go
package pipeline

type Target string

const (
    TargetBundle Target = "bundle"
    TargetJSX    Target = "jsx"
    TargetTSX    Target = "tsx"
    TargetEJS    Target = "ejs"
)

type Request struct {
    Source      Source
    Target      Target
    Destination string
    Limits      Limits
}

type Result struct {
    OutputDir string
    Manifest  Manifest
    Metrics   Metrics
    Warnings  []Warning
}

type Service interface {
    Convert(ctx context.Context, req Request) (Result, error)
}
~~~

~~~go
package netguard

type Limits struct {
    RequestTimeout    time.Duration
    MaxRedirects      int
    MaxHTMLBytes      int64
    MaxAssetBytes     int64
    MaxTotalBytes     int64
    MaxAssets         int
    MaxConcurrent     int
}

type Response struct {
    FinalURL    *url.URL
    StatusCode  int
    ContentType string
    Body        []byte
}

type Fetcher interface {
    Fetch(ctx context.Context, rawURL string, limits Limits) (Response, error)
}
~~~

All callers must depend on these interfaces rather than creating ad-hoc HTTP clients or duplicating pipeline orchestration.

---

## Phase 0 — Establish the Measurement Harness

### Task 1: Record the baseline and create repository quality gates

**Files:**
- Create: docs/engineering-scorecard.md
- Create: .github/workflows/ci.yml
- Create: scripts/check-coverage.sh
- Modify: README.md
- Modify: go.mod

**Produces:** A repeatable command set and CI artifact showing the starting quality state.

- [ ] **Step 1:** Document the current score, known risks, package inventory, and supported outputs in docs/engineering-scorecard.md.
- [ ] **Step 2:** Pin the supported Go version in go.mod and use the identical version in CI.
- [ ] **Step 3:** Add CI jobs for gofmt verification, go vet, go test ./..., go test -race ./..., staticcheck, and govulncheck.
- [ ] **Step 4:** Add scripts/check-coverage.sh that fails below 80% overall coverage and prints package-level coverage.
- [ ] **Step 5:** Run the full gate locally.

~~~bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
go test ./... -coverprofile=coverage.out
bash scripts/check-coverage.sh coverage.out 80
staticcheck ./...
govulncheck ./...
~~~

- [ ] **Step 6:** Preserve current failures as named issues or an explicit baseline table; do not disable failing checks to obtain a green badge.
- [ ] **Step 7:** Commit with message: chore: establish engineering quality baseline

**Acceptance gate:** CI runs on pull requests and main; failures identify exact packages and commands.

### Task 2: Repair current release and repository hygiene defects

**Files:**
- Modify: nixpacks.toml
- Modify: README.md
- Create: LICENSE
- Create: SECURITY.md
- Create: CONTRIBUTING.md
- Create: .github/dependabot.yml

- [ ] **Step 1:** Change the Nixpacks build from the nonexistent ./api path to the actual server entry point.
- [ ] **Step 2:** Replace the README clone placeholder with https://github.com/omariomari2/uncluster.
- [ ] **Step 3:** Add the actual chosen license file and make README license wording match it.
- [ ] **Step 4:** Document responsible disclosure and supported versions in SECURITY.md.
- [ ] **Step 5:** Add contribution commands and test expectations.
- [ ] **Step 6:** Enable monthly Go module and GitHub Actions dependency updates.
- [ ] **Step 7:** Verify the deployment build command from a clean clone.
- [ ] **Step 8:** Commit with message: chore: repair deployment and repository metadata

**Acceptance gate:** No documented path or command refers to a missing file or directory.

---

## Phase 1 — Secure Every Input Boundary

### Task 3: Add a single SSRF-safe network boundary

**Files:**
- Create: internal/netguard/policy.go
- Create: internal/netguard/resolver.go
- Create: internal/netguard/client.go
- Create: internal/netguard/policy_test.go
- Create: internal/netguard/client_test.go
- Modify: existing fetcher and scraper packages to accept netguard.Fetcher

**Consumes:** netguard.Limits and netguard.Fetcher contracts defined above.

**Produces:** One audited path for all external HTTP access.

- [ ] **Step 1:** Write table tests that reject localhost, IPv4 and IPv6 loopback, RFC1918, link-local, multicast, unspecified, documentation, carrier-grade NAT, and cloud-metadata addresses.
- [ ] **Step 2:** Add tests for decimal, hexadecimal, octal, IPv4-mapped IPv6, trailing-dot hostnames, user-info confusion, and mixed-case schemes.
- [ ] **Step 3:** Add a local redirect server test proving a public-looking URL cannot redirect to a blocked address.
- [ ] **Step 4:** Implement scheme allowlisting for HTTP and HTTPS only.
- [ ] **Step 5:** Resolve all A and AAAA records and reject the request if any selected destination violates policy.
- [ ] **Step 6:** Re-run policy after each redirect and bind the dial operation to the validated address.
- [ ] **Step 7:** Replace every direct http.Get or custom unguarded client in fetcher, scraper, extractor, and handlers.
- [ ] **Step 8:** Run focused and race tests.

~~~bash
go test ./internal/netguard -run Test -v
go test -race ./internal/netguard ./internal/fetcher ./internal/scraper
~~~

- [ ] **Step 9:** Commit with message: security: enforce SSRF-safe remote fetching

**Acceptance gate:** Security tests prove that direct, DNS-based, and redirect-based private-network access is denied.

### Task 4: Enforce resource budgets and cancellation

**Files:**
- Create: internal/netguard/limits.go
- Create: internal/netguard/limits_test.go
- Modify: internal/bundle/bundle.go
- Modify: fetch, scrape, and extraction paths
- Modify: main.go configuration loading

- [ ] **Step 1:** Add failing tests for response bodies one byte over each configured limit.
- [ ] **Step 2:** Add archive tests for excessive file count, decompressed bytes, path depth, duplicate names, and compression-ratio bombs.
- [ ] **Step 3:** Implement limited readers for HTML and assets and an aggregate byte counter for each conversion.
- [ ] **Step 4:** Add a bounded worker pool for asset downloads.
- [ ] **Step 5:** Propagate context cancellation through fetch, extraction, conversion, generation, and ZIP writing.
- [ ] **Step 6:** Expose safe defaults through environment variables with documented upper bounds.
- [ ] **Step 7:** Add handler tests mapping limit failures to HTTP 413 or 422 without leaking internal paths.
- [ ] **Step 8:** Commit with message: security: enforce conversion resource budgets

**Acceptance gate:** No input can cause unbounded network reads, archive extraction, file generation, goroutine creation, or wall-clock execution.

### Task 5: Add API abuse controls

**Files:**
- Create: internal/api/middleware.go
- Create: internal/api/middleware_test.go
- Modify: main.go
- Modify: API documentation

- [ ] **Step 1:** Add request-ID, body-size, concurrency, timeout, and per-client rate-limit middleware.
- [ ] **Step 2:** Make authentication optional for local mode but mandatory by configuration for public deployments.
- [ ] **Step 3:** Add tests for rate-limit reset, concurrent saturation, oversized bodies, cancellation, and sanitized errors.
- [ ] **Step 4:** Add explicit CORS origin configuration; reject wildcard credentials.
- [ ] **Step 5:** Document deployment-safe defaults.
- [ ] **Step 6:** Commit with message: security: protect public conversion endpoints

**Acceptance gate:** The public deployment cannot be used as an unrestricted anonymous scraper.

---

## Phase 2 — Make Conversion Correct and Deterministic

### Task 6: Harden JSX and TSX rendering

**Files:**
- Modify: internal/converter/jsx.go
- Create: internal/converter/jsx_test.go
- Create: internal/converter/testdata/
- Create: internal/converter/golden_test.go

- [ ] **Step 1:** Add golden fixtures for quotes, backslashes, braces, template delimiters, ampersands, Unicode, JSX-sensitive text, comments, SVG attributes, inline styles, and malformed input.
- [ ] **Step 2:** Add tests proving source onclick and related event strings are dropped by default and reported as warnings.
- [ ] **Step 3:** Implement target-aware text and attribute escaping.
- [ ] **Step 4:** Define explicit policies for style conversion, data and aria attributes, custom elements, boolean attributes, SVG names, and unknown attributes.
- [ ] **Step 5:** Parse generated JSX/TSX with a real frontend parser or validate it by compiling the generated fixture project.
- [ ] **Step 6:** Run golden tests twice and assert byte-identical output.
- [ ] **Step 7:** Commit with message: fix: make JSX conversion safe and deterministic

**Acceptance gate:** Every generated fixture parses or compiles, and event-handler source is never executed implicitly.

### Task 7: Make asset extraction and rewriting reliable

**Files:**
- Modify: internal/extractor package
- Create: internal/extractor/extractor_test.go
- Create: internal/extractor/rewrite_test.go
- Create: testdata/assets/

- [ ] **Step 1:** Add fixtures covering CSS url(), srcset, base href, protocol-relative URLs, query strings, fragments, duplicate filenames, MIME/extension mismatch, nested stylesheets, fonts, and data URLs.
- [ ] **Step 2:** Define canonical asset identities using final URL plus normalized content metadata.
- [ ] **Step 3:** Generate collision-resistant local names while retaining readable extensions.
- [ ] **Step 4:** Rewrite references only after successful download and validation; preserve the original reference with a warning on recoverable failure.
- [ ] **Step 5:** Deduplicate identical content by digest.
- [ ] **Step 6:** Verify every local reference in the manifest resolves to an output file.
- [ ] **Step 7:** Commit with message: fix: make asset localization deterministic

**Acceptance gate:** The asset manifest is complete, reproducible, collision-safe, and contains no dangling rewritten reference.

### Task 8: Validate every generated project

**Files:**
- Create: internal/validator/project.go
- Create: internal/validator/project_test.go
- Create: testdata/projects/
- Modify: internal/generator package
- Modify: pipeline result model

- [ ] **Step 1:** Define validators for required files, parseable manifests, normalized relative paths, and forbidden absolute paths.
- [ ] **Step 2:** Generate Vite/React, Express, and EJS projects from representative corpus fixtures.
- [ ] **Step 3:** Run package installation and production builds in isolated temporary directories using a pinned Node version and dependency cache.
- [ ] **Step 4:** Fail conversion when the selected target cannot build; return stage-specific diagnostics.
- [ ] **Step 5:** Add snapshot assertions for package manifests and entry points.
- [ ] **Step 6:** Commit with message: test: validate generated project builds

**Acceptance gate:** All supported targets build from clean output without accessing source-repository files.

---

## Phase 3 — Consolidate the Architecture

### Task 9: Introduce the shared pipeline service

**Files:**
- Create: internal/pipeline/service.go
- Create: internal/pipeline/types.go
- Create: internal/pipeline/errors.go
- Create: internal/pipeline/service_test.go
- Modify: main.go
- Modify: cmd/uncluster/main.go
- Modify: cmd/uncluster-split as applicable

- [ ] **Step 1:** Write contract tests using fake fetcher, extractor, converter, generator, and validator stages.
- [ ] **Step 2:** Implement typed stage errors with safe public messages and wrapped internal causes.
- [ ] **Step 3:** Move orchestration from handlers and CLI commands into pipeline.Service.
- [ ] **Step 4:** Make HTTP and CLI adapters translate their inputs into the same pipeline.Request.
- [ ] **Step 5:** Add equivalence tests proving both adapters produce matching manifests.
- [ ] **Step 6:** Remove duplicated extraction and generation logic only after equivalence tests pass.
- [ ] **Step 7:** Commit with message: refactor: unify server and CLI conversion pipelines

**Acceptance gate:** Transport code contains no conversion business logic.

### Task 10: Split the HTTP layer and add versioned error contracts

**Files:**
- Create: internal/api/server.go
- Create: internal/api/handlers.go
- Create: internal/api/errors.go
- Create: internal/api/handlers_test.go
- Modify: main.go

- [ ] **Step 1:** Define a stable JSON error shape containing code, message, request_id, stage, and retryable.
- [ ] **Step 2:** Add handler tests for valid requests, invalid targets, oversized inputs, canceled conversions, safe-fetch failures, generation errors, and internal failures.
- [ ] **Step 3:** Move route registration and handlers out of main.go.
- [ ] **Step 4:** Keep main.go limited to configuration, dependency construction, lifecycle, and shutdown.
- [ ] **Step 5:** Add graceful shutdown and reject new work while draining active conversions.
- [ ] **Step 6:** Commit with message: refactor: isolate API transport and lifecycle

**Acceptance gate:** main.go is an application-composition root, not a handler collection.

### Task 11: Add structured observability

**Files:**
- Create: internal/observability/logging.go
- Create: internal/observability/metrics.go
- Modify: pipeline stages and API middleware
- Create: docs/operations.md

- [ ] **Step 1:** Define stage metrics for duration, input bytes, assets, output bytes, warnings, failures, and target type.
- [ ] **Step 2:** Add structured JSON logging with request IDs and no raw document or secret content.
- [ ] **Step 3:** Add /healthz and /readyz with distinct semantics.
- [ ] **Step 4:** Expose Prometheus-compatible metrics or document the selected host-native equivalent.
- [ ] **Step 5:** Add tests ensuring logs redact URLs with credentials, local paths, tokens, and query parameters marked sensitive.
- [ ] **Step 6:** Document alerts for error rate, p95 latency, saturation, memory, and remote-fetch rejection rate.
- [ ] **Step 7:** Commit with message: feat: add conversion observability

**Acceptance gate:** Operators can diagnose a failing stage without logging user documents.

---

## Phase 4 — Build a Serious Reliability Suite

### Task 12: Create the licensed conversion corpus

**Files:**
- Create: testdata/corpus/manifest.json
- Create: testdata/corpus/synthetic/
- Create: testdata/corpus/adversarial/
- Create: testdata/corpus/licensed/
- Create: docs/corpus-policy.md

- [ ] **Step 1:** Define manifest fields for source, license, expected targets, expected warnings, and prohibited outputs.
- [ ] **Step 2:** Add at least 30 synthetic fixtures across static pages, forms, SVG, tables, media, malformed HTML, and framework-export patterns.
- [ ] **Step 3:** Add at least 20 adversarial fixtures for traversal, SSRF URLs, archive bombs, hostile attributes, huge nodes, and encoding edge cases.
- [ ] **Step 4:** Include external fixtures only when redistribution rights are documented.
- [ ] **Step 5:** Add a test that fails if a corpus entry lacks license and expectation metadata.
- [ ] **Step 6:** Commit with message: test: add licensed conversion corpus

**Acceptance gate:** Every corpus input has explicit provenance and machine-readable expectations.

### Task 13: Expand unit and integration coverage

**Files:**
- Add tests beside converter, extractor, analyzer, scraper, fetcher, generator, validator, pipeline, API, CLI, and bundle packages
- Create: internal/testutil/

- [ ] **Step 1:** Add deterministic local HTTP, DNS-policy, archive, filesystem, and clock helpers.
- [ ] **Step 2:** Cover success, boundary, invalid input, partial failure, cancellation, and retry behavior for each package.
- [ ] **Step 3:** Add CLI tests for exit codes, stdout/stderr separation, exact destination, overwrite policy, and cancellation.
- [ ] **Step 4:** Add server integration tests using httptest with the real middleware and fake remote services.
- [ ] **Step 5:** Add generated-project build tests for every target.
- [ ] **Step 6:** Raise coverage in stages: 65%, 75%, then 80%, without excluding difficult packages.
- [ ] **Step 7:** Commit with message: test: cover transformation and delivery pipelines

**Acceptance gate:** Overall coverage is at least 80%; netguard, bundle, validator, and pipeline are at least 90%.

### Task 14: Add fuzz, race, and determinism gates

**Files:**
- Create fuzz tests in converter, extractor, analyzer, bundle, and URL policy packages
- Modify: .github/workflows/ci.yml

- [ ] **Step 1:** Add seed corpora from every regression fixture.
- [ ] **Step 2:** Assert no panic, traversal, invalid UTF-8 output, unbounded allocation, or nondeterministic result.
- [ ] **Step 3:** Run short fuzz smoke jobs on every PR and longer scheduled jobs nightly.
- [ ] **Step 4:** Run the complete suite with the race detector.
- [ ] **Step 5:** Add repeated determinism tests using at least 20 identical conversions.
- [ ] **Step 6:** Store crashing inputs as regression fixtures.
- [ ] **Step 7:** Commit with message: test: add fuzz race and determinism gates

**Acceptance gate:** CI is race-clean; no known fuzz crash remains unresolved.

---

## Phase 5 — Reproducible Delivery and Operations

### Task 15: Build one production container

**Files:**
- Create: Dockerfile
- Create: .dockerignore
- Create: compose.yaml
- Modify: nixpacks.toml or remove it after migration
- Create: scripts/smoke.sh

- [ ] **Step 1:** Add a multi-stage Go build using a pinned builder and minimal non-root runtime.
- [ ] **Step 2:** Include CA certificates and only the files required at runtime.
- [ ] **Step 3:** Set a read-only root filesystem and writable temporary/output mounts in Compose.
- [ ] **Step 4:** Add container health and readiness checks.
- [ ] **Step 5:** Build and run the image locally, then execute a real conversion through scripts/smoke.sh.
- [ ] **Step 6:** Scan the image for high and critical vulnerabilities.
- [ ] **Step 7:** Commit with message: build: add reproducible production container

**Acceptance gate:** The same image digest tested in CI is deployed.

### Task 16: Add release automation and supply-chain artifacts

**Files:**
- Create: .github/workflows/release.yml
- Create: .goreleaser.yml
- Create: CHANGELOG.md
- Create: docs/releasing.md

- [ ] **Step 1:** Build CLI binaries for supported Linux, macOS, and Windows architectures.
- [ ] **Step 2:** Generate checksums and an SPDX or CycloneDX SBOM.
- [ ] **Step 3:** Sign release artifacts and container provenance using GitHub OIDC-compatible tooling.
- [ ] **Step 4:** Publish container images only from version tags.
- [ ] **Step 5:** Create a release candidate and install each binary in an isolated smoke job.
- [ ] **Step 6:** Document semantic versioning and rollback.
- [ ] **Step 7:** Commit with message: ci: automate signed releases

**Acceptance gate:** A tagged release is reproducible, checksummed, signed, scanned, and smoke-tested.

### Task 17: Verify the hosted service

**Files:**
- Create: .github/workflows/deploy.yml
- Create: docs/deployment.md
- Modify: README.md

- [ ] **Step 1:** Deploy from the verified container rather than an unrelated build path.
- [ ] **Step 2:** Use environment-scoped secrets and protected production approvals.
- [ ] **Step 3:** Run post-deploy health, readiness, conversion, SSRF-rejection, and limit-enforcement smoke tests.
- [ ] **Step 4:** Publish deployment URL, version, commit SHA, and status in the health response.
- [ ] **Step 5:** Add rollback to the last known-good image digest.
- [ ] **Step 6:** Commit with message: ci: verify production deployment

**Acceptance gate:** Production behavior is tied to a passing commit and observable release artifact.

---

## Phase 6 — Demonstrate Measurable Engineering Quality

### Task 18: Add conversion accuracy and performance benchmarks

**Files:**
- Create: benchmarks/
- Create: internal/benchmark/
- Create: docs/benchmarks.md
- Modify: .github/workflows/ci.yml

- [ ] **Step 1:** Define accuracy metrics: build success, reference integrity, required-node preservation, forbidden-output absence, warning count, and deterministic digest.
- [ ] **Step 2:** Define performance metrics: wall time, allocations, peak memory, network bytes, and output size by corpus class.
- [ ] **Step 3:** Add Go benchmarks for parser, converter, analyzer, extractor, and bundle operations.
- [ ] **Step 4:** Run end-to-end benchmarks on small, medium, and large licensed fixtures.
- [ ] **Step 5:** Store machine-readable baseline results.
- [ ] **Step 6:** Fail CI on severe regressions, using explicit thresholds documented in benchmarks.md.
- [ ] **Step 7:** Publish charts generated from CI artifacts; never hand-edit claimed numbers.
- [ ] **Step 8:** Commit with message: perf: benchmark conversion quality and cost

**Acceptance gate:** README performance and accuracy claims link to reproducible benchmark artifacts.

### Task 19: Add an end-to-end demo and examples

**Files:**
- Create: examples/
- Create: docs/demo.md
- Modify: README.md

- [ ] **Step 1:** Add redistributable before-and-after examples for HTML, JSX, TSX, EJS, and a complete Vite project.
- [ ] **Step 2:** Record a short demo showing CLI, API, output inspection, and clean generated build.
- [ ] **Step 3:** Add architecture and data-flow diagrams.
- [ ] **Step 4:** Document limitations: dynamic runtime behavior, proprietary exports, unsupported framework features, unsafe scripts, and visual fidelity.
- [ ] **Step 5:** Add a contribution guide for new corpus fixtures and conversion rules.
- [ ] **Step 6:** Commit with message: docs: publish reproducible conversion examples

**Acceptance gate:** A new user can understand, run, validate, and evaluate the project without reading source code.

---

## Phase 7 — Final 90+ Release Gate

### Task 20: Run the release audit

**Files:**
- Update: docs/engineering-scorecard.md
- Update: README.md
- Update: CHANGELOG.md
- Create: docs/audits/90-readiness.md

- [ ] **Step 1:** Run every local gate from a clean clone.
- [ ] **Step 2:** Confirm the full CI matrix is green on the release commit.
- [ ] **Step 3:** Confirm all generated targets build across the corpus.
- [ ] **Step 4:** Confirm SSRF, traversal, archive, rate-limit, timeout, and prompt/script safety tests pass.
- [ ] **Step 5:** Confirm coverage thresholds and benchmark regression thresholds pass.
- [ ] **Step 6:** Confirm production uses the verified release digest.
- [ ] **Step 7:** Confirm README, API docs, live demo, license, security policy, and benchmark claims match reality.
- [ ] **Step 8:** Perform an independent security review of netguard, archive handling, generated-code boundaries, and public endpoints.
- [ ] **Step 9:** Record every residual risk with owner, severity, mitigation, and review date.
- [ ] **Step 10:** Rescore each rubric dimension using linked evidence.
- [ ] **Step 11:** Tag version 1.0.0 only if the evidence supports at least 90/100.
- [ ] **Step 12:** Commit with message: release: complete 90-readiness audit

---

## 4. Recommended Execution Order

| Milestone | Tasks | Expected outcome |
|---|---|---|
| M1: Trustworthy baseline | 1–2 | Green, honest repository foundation |
| M2: Safe public service | 3–5 | SSRF-safe, bounded, abuse-resistant remote processing |
| M3: Correct transformations | 6–8 | Compilable, deterministic, reference-complete output |
| M4: Maintainable core | 9–11 | Shared pipeline, thin adapters, typed errors, observability |
| M5: Proven reliability | 12–14 | Corpus-driven tests, coverage, fuzzing, race safety |
| M6: Reproducible delivery | 15–17 | Verified container, releases, deployment, rollback |
| M7: Evidence and adoption | 18–19 | Benchmarks, examples, demo, transparent limitations |
| M8: 90+ release | 20 | Independent audit and evidence-backed score |

Security tasks 3–5 must finish before promoting or advertising the public scrape endpoints. Conversion correctness tasks 6–8 must finish before claiming production-ready JSX/TSX generation. Release work must not bypass the test corpus or security gate.

## 5. Suggested Pull Request Sequence

1. PR 1 — Baseline CI, license, metadata, and deployment-path repair
2. PR 2 — SSRF-safe network boundary
3. PR 3 — Resource limits, archive defenses, and API abuse controls
4. PR 4 — JSX/TSX correctness and golden fixtures
5. PR 5 — Asset rewriting and generated-project validation
6. PR 6 — Shared pipeline and thin API/CLI adapters
7. PR 7 — Structured logs, metrics, health, and graceful shutdown
8. PR 8 — Licensed corpus and broad unit/integration coverage
9. PR 9 — Fuzz, race, and deterministic-output gates
10. PR 10 — Production container and automated releases
11. PR 11 — Verified deployment and rollback
12. PR 12 — Benchmarks, examples, demo, and 90-readiness audit

Every PR must include its own tests, documentation changes, and rollback note. Avoid one “90-point rewrite” branch; the roadmap is designed to improve main through independently valuable, reviewable increments.

## 6. Risks and Non-Goals

### Primary risks

- SSRF defenses that validate only the original hostname and miss redirects or DNS rebinding.
- Superficial test coverage that counts lines without validating generated projects.
- Golden snapshots that approve invalid output without compiling it.
- Network-dependent tests that become flaky or silently skip.
- Benchmarks run on changing environments without recorded metadata.
- Security middleware that protects HTTP but leaves CLI/archive paths unsafe.
- Documentation drifting ahead of implementation.
- Treating a deployed URL as proof of reliability without release provenance.

### Non-goals for the 90 roadmap

- Building a visual page editor.
- Executing arbitrary source-site JavaScript.
- Guaranteeing pixel-perfect conversion for every framework.
- Adding required cloud AI services.
- Supporting authenticated scraping of third-party sites.
- Circumventing licensing, robots, access-control, or terms-of-service restrictions.
- Replacing deterministic component analysis with an opaque LLM-only pipeline.

## 7. Weekly Evidence Review

At the end of each milestone:

1. Update docs/engineering-scorecard.md with links to merged PRs and CI artifacts.
2. Re-run the full command set from a clean clone.
3. Add each fixed defect as a regression test.
4. Re-evaluate whether README statements remain accurate.
5. Review security limits against new code paths.
6. Record benchmark movement rather than relying on impressions.
7. Stop the milestone if a quality gate is disabled, skipped, or made non-blocking without a written exception.

The objective is not to manufacture a high score. The objective is to create enough verifiable engineering evidence that a 90+ score becomes the conservative conclusion.
