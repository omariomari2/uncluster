package scraper

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

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

// <picture><source srcset> is the standard responsive-image pattern. Discovery
// previously handled srcset only on <img>, so these assets were never
// downloaded even though the rewrite walk was ready to localise them.
func TestFindAllAssetURLsDiscoversPictureSourceSrcset(t *testing.T) {
	doc := parseScraperDocument(t, `
		<picture>
		  <source srcset="/img/wide.png 1x, /img/wide@2x.png 2x" media="(min-width: 800px)">
		  <img src="/img/fallback.png" srcset="/img/small.png 1x">
		</picture>`)
	base, err := url.Parse("https://example.com/page")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	_, _, binaryURLs := findAllAssetURLs(doc, base)

	got := make(map[string]bool, len(binaryURLs))
	for _, u := range binaryURLs {
		got[u] = true
	}

	for _, want := range []string{
		"https://example.com/img/wide.png",
		"https://example.com/img/wide@2x.png",
		"https://example.com/img/fallback.png",
		"https://example.com/img/small.png",
	} {
		if !got[want] {
			t.Errorf("findAllAssetURLs() did not discover %s; got %v", want, binaryURLs)
		}
	}
}

func TestRewriteHTMLPathsLocalizesPictureSourceSrcset(t *testing.T) {
	doc := parseScraperDocument(t, `
		<picture>
		  <source srcset="/img/wide.png 1x, /img/wide@2x.png 2x">
		  <img src="/img/fallback.png">
		</picture>`)
	base, err := url.Parse("https://example.com/page")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	rewriteHTMLPaths(doc, map[string]string{
		"https://example.com/img/wide.png":     "assets/wide.png",
		"https://example.com/img/wide@2x.png":  "assets/wide-2x.png",
		"https://example.com/img/fallback.png": "assets/fallback.png",
	}, base)

	got := renderScraperDocument(t, doc)
	for _, want := range []string{
		`srcset="assets/wide.png 1x, assets/wide-2x.png 2x"`,
		`src="assets/fallback.png"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rewriteHTMLPaths() output missing %q; got %s", want, got)
		}
	}
}
