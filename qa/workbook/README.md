# Uncluster quality workbook

Canonical source of truth for feature discovery, test design, execution evidence,
and defect tracking.

> **Generated file — do not edit by hand.**
> Every page here is rendered from [`qa/source/quality-model.json`](../source/quality-model.json).
> Edit the JSON, then regenerate with `go run ./qa/tools/qualitydoc`.

| Page | Contents |
|---|---|
| [Features](features.md) | Every discovered feature with user story, expected behaviour, edge cases, validation rules, dependencies, and assumptions |
| [Test cases](test-cases.md) | Designed scenarios per feature, with steps, fixtures, and expected results |
| [Defects](defects.md) | Defect records with reproduction, root cause, fix, and verification |
| [Executions](executions.md) | Commands actually run, with exit codes and observed results |
| [Iterations](iterations.md) | Per-iteration coverage, risks, and confidence |

## Snapshot

| Metric | Value |
|---|---:|
| Features documented | 53 |
| Test cases designed | 373 |
| Test cases executed | 21 (5.6%) |
| Defects recorded | 20 |
| Defects open | 0 |
| Execution records | 36 |
| Iterations completed | 1 |

| Field | Value |
|---|---|
| Schema version | 1 |
| Created | 2026-08-23 |
| Last updated | 2026-08-23 |
| Authoritative basis | Current worktree implementation |
| Workbook status | Published - Markdown workbook at qa/workbook, generated from this model by 'go run ./qa/tools/qualitydoc' |
| Execution environment | Windows; Go 1.27; 'go test ./...' runs natively and passes as of 2026-08-23 |
| Phase | Phase 1 reconciled; Phase 2 comprehensive suites designed; execution in progress |

## Test case status

| Status | Count | Share |
|---|---:|---:|
| Not Run | 352 | 94.4% |
| Passed | 21 | 5.6% |

## Test cases by execution type

| Execution type | Count | Share |
|---|---:|---:|
| Automated-Go | 157 | 42.1% |
| Automated-Browser | 144 | 38.6% |
| Automated-HTTP | 66 | 17.7% |
| Manual | 6 | 1.6% |

## Test cases by scenario category

| Category | Count | Share |
|---|---:|---:|
| Boundary Condition | 53 | 14.2% |
| Error Path | 53 | 14.2% |
| Happy Path | 53 | 14.2% |
| Invalid Input | 53 | 14.2% |
| Performance | 53 | 14.2% |
| Permission / Security | 53 | 14.2% |
| Accessibility | 17 | 4.6% |
| Mobile / Responsive | 17 | 4.6% |
| Transformation Regression | 8 | 2.1% |
| Accessibility Regression | 3 | 0.8% |
| Generation Regression | 2 | 0.5% |
| Boundary Regression | 1 | 0.3% |
| Cross-Output Regression | 1 | 0.3% |
| Documentation Regression | 1 | 0.3% |
| Error Handling Regression | 1 | 0.3% |
| Mobile / Responsive Regression | 1 | 0.3% |
| Performance Regression | 1 | 0.3% |
| Security / Transformation Regression | 1 | 0.3% |
| Workflow Regression | 1 | 0.3% |

## Feature status

| Status | Count | Share |
|---|---:|---:|
| Discovered - Comprehensive Test Suite Designed | 42 | 79.2% |
| Partially Tested - Defect Fixed | 2 | 3.8% |
| Partially Tested - Defects Fixed | 2 | 3.8% |
| Partially Tested - API Reference Fixed | 1 | 1.9% |
| Partially Tested - Four Defects Fixed | 1 | 1.9% |
| Partially Tested - Smoke Passed | 1 | 1.9% |
| Partially Tested - Storage Failure Fixed | 1 | 1.9% |
| Partially Tested - Three Defects Fixed | 1 | 1.9% |
| Partially Tested - Transport Reuse Fixed | 1 | 1.9% |
| Partially Tested - Two Defects Fixed | 1 | 1.9% |

## Defects by severity

| Severity | Count | Share |
|---|---:|---:|
| High | 10 | 50.0% |
| Medium | 8 | 40.0% |
| Critical | 1 | 5.0% |
| Low | 1 | 5.0% |

## Defects by status

| Status | Count | Share |
|---|---:|---:|
| Fixed | 20 | 100.0% |

## Inventory reconciliation

| Surface | Discovered | Feature records |
|---|---:|---:|
| Static screens | 2 | 17 |
| HTTP routes | 11 | 11 |
| CLI modes | 7 | 9 |
| Transformation core | — | 11 |
| Configuration and deployment | — | 5 |

**Status:** Reconciled &nbsp;·&nbsp; **Unmatched surfaces:** 0

**Evidence:** main.go; cmd/uncluster/main.go; cmd/uncluster-split/main.go; dist/*; internal/**/*.go; Makefile; nixpacks.toml; README.md; go.mod

## Blockers

| ID | Area | Status | Detail | Impact |
|---|---|---|---|---|
| BLK-001 | Canonical workbook | Resolved - superseded by Markdown workbook | Artifact dependency loader, template picker, mark script, Node, and @oai/artifact-tool are unavailable, so the XLSX export could not be produced. Superseded by a Markdown workbook rendered from this model by qa/tools/qualitydoc. | None. The canonical workbook is qa/workbook/, regenerated with 'go run ./qa/tools/qualitydoc'. |
| BLK-002 | Go tests | Resolved - no longer reproducible | Application Control previously blocked generated *.test.exe in temp and workspace paths. Re-verified on 2026-08-23: 'go test -count=1 ./...' now runs natively and every package passes, so the block no longer applies and the WebAssembly fallback in qa/wasm is no longer required. | None. Native assertions execute normally; see EXEC-036. |

## Controlled vocabularies

Values in use across the model. Keep new records within these sets so the
rollups above stay meaningful.

- **Feature status:** `Discovered - Comprehensive Test Suite Designed`, `Partially Tested - API Reference Fixed`, `Partially Tested - Defect Fixed`, `Partially Tested - Defects Fixed`, `Partially Tested - Four Defects Fixed`, `Partially Tested - Smoke Passed`, `Partially Tested - Storage Failure Fixed`, `Partially Tested - Three Defects Fixed`, `Partially Tested - Transport Reuse Fixed`, `Partially Tested - Two Defects Fixed`
- **Feature severity:** `Critical fixed`, `High fixed`, `Medium and Low fixed`, `Medium fixed`, `None confirmed`
- **Test case status:** `Not Run`, `Passed`
- **Execution type:** `Automated-Browser`, `Automated-Go`, `Automated-HTTP`, `Manual`
- **Scenario category:** `Accessibility`, `Accessibility Regression`, `Boundary Condition`, `Boundary Regression`, `Cross-Output Regression`, `Documentation Regression`, `Error Handling Regression`, `Error Path`, `Generation Regression`, `Happy Path`, `Invalid Input`, `Mobile / Responsive`, `Mobile / Responsive Regression`, `Performance`, `Performance Regression`, `Permission / Security`, `Security / Transformation Regression`, `Transformation Regression`, `Workflow Regression`
- **Defect status:** `Fixed`
- **Defect severity:** `Critical`, `High`, `Low`, `Medium`
