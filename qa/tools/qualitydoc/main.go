// Command qualitydoc renders qa/source/quality-model.json into the canonical
// Markdown quality workbook under qa/workbook.
//
// The JSON model stays the single source of truth; every Markdown file this
// command writes is a generated view of it. Regenerate with:
//
//	go run ./qa/tools/qualitydoc
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type model struct {
	Metadata   metadata    `json:"metadata"`
	Features   []feature   `json:"features"`
	TestCases  []testCase  `json:"testCases"`
	Defects    []defect    `json:"defects"`
	Executions []execution `json:"executions"`
	Iterations []iteration `json:"iterations"`
	Blockers   []blocker   `json:"blockers"`
}

type metadata struct {
	Product              string         `json:"product"`
	SchemaVersion        int            `json:"schemaVersion"`
	CreatedDate          string         `json:"createdDate"`
	UpdatedDate          string         `json:"updatedDate"`
	AuthoritativeBasis   string         `json:"authoritativeBasis"`
	WorkbookStatus       string         `json:"workbookStatus"`
	ExecutionEnvironment string         `json:"executionEnvironment"`
	Phase                string         `json:"phase"`
	Reconciliation       reconciliation `json:"inventoryReconciliation"`
}

type reconciliation struct {
	Status                      string `json:"status"`
	Screens                     int    `json:"screens"`
	UIFeatureRecords            int    `json:"uiFeatureRecords"`
	APIRoutes                   int    `json:"apiRoutes"`
	APIFeatureRecords           int    `json:"apiFeatureRecords"`
	CLIModes                    int    `json:"cliModes"`
	CLIFeatureRecords           int    `json:"cliFeatureRecords"`
	CoreFeatureRecords          int    `json:"coreFeatureRecords"`
	ConfigurationFeatureRecords int    `json:"configurationFeatureRecords"`
	UnmatchedSurfaces           int    `json:"unmatchedSurfaces"`
	Evidence                    string `json:"evidence"`
}

type feature struct {
	FeatureID        string   `json:"featureId"`
	FeatureName      string   `json:"featureName"`
	UserStory        string   `json:"userStory"`
	ExpectedBehavior string   `json:"expectedBehavior"`
	EdgeCases        string   `json:"edgeCases"`
	ValidationRules  string   `json:"validationRules"`
	Dependencies     string   `json:"dependencies"`
	Assumptions      string   `json:"assumptions"`
	TestCases        []string `json:"testCases"`
	CurrentStatus    string   `json:"currentStatus"`
	DefectCount      int      `json:"defectCount"`
	Severity         string   `json:"severity"`
	Notes            string   `json:"notes"`
	LastTestedDate   string   `json:"lastTestedDate"`
}

type testCase struct {
	TestCaseID       string   `json:"testCaseId"`
	FeatureID        string   `json:"featureId"`
	Title            string   `json:"title"`
	Category         string   `json:"category"`
	Preconditions    string   `json:"preconditions"`
	Steps            []string `json:"steps"`
	InputFixture     string   `json:"inputFixture"`
	ExpectedResult   string   `json:"expectedResult"`
	ExecutionType    string   `json:"executionType"`
	Status           string   `json:"status"`
	ActualResult     string   `json:"actualResult"`
	Evidence         string   `json:"evidence"`
	EvidenceLocation string   `json:"evidenceLocation"`
	LastExecutedDate string   `json:"lastExecutedDate"`
}

type defect struct {
	DefectID            string   `json:"defectId"`
	FeatureID           string   `json:"featureId"`
	Title               string   `json:"title"`
	Status              string   `json:"status"`
	Severity            string   `json:"severity"`
	ReproductionSteps   []string `json:"reproductionSteps"`
	ExpectedResult      string   `json:"expectedResult"`
	ActualResult        string   `json:"actualResult"`
	RootCauseHypothesis string   `json:"rootCauseHypothesis"`
	TestCaseID          string   `json:"testCaseId"`
	RootCause           string   `json:"rootCause"`
	Fix                 string   `json:"fix"`
	Verification        string   `json:"verification"`
	LastTestedDate      string   `json:"lastTestedDate"`
}

