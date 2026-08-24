package nodejs

import (
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

	for _, want := range []string{
		"<P />",
		"<SectionCard />",
		"<SectionCard2 />",
		"<P2 />",
	} {
		if !strings.Contains(mainComponent, want) {
			t.Errorf("MainComponent missing %q; got %s", want, mainComponent)
		}
	}
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
