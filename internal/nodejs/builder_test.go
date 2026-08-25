package nodejs

import (
	"regexp"
	"strings"
	"testing"

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
			Content:  `.hero { background: url("../../assets/hero.png"); }`,
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
		if !strings.Contains(content, `url("/assets/hero.png")`) {
			t.Errorf("generated %s did not target the public asset root; got %s", filename, content)
		}
	}
}