type execution struct {
	ExecutionID  string `json:"executionId"`
	TestCaseID   string `json:"testCaseId"`
	Command      string `json:"command"`
	Date         string `json:"date"`
	Status       string `json:"status"`
	ExitCode     *int   `json:"exitCode"`
	ActualResult string `json:"actualResult"`
	Evidence     string `json:"evidence"`
}

type iteration struct {
	IterationID      string   `json:"iterationId"`
	Date             string   `json:"date"`
	CoverageSummary  string   `json:"coverageSummary"`
	FeaturesTested   int      `json:"featuresTested"`
	DefectsFound     int      `json:"defectsFound"`
	SuspectedDefects int      `json:"suspectedDefects"`
	DefectsFixed     int      `json:"defectsFixed"`
	RemainingRisks   []string `json:"remainingRisks"`
	ConfidenceScore  int      `json:"confidenceScore"`
}

type blocker struct {
	BlockerID string `json:"blockerId"`
	Area      string `json:"area"`
	Detail    string `json:"detail"`
	Impact    string `json:"impact"`
	Status    string `json:"status"`
}

// categoryOrder controls the grouping order of feature sections. Feature IDs
// not matching a known category are appended afterwards in sorted order.
var categoryOrder = []string{"UI", "API", "CLI", "CORE", "CFG"}

var categoryTitles = map[string]string{
	"UI":   "Browser UI",
	"API":  "HTTP API",
	"CLI":  "Command line",
	"CORE": "Transformation core",
	"CFG":  "Configuration and deployment",
}

func main() {
	modelPath := flag.String("model", filepath.Join("qa", "source", "quality-model.json"), "path to the quality model JSON")
	outDir := flag.String("out", filepath.Join("qa", "workbook"), "directory to write the Markdown workbook into")
	flag.Parse()

	if err := run(*modelPath, *outDir); err != nil {
		fmt.Fprintln(os.Stderr, "qualitydoc:", err)
		os.Exit(1)
	}
}

func run(modelPath, outDir string) error {
	raw, err := os.ReadFile(modelPath)
	if err != nil {
		return err
	}

	var m model
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	files := map[string]string{
		"README.md":     renderSummary(&m),
		"features.md":   renderFeatures(&m),
		"test-cases.md": renderTestCases(&m),
		"defects.md":    renderDefects(&m),
		"executions.md": renderExecutions(&m),
		"iterations.md": renderIterations(&m),
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		path := filepath.Join(outDir, name)
		if err := os.WriteFile(path, []byte(files[name]), 0o644); err != nil {
			return err
		}
		fmt.Println("wrote", path)
	}
	return nil
}

// ---------- rendering ----------

