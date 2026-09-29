package nodejs

import (
	"regexp"
	"strings"
	"testing"

	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"
)

func TestGenerateProjectLoadsOriginalJavaScript(t *testing.T) {
	project, err := GenerateProject(&ProjectConfig{
		ProjectName: "script-fixture",
		HTML:        `<html><body><main>Fixture</main></body></html>`,
		JS:          `window.inlineReady = true`,
		ExternalJS: []fetcher.FetchedResource{
			{
				Filename: "vendor.js",
				Content:  `window.vendorReady = true`,
			},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject() error = %v", err)
	}

	for filename, wantContent := range map[string]string{
		"public/scripts/main.js":            `window.inlineReady = true`,
		"public/scripts/external/vendor.js": `window.vendorReady = true`,
	} {
		if got, ok := project.Files[filename]; !ok {
			t.Errorf("GenerateProject() missing %q", filename)
		} else if got != wantContent {
			t.Errorf("GenerateProject() %q = %q, want %q", filename, got, wantContent)
		}
	}

	mainTSX := project.Files["src/main.tsx"]
	for _, want := range []string{
		`'/scripts/main.js'`,
		`'/scripts/external/vendor.js'`,
		`loadOriginalScripts`,
	} {
		if !strings.Contains(mainTSX, want) {
			t.Errorf("src/main.tsx missing %q; got %s", want, mainTSX)
		}
	}
}

func TestGenerateProjectHydratesASourceMatchingBodyWithoutLayoutWrappers(t *testing.T) {
	project, err := GenerateProject(&ProjectConfig{
		ProjectName: "document-shell",
		HTML: `<!doctype html>
			<html lang="en" data-wf-domain="example.test" data-wf-page="page-id">
			<head><title>Source title</title></head>
			<body class="site-body" data-wf-site="site-id"><main id="content">Fixture</main></body>
			</html>`,
	})
	if err != nil {
		t.Fatalf("GenerateProject() error = %v", err)
	}

	indexHTML := project.Files["src/index.html"]
	for _, want := range []string{
		`<html lang="en" data-wf-domain="example.test" data-wf-page="page-id">`,
		`<body class="site-body" data-wf-site="site-id">`,
		`<main id="content">Fixture</main>`,
	} {
		if !strings.Contains(indexHTML, want) {
			t.Errorf("src/index.html missing %q; got:\n%s", want, indexHTML)
		}
	}
	if strings.Contains(indexHTML, `id="root"`) {
		t.Errorf("src/index.html contains layout-changing #root wrapper; got:\n%s", indexHTML)
	}
	if script, body := strings.Index(indexHTML, `src="/main.tsx"`), strings.Index(indexHTML, "<body"); script == -1 || script > body {
		t.Errorf("src/index.html bootstrap must stay in <head> outside the hydrated body; got:\n%s", indexHTML)
	}

	appTSX := project.Files["src/App.tsx"]
	if strings.Contains(appTSX, `className="App"`) || strings.Contains(appTSX, "import React") {
		t.Errorf("src/App.tsx contains an unnecessary React binding or DOM wrapper; got:\n%s", appTSX)
	}
	mainTSX := project.Files["src/main.tsx"]
	if !strings.Contains(mainTSX, `hydrateRoot(document.body, <App />)`) || strings.Contains(mainTSX, "createRoot") {
		t.Errorf("src/main.tsx does not hydrate the source-matching body shell; got:\n%s", mainTSX)
	}
}

func TestGenerateProjectLoadsScriptsOnceInSourceOrderAfterHydrationCommit(t *testing.T) {
	project, err := GenerateProject(&ProjectConfig{
		ProjectName: "ordered-scripts",
		HTML: `<html><head>
			<script src="/scripts/external/head.js" data-bootstrap="true"></script>
			</head><body><main>Fixture</main>
			<script src="/scripts/main.js"></script>
			<script type="module" src="/scripts/external/tail.js"></script>
			</body></html>`,
		JS: `window.order = ['inline']`,
		ExternalJS: []fetcher.FetchedResource{
			{Filename: "tail.js", Content: `window.order.push('tail')`},
			{Filename: "head.js", Content: `window.order = ['head']`},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject() error = %v", err)
	}

	indexHTML := project.Files["src/index.html"]
	for _, original := range []string{"head.js", "tail.js", "/scripts/main.js"} {
		if strings.Contains(indexHTML, original) {
			t.Errorf("src/index.html executes original script %q before hydration; got:\n%s", original, indexHTML)
		}
	}

	mainTSX := project.Files["src/main.tsx"]
	head := strings.Index(mainTSX, `src: '/scripts/external/head.js'`)
	inline := strings.Index(mainTSX, `src: '/scripts/main.js'`)
	tail := strings.Index(mainTSX, `src: '/scripts/external/tail.js'`)
	if head == -1 || inline == -1 || tail == -1 || !(head < inline && inline < tail) {
		t.Errorf("src/main.tsx script descriptors are not in source order; got:\n%s", mainTSX)
	}
	for _, want := range []string{
		`parent: 'head'`,
		`'data-bootstrap': 'true'`,
		`useEffect(() => {`,
		`void loadOriginalScripts()`,
		`<HydrationComplete />`,
	} {
		if !strings.Contains(mainTSX, want) {
			t.Errorf("src/main.tsx missing %q; got:\n%s", want, mainTSX)
		}
	}
	if effect, hydrate := strings.Index(mainTSX, `useEffect(() => {`), strings.Index(mainTSX, `hydrateRoot(`); effect == -1 || hydrate == -1 || effect > hydrate {
		t.Errorf("script loading is not registered as a hydration commit effect; got:\n%s", mainTSX)
	}
	if strings.Count(mainTSX, `src: '/scripts/external/head.js'`) != 1 ||
		strings.Count(mainTSX, `src: '/scripts/main.js'`) != 1 ||
		strings.Count(mainTSX, `src: '/scripts/external/tail.js'`) != 1 {
		t.Errorf("an original script was scheduled more than once; got:\n%s", mainTSX)
	}
}

func TestGenerateProjectKeepsIndividualInlineScriptsInterleavedWithExternalScripts(t *testing.T) {
	project, err := GenerateProject(&ProjectConfig{
		ProjectName: "interleaved-scripts",
		HTML: `<html><head><script src="inline/script-1.js"></script></head><body>
			<main>Fixture</main>
			<script src="external/js/vendor.js"></script>
			<script src="inline/script-2.js"></script>
			</body></html>`,
		InlineJS: []extractor.InlineResource{
			{Path: "inline/script-1.js", Content: `window.order = ['one']`},
			{Path: "inline/script-2.js", Content: `window.order.push('two')`},
		},
		ExternalJS: []fetcher.FetchedResource{
			{Filename: "vendor.js", Content: `window.order.push('vendor')`},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject() error = %v", err)
	}

	for path, want := range map[string]string{
		"public/scripts/inline/script-1.js": `window.order = ['one']`,
		"public/scripts/inline/script-2.js": `window.order.push('two')`,
	} {
		if got := project.Files[path]; got != want {
			t.Errorf("generated %s = %q, want %q", path, got, want)
		}
	}

	mainTSX := project.Files["src/main.tsx"]
	one := strings.Index(mainTSX, `src: '/scripts/inline/script-1.js'`)
	vendor := strings.Index(mainTSX, `src: '/scripts/external/vendor.js'`)
	two := strings.Index(mainTSX, `src: '/scripts/inline/script-2.js'`)
	if one == -1 || vendor == -1 || two == -1 || !(one < vendor && vendor < two) {
		t.Errorf("inline and external scripts are not preserved in document order; got:\n%s", mainTSX)
	}
}

func TestGenerateTSXViewsPreservesTopLevelContentAndDuplicates(t *testing.T) {
	sectionFiles, mainComponent, _, err := generateTSXViews(`
		<html><body><div id="app">
			<p>Intro outside semantic sections</p>
			<section class="card"><p>Repeated card</p></section>
			<section class="card"><p>Repeated card</p></section>
			<p>Outro outside semantic sections</p>
		</div></body></html>`, "", "", nil, nil)
	if err != nil {
		t.Fatalf("generateTSXViews() error = %v", err)
	}

	if len(sectionFiles) != 4 {
		t.Errorf("generateTSXViews() generated %d section files, want 4", len(sectionFiles))
	}

	// The regression this guards (DEF-017) is that content is dropped or that
	// identical siblings collapse into one component. Assert that shape rather
	// than specific names, which decomposition_test.go owns and which improve
	// independently of this behaviour.
	refs := componentRefs(mainComponent)
	if len(refs) != 4 {
		t.Errorf("MainComponent references %d components, want 4; got %s", len(refs), mainComponent)
	}

	seen := map[string]bool{}
	for _, ref := range refs {
		if seen[ref] {
			t.Errorf("identical siblings collapsed into one component %q; got %s", ref, mainComponent)
		}
		seen[ref] = true
	}

	for _, ref := range refs {
		if _, ok := sectionFiles["src/components/"+ref+".tsx"]; !ok {
			t.Errorf("MainComponent references %q with no matching section file", ref)
		}
	}
}

func TestFlatMainComponentFallbackDoesNotImportUnusedReactBinding(t *testing.T) {
	got := generateMainComponentTSX([]tsxComponent{{Name: "Hero"}})

	if strings.Contains(got, "import React") {
		t.Fatalf("flat MainComponent fallback contains an unused React import; got:\n%s", got)
	}
}

var componentRefPattern = regexp.MustCompile(`<([A-Z][A-Za-z0-9]*) />`)

// componentRefs returns the component names rendered by MainComponent.
func componentRefs(mainComponent string) []string {
	matches := componentRefPattern.FindAllStringSubmatch(mainComponent, -1)
	refs := make([]string, 0, len(matches))
	for _, m := range matches {
		refs = append(refs, m[1])
	}
	return refs
}

func TestGeneratedProjectsAdaptLocalizedCSSForPublicAssets(t *testing.T) {
	externalCSS := []fetcher.FetchedResource{
		{
			Filename: "site.css",
			Content: `.hero { background: url("../../assets/hero.png"); }
.icon { mask-image: url('../assets/icon.svg#mark'); }
.font { src: url(../../../assets/font.woff2?v=1); }
.data { background: url(data:image/svg+xml;base64,PHN2Zz4=); }
.remote { background: url("https://cdn.example/image.png"); }`,
		},
	}

	tsxProject, err := GenerateProject(&ProjectConfig{
		ProjectName: "tsx-assets",
		HTML:        `<html><body><main>Fixture</main></body></html>`,
		ExternalCSS: externalCSS,
	})
	if err != nil {
		t.Fatalf("GenerateProject() error = %v", err)
	}

	ejsProject, err := GenerateEJSProject(&EJSProjectConfig{
		ProjectName: "ejs-assets",
		HTML:        `<html><body><main>Fixture</main></body></html>`,
		ExternalCSS: externalCSS,
	})
	if err != nil {
		t.Fatalf("GenerateEJSProject() error = %v", err)
	}

	for filename, content := range map[string]string{
		"src/styles/external/site.css": tsxProject.Files["src/styles/external/site.css"],
		"public/external/css/site.css": ejsProject.Files["public/external/css/site.css"],
	} {
		for _, want := range []string{
			`url("/assets/hero.png")`,
			`url('/assets/icon.svg#mark')`,
			`url(/assets/font.woff2?v=1)`,
			`url(data:image/svg+xml;base64,PHN2Zz4=)`,
			`url("https://cdn.example/image.png")`,
		} {
			if !strings.Contains(content, want) {
				t.Errorf("generated %s missing rebased CSS URL %q; got %s", filename, want, content)
			}
		}
	}
}

func TestGenerateProjectDoesNotDoubleLoadRecursivelyImportedStylesheets(t *testing.T) {
	project, err := GenerateProject(&ProjectConfig{
		ProjectName: "css-imports",
		HTML:        `<html><body><main>Fixture</main></body></html>`,
		ExternalCSS: []fetcher.FetchedResource{
			{Filename: "site.css", Type: "css", Content: `@import "./theme.css";`},
			{Filename: "theme.css", Type: "css-import", Content: `body { color: navy; }`},
		},
	})
	if err != nil {
		t.Fatalf("GenerateProject() error = %v", err)
	}

	if _, ok := project.Files["src/styles/external/theme.css"]; !ok {
		t.Fatal("recursively imported stylesheet file was not emitted")
	}
	mainTSX := project.Files["src/main.tsx"]
	if !strings.Contains(mainTSX, `import './styles/external/site.css'`) {
		t.Errorf("src/main.tsx does not import the top-level stylesheet; got:\n%s", mainTSX)
	}
	if strings.Contains(mainTSX, `import './styles/external/theme.css'`) {
		t.Errorf("src/main.tsx directly imports a stylesheet already reached through @import; got:\n%s", mainTSX)
	}
}
