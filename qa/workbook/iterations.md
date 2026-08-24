# Iterations

> Generated from [`qa/source/quality-model.json`](../source/quality-model.json)
> by `go run ./qa/tools/qualitydoc`. Do not edit by hand.
> [Back to the coverage summary](README.md).

## ITER-001 — 2026-08-23

| Metric | Value |
|---|---:|
| Features tested | 11 |
| Defects found | 20 |
| Defects fixed | 20 |
| Suspected defects outstanding | 0 |
| Confidence score | 92% |

**Coverage summary**

53 reconciled feature records and 373 designed cases. Eight UI, four converter, five scraper, three TSX/EJS scaffolder, and one guarded-transport regressions pass after independent red/green remediation loops. All 20 discovered defects now have confirmed fixes. The repository builds, all runnable native package suites pass, and analyzer/safehttp/nodejs pass through the WebAssembly fallback where native execution is blocked.

**Remaining risks**

- Workbook artifact runtime unavailable
- Windows Application Control blocks native analyzer and safehttp test executables; equivalent WebAssembly suites pass
- Most of the 373 cases remain unexecuted
- Windows Application Control still requires WebAssembly fallback for some package executables