func renderSummary(m *model) string {
	var b strings.Builder
	title(&b, m.Metadata.Product+" quality workbook")

	b.WriteString("Canonical source of truth for feature discovery, test design, execution evidence,\n")
	b.WriteString("and defect tracking.\n\n")
	b.WriteString("> **Generated file — do not edit by hand.**\n")
	b.WriteString("> Every page here is rendered from [`qa/source/quality-model.json`](../source/quality-model.json).\n")
	b.WriteString("> Edit the JSON, then regenerate with `go run ./qa/tools/qualitydoc`.\n\n")

	b.WriteString("| Page | Contents |\n|---|---|\n")
	b.WriteString("| [Features](features.md) | Every discovered feature with user story, expected behaviour, edge cases, validation rules, dependencies, and assumptions |\n")
	b.WriteString("| [Test cases](test-cases.md) | Designed scenarios per feature, with steps, fixtures, and expected results |\n")
	b.WriteString("| [Defects](defects.md) | Defect records with reproduction, root cause, fix, and verification |\n")
	b.WriteString("| [Executions](executions.md) | Commands actually run, with exit codes and observed results |\n")
	b.WriteString("| [Iterations](iterations.md) | Per-iteration coverage, risks, and confidence |\n\n")

	section(&b, "Snapshot")
	b.WriteString("| Metric | Value |\n|---|---:|\n")
	row(&b, "Features documented", fmt.Sprint(len(m.Features)))
	row(&b, "Test cases designed", fmt.Sprint(len(m.TestCases)))
	row(&b, "Test cases executed", fmt.Sprintf("%d (%s)", executedCount(m), pct(executedCount(m), len(m.TestCases))))
	row(&b, "Defects recorded", fmt.Sprint(len(m.Defects)))
	row(&b, "Defects open", fmt.Sprint(openDefects(m)))
	row(&b, "Execution records", fmt.Sprint(len(m.Executions)))
	row(&b, "Iterations completed", fmt.Sprint(len(m.Iterations)))
	b.WriteString("\n")
	b.WriteString("| Field | Value |\n|---|---|\n")
	row(&b, "Schema version", fmt.Sprint(m.Metadata.SchemaVersion))
	row(&b, "Created", cell(m.Metadata.CreatedDate))
	row(&b, "Last updated", cell(m.Metadata.UpdatedDate))
	row(&b, "Authoritative basis", cell(m.Metadata.AuthoritativeBasis))
	row(&b, "Workbook status", cell(m.Metadata.WorkbookStatus))
	row(&b, "Execution environment", cell(m.Metadata.ExecutionEnvironment))
	row(&b, "Phase", cell(m.Metadata.Phase))
	b.WriteString("\n")

	section(&b, "Test case status")
	tally(&b, "Status", collect(m.TestCases, func(t testCase) string { return t.Status }), len(m.TestCases))
	section(&b, "Test cases by execution type")
	tally(&b, "Execution type", collect(m.TestCases, func(t testCase) string { return t.ExecutionType }), len(m.TestCases))
	section(&b, "Test cases by scenario category")
	tally(&b, "Category", collect(m.TestCases, func(t testCase) string { return t.Category }), len(m.TestCases))

	section(&b, "Feature status")
	tally(&b, "Status", collect(m.Features, func(f feature) string { return f.CurrentStatus }), len(m.Features))

	section(&b, "Defects by severity")
	tally(&b, "Severity", collect(m.Defects, func(d defect) string { return d.Severity }), len(m.Defects))
	section(&b, "Defects by status")
	tally(&b, "Status", collect(m.Defects, func(d defect) string { return d.Status }), len(m.Defects))

	section(&b, "Inventory reconciliation")
	r := m.Metadata.Reconciliation
	b.WriteString("| Surface | Discovered | Feature records |\n|---|---:|---:|\n")
	b.WriteString(fmt.Sprintf("| Static screens | %d | %d |\n", r.Screens, r.UIFeatureRecords))
	b.WriteString(fmt.Sprintf("| HTTP routes | %d | %d |\n", r.APIRoutes, r.APIFeatureRecords))
	b.WriteString(fmt.Sprintf("| CLI modes | %d | %d |\n", r.CLIModes, r.CLIFeatureRecords))
	b.WriteString(fmt.Sprintf("| Transformation core | — | %d |\n", r.CoreFeatureRecords))
	b.WriteString(fmt.Sprintf("| Configuration and deployment | — | %d |\n", r.ConfigurationFeatureRecords))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("**Status:** %s &nbsp;·&nbsp; **Unmatched surfaces:** %d\n\n", cell(r.Status), r.UnmatchedSurfaces))
	b.WriteString(fmt.Sprintf("**Evidence:** %s\n\n", cell(r.Evidence)))

	if len(m.Blockers) > 0 {
		section(&b, "Blockers")
		b.WriteString("| ID | Area | Status | Detail | Impact |\n|---|---|---|---|---|\n")
		blockers := append([]blocker(nil), m.Blockers...)
		sort.Slice(blockers, func(i, j int) bool { return blockers[i].BlockerID < blockers[j].BlockerID })
		for _, bl := range blockers {
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				cell(bl.BlockerID), cell(bl.Area), cell(bl.Status), cell(bl.Detail), cell(bl.Impact)))
		}
		b.WriteString("\n")
	}

	section(&b, "Controlled vocabularies")
	b.WriteString("Values in use across the model. Keep new records within these sets so the\n")
	b.WriteString("rollups above stay meaningful.\n\n")
	vocab(&b, "Feature status", collect(m.Features, func(f feature) string { return f.CurrentStatus }))
	vocab(&b, "Feature severity", collect(m.Features, func(f feature) string { return f.Severity }))
	vocab(&b, "Test case status", collect(m.TestCases, func(t testCase) string { return t.Status }))
	vocab(&b, "Execution type", collect(m.TestCases, func(t testCase) string { return t.ExecutionType }))
	vocab(&b, "Scenario category", collect(m.TestCases, func(t testCase) string { return t.Category }))
	vocab(&b, "Defect status", collect(m.Defects, func(d defect) string { return d.Status }))
	vocab(&b, "Defect severity", collect(m.Defects, func(d defect) string { return d.Severity }))

	return b.String()
}

