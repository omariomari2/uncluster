# Optional Runtime Capture

**Status:** Candidate idea, not approved for implementation
**Recorded:** 2026-09-29

## Problem Statement

How might Uncluster capture public, runtime-rendered marketing pages without making
its normal static workflow slow or bundling a browser automation stack?

## Recommended Direction

Keep the existing HTTP/DOM scraper as the default. Add an optional runtime-capture
path that launches the user's installed Chrome in headless mode with an isolated
temporary profile and controls only the small Chrome DevTools Protocol surface
needed to navigate, wait, observe resources, and read the rendered DOM.

Use a deterministic static-page gate plus an explicit override:

```text
--render=auto     static first; use Chrome only when runtime rendering is indicated
--render=never    current static behavior
--render=always   always use installed Chrome
```

The first implementation candidate is Rod connected to a Chrome process launched
by Uncluster. It supports the project's current Go 1.21 baseline and does not need
to download or bundle Chromium when given an explicit browser address. Preserve a
small internal capture interface so the implementation can later be replaced by a
narrow custom CDP client if measured binary growth is unacceptable.

## Key Assumptions to Validate

- [ ] Static signals can identify most client-rendered shells without starting
  Chrome for ordinary pages. Test against a mixed corpus of static and SPA pages.
- [ ] Rod's measured binary-size and startup-cost increase remains acceptable.
  Compare release binaries and cold/warm capture times before integration.
- [ ] Rendered DOM plus runtime-request URLs materially improves the existing
  HTML, EJS, and TSX outputs. Test using failures the static scraper cannot capture.
- [ ] Re-fetching public runtime assets through the guarded Go fetcher is adequate
  for the target class of public marketing pages. Record cookie- or header-bound
  resources as unsupported rather than silently omitting them.

## MVP Scope

- Local CLI only; do not expose browser rendering through the HTTP server yet.
- Require installed Chrome or an explicitly supplied compatible executable.
- Use a fresh temporary browser profile, fixed viewport, bounded network/DOM quiet
  periods, and a hard whole-job timeout.
- Capture the final URL, rendered document HTML, requested resource URLs and MIME
  types, console failures, and load failures.
- Feed the capture into the existing localization and HTML/EJS/TSX generation
  pipeline rather than creating separate generators.
- Support `auto`, `never`, and `always` selection, with the chosen strategy and
  reason recorded in `uncluster-capture.json`.

## Not Doing (and Why)

- Bundling Chrome, Playwright, or Node.js: this conflicts with the lightweight-core
  requirement.
- Screenshot comparison in Uncluster: runtime capture needs browser state, not a
  visual regression system.
- LLM-based capture decisions: the gate must stay fast and reproducible.
- Automatic clicking or arbitrary interaction exploration: it can mutate remote
  state and cannot be made generally deterministic.
- Logged-in profile capture: exposing real cookies, passwords, and extensions is
  an unacceptable default.
- Claiming arbitrary application behavior is editable: a rendered DOM records the
  result of execution, not framework state, closures, routers, or event handlers.

## Open Questions

- What binary-size and runtime budgets define "lightweight"?
- Should runtime output default to an editable snapshot or preserve original app
  scripts as a less-editable compatibility replay?
- Is the initial asset strategy URL observation plus guarded re-fetching, or should
  MHTML be retained as a capped fallback resource archive?
- Which deterministic score should trigger `--render=auto`, and how will false
  positives and negatives be reported?

## Reference Notes

- Chrome headless `--dump-dom` executes page scripts before serializing the DOM,
  but does not provide the readiness and resource controls required for the full
  design: <https://developer.chrome.com/docs/automation-and-testing/headless-cli>
- Chrome 136 and later require remote debugging to use a non-default data
  directory: <https://developer.chrome.com/blog/remote-debugging-port>
- The CDP snapshot methods can expose rendered DOM/layout data and MHTML resources:
  <https://chromedevtools.github.io/devtools-protocol/>
- Rod's current module targets Go 1.21:
  <https://raw.githubusercontent.com/go-rod/rod/main/go.mod>
