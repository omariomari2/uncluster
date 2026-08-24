package scraper

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/omariomari2/uncluster/internal/extractor"
	"golang.org/x/net/html"
)

func parseScraperDocument(t *testing.T, source string) *html.Node {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("html.Parse() error = %v", err)
	}
	return doc
}

func renderScraperDocument(t *testing.T, doc *html.Node) string {
	t.Helper()

	var output bytes.Buffer
	if err := html.Render(&output, doc); err != nil {
		t.Fatalf("html.Render() error = %v", err)
	}
	return output.String()
}

func TestRewriteHTMLPathsUsesPortableRelativePaths(t *testing.T) {
	doc := parseScraperDocument(t, `
		<html><head><link rel="stylesheet" href="/site.css"></head>
		<body><img src="/images/logo.png"></body></html>`)
	base, err := url.Parse("https://example.com/pages/home")
	if err != nil {
		t.Fatal(err)
	}

	rewriteHTMLPaths(doc, map[string]string{
		"https://example.com/site.css":        "external/css/site.css",
		"https://example.com/images/logo.png": "assets/logo.png",
	}, base)

	got := renderScraperDocument(t, doc)
	for _, want := range []string{
		`href="external/css/site.css"`,
		`src="assets/logo.png"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rewriteHTMLPaths() output missing %q; got %s", want, got)
		}
	}
}

func TestRewriteHTMLPathsRewritesEverySrcsetCandidate(t *testing.T) {
	doc := parseScraperDocument(t, `
		<img srcset="/images/small.png 1x, /images/large.png 2x">`)
	base, err := url.Parse("https://example.com/")
	if err != nil {
		t.Fatal(err)
	}

	rewriteHTMLPaths(doc, map[string]string{
		"https://example.com/images/small.png": "assets/small.png",
		"https://example.com/images/large.png": "assets/large.png",
	}, base)

	got := renderScraperDocument(t, doc)
	want := `srcset="assets/small.png 1x, assets/large.png 2x"`
	if !strings.Contains(got, want) {
		t.Fatalf("rewriteHTMLPaths() srcset missing %q; got %s", want, got)
	}
}

func TestExtractInlineResourcesPreservesDataScripts(t *testing.T) {
	doc := parseScraperDocument(t, `
		<script type="application/ld+json">{"name":"Uncluster"}</script>
		<script>window.ready = true</script>`)
	var cssContent strings.Builder
	var jsContent strings.Builder
	var inlineCSS []extractor.InlineResource
	var inlineJS []extractor.InlineResource
	cssIndex := 0
	jsIndex := 0

	extractInlineResources(
		doc,
		&cssContent,
		&jsContent,
		&inlineCSS,
		&inlineJS,
		&cssIndex,
		&jsIndex,
	)

	if len(inlineJS) != 1 {
		t.Fatalf("extractInlineResources() extracted %d scripts; want 1 executable script", len(inlineJS))
	}
	got := renderScraperDocument(t, doc)
	for _, want := range []string{
		`type="application/ld+json"`,
		`{"name":"Uncluster"}`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("extractInlineResources() removed data script content %q; got %s", want, got)
		}
	}
}

func TestExtractInlineResourcesUsesPortableRelativePaths(t *testing.T) {
	doc := parseScraperDocument(t, `
		<style>body { color: red; }</style>
		<script>window.ready = true</script>`)
	var cssContent strings.Builder
	var jsContent strings.Builder
	var inlineCSS []extractor.InlineResource
	var inlineJS []extractor.InlineResource
	cssIndex := 0
	jsIndex := 0

	extractInlineResources(
		doc,
		&cssContent,
		&jsContent,
		&inlineCSS,
		&inlineJS,
		&cssIndex,
		&jsIndex,
	)

	got := renderScraperDocument(t, doc)
	for _, want := range []string{
		`href="inline/style-1.css"`,
		`src="inline/script-1.js"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("extractInlineResources() output missing %q; got %s", want, got)
		}
	}
}

func TestRewriteCSSURLsUsesDownloadedAssets(t *testing.T) {
	css := `.hero { background: url("../images/hero.png"); }
@font-face { src: url('https://cdn.example.com/fonts/site.woff2') format('woff2'); }
.embedded { background: url(data:image/png;base64,AAAA); }`

	got := rewriteCSSURLs(
		css,
		"https://example.com/styles/site.css",
		"external/css/site.css",
		map[string]string{
			"https://example.com/images/hero.png":      "assets/hero.png",
			"https://cdn.example.com/fonts/site.woff2": "assets/site.woff2",
		},
	)

	for _, want := range []string{
		`url("../../assets/hero.png")`,
		`url('../../assets/site.woff2')`,
		`url(data:image/png;base64,AAAA)`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rewriteCSSURLs() output missing %q; got %s", want, got)
		}
	}
}