func renderFeatures(m *model) string {
	var b strings.Builder
	title(&b, "Features")
	generatedNote(&b)

	features := append([]feature(nil), m.Features...)
	sort.Slice(features, func(i, j int) bool { return features[i].FeatureID < features[j].FeatureID })

	defectsByFeature := map[string][]defect{}
	for _, d := range m.Defects {
		defectsByFeature[d.FeatureID] = append(defectsByFeature[d.FeatureID], d)
	}

	section(&b, "Index")
	b.WriteString("| Feature ID | Feature name | Current status | Test cases | Defects | Severity | Last tested |\n")
	b.WriteString("|---|---|---|---:|---:|---|---|\n")
	for _, f := range features {
		b.WriteString(fmt.Sprintf("| [%s](#%s) | %s | %s | %d | %d | %s | %s |\n",
			cell(f.FeatureID), anchorID(f.FeatureID), cell(f.FeatureName), cell(f.CurrentStatus),
			len(f.TestCases), f.DefectCount, cell(f.Severity), cell(f.LastTestedDate)))
	}
	b.WriteString("\n")

	for _, category := range orderedCategories(features) {
		section(&b, categoryTitle(category))
		for _, f := range features {
			if featureCategory(f.FeatureID) != category {
				continue
			}
			b.WriteString(fmt.Sprintf("<a id=\"%s\"></a>\n\n", anchorID(f.FeatureID)))
			b.WriteString(fmt.Sprintf("### %s — %s\n\n", f.FeatureID, f.FeatureName))

			b.WriteString(fmt.Sprintf("> %s\n\n", oneLine(f.UserStory)))

			field(&b, "Expected behaviour", f.ExpectedBehavior)
			field(&b, "Edge cases", f.EdgeCases)
			field(&b, "Validation rules", f.ValidationRules)
			field(&b, "Dependencies", f.Dependencies)
			field(&b, "Assumptions", f.Assumptions)
			field(&b, "Notes", f.Notes)

			b.WriteString(fmt.Sprintf("**Status:** %s &nbsp;·&nbsp; **Severity:** %s &nbsp;·&nbsp; **Defect count:** %d &nbsp;·&nbsp; **Last tested:** %s\n\n",
				cell(f.CurrentStatus), cell(f.Severity), f.DefectCount, cell(f.LastTestedDate)))

			if len(f.TestCases) > 0 {
				ids := append([]string(nil), f.TestCases...)
				sort.Strings(ids)
				b.WriteString(fmt.Sprintf("**Test cases (%d):** [%s … %s](test-cases.md#%s)\n\n",
					len(ids), ids[0], ids[len(ids)-1], anchorID(f.FeatureID)))
			}

			if ds := defectsByFeature[f.FeatureID]; len(ds) > 0 {
				sort.Slice(ds, func(i, j int) bool { return ds[i].DefectID < ds[j].DefectID })
				links := make([]string, 0, len(ds))
				for _, d := range ds {
					links = append(links, fmt.Sprintf("[%s](defects.md#%s) (%s, %s)", d.DefectID, anchorID(d.DefectID), d.Severity, d.Status))
				}
				b.WriteString("**Defects:** " + strings.Join(links, ", ") + "\n\n")
			}
		}
	}

	return b.String()
}

