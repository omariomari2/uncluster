package nodejs

import (
	"sort"
	"strings"
	"testing"

	"github.com/omariomari2/uncluster/internal/htmlutil"

	"golang.org/x/net/html"
)

func componentNamesFor(t *testing.T, markup string) []string {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(markup))
	if err != nil {
		t.Fatalf("html.Parse() error = %v", err)
	}
	body := htmlutil.FindElement(doc, "body")
	if body == nil {
		t.Fatal("no body in fixture")
	}

	sections := filterComponentCandidates(contentChildren(selectComponentRoot(body)))

	used := map[string]int{}
	names := make([]string, 0, len(sections))
	for i, n := range sections {
		names = append(names, toPascalCase(buildComponentName(n, i, used)))
	}
	sort.Strings(names)
	return names
}

// A component name should describe the region, not restate the tag. The old
// naming produced FooterSiteFooter and HeaderSiteHeader.
func TestComponentNamesDoNotRestateTheTag(t *testing.T) {
	got := componentNamesFor(t, `<body>
		<header class="site-header">head</header>
		<footer class="site-footer">foot</footer>
	</body>`)

	want := []string{"SiteFooter", "SiteHeader"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("component names = %v, want %v", got, want)
	}
}

func TestComponentNameSourcePrecedence(t *testing.T) {
	cases := []struct {
		name   string
		markup string
		want   string
	}{
		{"id wins over class", `<section id="hero" class="banner">x</section>`, "Hero"},
		{"class when no id", `<div class="features">x</div>`, "Features"},
		{"BEM modifier reduced to block", `<article class="post--featured">x</article>`, "Post"},
		{"first class of several", `<article class="post post--featured">x</article>`, "Post"},
		{"semantic tag when anonymous", `<aside>x</aside>`, "Aside"},
		{"utility classes skipped", `<section class="mt-4 pricing">x</section>`, "Pricing"},
		{"anonymous div falls back to block", `<div>x</div>`, "Block1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Two siblings so the root is not descended into.
			got := componentNamesFor(t, "<body>"+tc.markup+"<footer>f</footer></body>")
			found := false
			for _, n := range got {
				if n == tc.want {
					found = true
				}
			}
			if !found {
				t.Errorf("component names = %v, want one to be %q", got, tc.want)
			}
		})
	}
}

// A page whose content sits inside <main> must not collapse into a single
// Main component holding everything.
func TestLayoutContainersAreExpandedIntoRegions(t *testing.T) {
	got := componentNamesFor(t, `<body>
		<div id="app">
			<header class="site-header">h</header>
			<main>
				<section id="hero">a</section>
				<div class="features">b</div>
				<aside class="sidebar">c</aside>
			</main>
			<footer class="site-footer">f</footer>
		</div>
	</body>`)

	for _, want := range []string{"Hero", "Features", "Sidebar", "SiteHeader", "SiteFooter"} {
		found := false
		for _, n := range got {
			if n == want {
				found = true
			}
		}
		if !found {
			t.Errorf("expected %q among components; got %v", want, got)
		}
	}
	for _, unwanted := range got {
		if unwanted == "Main" {
			t.Errorf("<main> became a component instead of expanding; got %v", got)
		}
	}
}

// An element carrying its own identity is a region, not a wrapper, even when
// it holds several children.
func TestNamedContainersAreNotExpanded(t *testing.T) {
	got := componentNamesFor(t, `<body>
		<div class="features"><div class="feature">a</div><div class="feature">b</div></div>
		<footer>f</footer>
	</body>`)

	for _, n := range got {
		if n == "Feature" || n == "Feature1" {
			t.Errorf("named container was expanded into its children; got %v", got)
		}
	}
}

func TestComponentNamesAreUnique(t *testing.T) {
	got := componentNamesFor(t, `<body>
		<section class="panel">a</section>
		<section class="panel">b</section>
		<section class="panel">c</section>
	</body>`)

	seen := map[string]bool{}
	for _, n := range got {
		if seen[n] {
			t.Errorf("duplicate component name %q in %v", n, got)
		}
		seen[n] = true
	}
	if len(got) != 3 {
		t.Errorf("expected 3 components, got %v", got)
	}
}

// MainComponent used to be a flat fragment of components, which silently
// dropped everything around and between them: the page wrapper that
// selectComponentRoot descended through, and the landmarks that
// expandLayoutContainers expanded. Composing in place keeps that structure.
func TestMainComponentPreservesSurroundingMarkup(t *testing.T) {
	_, main, _, err := generateTSXViews(`
		<html><body><div class="container" id="top">
			<header class="site-header">h</header>
			<main>
				<section id="hero">a</section>
				<aside class="sidebar">c</aside>
			</main>
			<footer class="site-footer">f</footer>
		</div></body></html>`, "", "", nil, nil)
	if err != nil {
		t.Fatalf("generateTSXViews() error = %v", err)
	}

	for _, want := range []string{
		// The wrapper selectComponentRoot descends through.
		`<div className="container" id="top">`,
		// The landmark whose children became components.
		"<main>",
		// Components sit where their markup was.
		"<SiteHeader />",
		"<Hero />",
		"<Sidebar />",
		"<SiteFooter />",
	} {
		if !strings.Contains(main, want) {
			t.Errorf("MainComponent missing %q; got:\n%s", want, main)
		}
	}

	// No placeholder may survive into the output.
	if strings.Contains(main, componentMarkerTag) {
		t.Errorf("component placeholder leaked into output; got:\n%s", main)
	}
}