func renderTestCases(m *model) string {
	var b strings.Builder
	title(&b, "Test cases")
	generatedNote(&b)

	cases := append([]testCase(nil), m.TestCases...)
	sort.Slice(cases, func(i, j int) bool { return cases[i].TestCaseID < cases[j].TestCaseID })

	// Preconditions are frequently identical boilerplate. When every case
	// shares one value, state it once instead of repeating it in 300+ rows;
	// otherwise each case carries its own column so nothing is lost.
	shared, uniform := sharedPreconditions(cases)
	if uniform {
		section(&b, "Shared preconditions")
		b.WriteString(fmt.Sprintf("Every case below runs under the same preconditions:\n\n> %s\n\n", oneLine(shared)))
	}

	names := map[string]string{}
	for _, f := range m.Features {
		names[f.FeatureID] = f.FeatureName
	}

	section(&b, "Index")
	b.WriteString("| Feature | Name | Cases | Executed |\n|---|---|---:|---:|\n")
	for _, id := range featureIDsOf(cases) {
		total, done := 0, 0
		for _, t := range cases {
			if t.FeatureID != id {
				continue
			}
			total++
			if isExecuted(t.Status) {
				done++
			}
		}
		b.WriteString(fmt.Sprintf("| [%s](#%s) | %s | %d | %d |\n", id, anchorID(id), cell(names[id]), total, done))
	}
	b.WriteString("\n")

	for _, id := range featureIDsOf(cases) {
		b.WriteString(fmt.Sprintf("<a id=\"%s\"></a>\n\n", anchorID(id)))
		b.WriteString(fmt.Sprintf("## %s — %s\n\n", id, names[id]))

		header := "| Test case ID | Category | Title | Steps | Input fixture | Expected result | Type | Status | Actual result | Evidence | Last executed |"
		divider := "|---|---|---|---|---|---|---|---|---|---|---|"
		if !uniform {
			header = "| Test case ID | Category | Title | Preconditions | Steps | Input fixture | Expected result | Type | Status | Actual result | Evidence | Last executed |"
			divider = "|---|---|---|---|---|---|---|---|---|---|---|---|"
		}
		b.WriteString(header + "\n" + divider + "\n")

		for _, t := range cases {
			if t.FeatureID != id {
				continue
			}
			pre := ""
			if !uniform {
				pre = cell(t.Preconditions) + " | "
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s%s | %s | %s | %s | %s | %s | %s | %s |\n",
				cell(t.TestCaseID), cell(t.Category), cell(t.Title), pre, numbered(t.Steps),
				cell(t.InputFixture), cell(t.ExpectedResult), cell(t.ExecutionType), cell(t.Status),
				cell(t.ActualResult), cell(firstNonEmpty(t.Evidence, t.EvidenceLocation)), cell(t.LastExecutedDate)))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func renderDefects(m *model) string {
	var b strings.Builder
	title(&b, "Defects")
	generatedNote(&b)

	defects := append([]defect(nil), m.Defects...)
	sort.Slice(defects, func(i, j int) bool { return defects[i].DefectID < defects[j].DefectID })

	names := map[string]string{}
	for _, f := range m.Features {
		names[f.FeatureID] = f.FeatureName
	}

	section(&b, "Index")
	b.WriteString("| Defect ID | Severity | Status | Feature | Title | Test case | Last tested |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, d := range defects {
		b.WriteString(fmt.Sprintf("| [%s](#%s) | %s | %s | [%s](features.md#%s) | %s | %s | %s |\n",
			cell(d.DefectID), anchorID(d.DefectID), cell(d.Severity), cell(d.Status),
			cell(d.FeatureID), anchorID(d.FeatureID), cell(d.Title), cell(d.TestCaseID), cell(d.LastTestedDate)))
	}
	b.WriteString("\n")

	section(&b, "Records")
	for _, d := range defects {
		b.WriteString(fmt.Sprintf("<a id=\"%s\"></a>\n\n", anchorID(d.DefectID)))
		b.WriteString(fmt.Sprintf("### %s — %s\n\n", d.DefectID, d.Title))
		b.WriteString(fmt.Sprintf("**Severity:** %s &nbsp;·&nbsp; **Status:** %s &nbsp;·&nbsp; **Feature:** [%s](features.md#%s) %s &nbsp;·&nbsp; **Last tested:** %s\n\n",
			cell(d.Severity), cell(d.Status), cell(d.FeatureID), anchorID(d.FeatureID), cell(names[d.FeatureID]), cell(d.LastTestedDate)))

		if len(d.ReproductionSteps) > 0 {
			b.WriteString("**Reproduction steps**\n\n")
			for i, s := range d.ReproductionSteps {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, strings.TrimSpace(s)))
			}
			b.WriteString("\n")
		}

		field(&b, "Expected result", d.ExpectedResult)
		field(&b, "Actual result", d.ActualResult)
		field(&b, "Root cause hypothesis", d.RootCauseHypothesis)
		field(&b, "Root cause", d.RootCause)
		field(&b, "Fix", d.Fix)
		field(&b, "Verification", d.Verification)

		if d.TestCaseID != "" {
			b.WriteString(fmt.Sprintf("**Regression test:** `%s`\n\n", d.TestCaseID))
		}
		b.WriteString("---\n\n")
	}

	return b.String()
}

func renderExecutions(m *model) string {
	var b strings.Builder
	title(&b, "Executions")
	generatedNote(&b)

	b.WriteString("Commands actually run against the worktree, with the exit code and result\n")
	b.WriteString("observed at the time. A test case may only be marked `Passed` when an\n")
	b.WriteString("execution record here backs it.\n\n")

	executions := append([]execution(nil), m.Executions...)
	sort.Slice(executions, func(i, j int) bool { return executions[i].ExecutionID < executions[j].ExecutionID })

	b.WriteString("| Execution ID | Date | Test case | Command | Status | Exit code | Actual result | Evidence |\n")
	b.WriteString("|---|---|---|---|---|---:|---|---|\n")
	for _, e := range executions {
		exit := "—"
		if e.ExitCode != nil {
			exit = fmt.Sprint(*e.ExitCode)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` | %s | %s | %s | %s |\n",
			cell(e.ExecutionID), cell(e.Date), cell(e.TestCaseID), inlineCode(e.Command),
			cell(e.Status), exit, cell(e.ActualResult), cell(e.Evidence)))
	}
	b.WriteString("\n")

	return b.String()
}

func renderIterations(m *model) string {
	var b strings.Builder
	title(&b, "Iterations")
	generatedNote(&b)

	iterations := append([]iteration(nil), m.Iterations...)
	sort.Slice(iterations, func(i, j int) bool { return iterations[i].IterationID < iterations[j].IterationID })

	for _, it := range iterations {
		b.WriteString(fmt.Sprintf("## %s — %s\n\n", it.IterationID, it.Date))
		b.WriteString("| Metric | Value |\n|---|---:|\n")
		row(&b, "Features tested", fmt.Sprint(it.FeaturesTested))
		row(&b, "Defects found", fmt.Sprint(it.DefectsFound))
		row(&b, "Defects fixed", fmt.Sprint(it.DefectsFixed))
		row(&b, "Suspected defects outstanding", fmt.Sprint(it.SuspectedDefects))
		row(&b, "Confidence score", fmt.Sprintf("%d%%", it.ConfidenceScore))
		b.WriteString("\n")

		field(&b, "Coverage summary", it.CoverageSummary)

		if len(it.RemainingRisks) > 0 {
			b.WriteString("**Remaining risks**\n\n")
			for _, r := range it.RemainingRisks {
				b.WriteString("- " + strings.TrimSpace(r) + "\n")
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// ---------- helpers ----------

func title(b *strings.Builder, s string) {
	b.WriteString("# " + s + "\n\n")
}

func section(b *strings.Builder, s string) {
	b.WriteString("## " + s + "\n\n")
}

func row(b *strings.Builder, label, value string) {
	b.WriteString(fmt.Sprintf("| %s | %s |\n", label, value))
}

func field(b *strings.Builder, label, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	b.WriteString(fmt.Sprintf("**%s**\n\n%s\n\n", label, oneLine(value)))
}

func generatedNote(b *strings.Builder) {
	b.WriteString("> Generated from [`qa/source/quality-model.json`](../source/quality-model.json)\n")
	b.WriteString("> by `go run ./qa/tools/qualitydoc`. Do not edit by hand.\n")
	b.WriteString("> [Back to the coverage summary](README.md).\n\n")
}

// cell renders a value for use inside a Markdown table cell.
func cell(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "<br>")
}

func inlineCode(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	s = strings.ReplaceAll(s, "|", "\\|")
	return strings.ReplaceAll(s, "`", "'")
}

// oneLine collapses internal newlines so a value stays on a single paragraph.
func oneLine(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "—"
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", " ")
}

func numbered(items []string) string {
	if len(items) == 0 {
		return "—"
	}
	parts := make([]string, 0, len(items))
	for i, s := range items {
		parts = append(parts, fmt.Sprintf("%d. %s", i+1, strings.ReplaceAll(strings.TrimSpace(s), "|", "\\|")))
	}
	return strings.Join(parts, "<br>")
}

func anchorID(id string) string {
	return strings.ToLower(id)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func isExecuted(status string) bool {
	return !strings.EqualFold(strings.TrimSpace(status), "Not Run")
}

func executedCount(m *model) int {
	n := 0
	for _, t := range m.TestCases {
		if isExecuted(t.Status) {
			n++
		}
	}
	return n
}

func openDefects(m *model) int {
	n := 0
	for _, d := range m.Defects {
		switch strings.ToLower(strings.TrimSpace(d.Status)) {
		case "fixed", "waived", "closed":
		default:
			n++
		}
	}
	return n
}

func pct(part, total int) string {
	if total == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", float64(part)*100/float64(total))
}

func collect[T any](items []T, key func(T) string) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, strings.TrimSpace(key(it)))
	}
	return out
}

// tally writes a count table sorted by descending count, then name.
func tally(b *strings.Builder, label string, values []string, total int) {
	counts := map[string]int{}
	for _, v := range values {
		if v == "" {
			v = "(unset)"
		}
		counts[v]++
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})

	b.WriteString(fmt.Sprintf("| %s | Count | Share |\n|---|---:|---:|\n", label))
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("| %s | %d | %s |\n", cell(k), counts[k], pct(counts[k], total)))
	}
	b.WriteString("\n")
}

func vocab(b *strings.Builder, label string, values []string) {
	seen := map[string]bool{}
	uniq := make([]string, 0)
	for _, v := range values {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		uniq = append(uniq, v)
	}
	if len(uniq) == 0 {
		return
	}
	sort.Strings(uniq)
	for i, v := range uniq {
		uniq[i] = "`" + v + "`"
	}
	b.WriteString(fmt.Sprintf("- **%s:** %s\n", label, strings.Join(uniq, ", ")))
}

func featureCategory(id string) string {
	parts := strings.Split(id, "-")
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

func categoryTitle(category string) string {
	if t, ok := categoryTitles[category]; ok {
		return fmt.Sprintf("%s (`FEAT-%s-*`)", t, category)
	}
	return fmt.Sprintf("`FEAT-%s-*`", category)
}

func orderedCategories(features []feature) []string {
	present := map[string]bool{}
	for _, f := range features {
		present[featureCategory(f.FeatureID)] = true
	}
	out := make([]string, 0, len(present))
	for _, c := range categoryOrder {
		if present[c] {
			out = append(out, c)
			delete(present, c)
		}
	}
	rest := make([]string, 0, len(present))
	for c := range present {
		rest = append(rest, c)
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// featureIDsOf returns the distinct feature IDs referenced by cases, sorted.
func featureIDsOf(cases []testCase) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	for _, t := range cases {
		if seen[t.FeatureID] {
			continue
		}
		seen[t.FeatureID] = true
		out = append(out, t.FeatureID)
	}
	sort.Strings(out)
	return out
}

func sharedPreconditions(cases []testCase) (string, bool) {
	if len(cases) == 0 {
		return "", false
	}
	first := strings.TrimSpace(cases[0].Preconditions)
	if first == "" {
		return "", false
	}
	for _, t := range cases {
		if strings.TrimSpace(t.Preconditions) != first {
			return "", false
		}
	}
	return first, true
}
