package scraper

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"

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

func TestFetchCSSResourcesFollowsImportsDepthFirst(t *testing.T) {
	fixtures := map[string]string{
		"https://example.com/css/root.css": `@import "nested/one.css";
.root { background: url("../images/root.png"); }`,
		"https://example.com/css/nested/one.css": `@import url('../two.css');
.one { background: url(../images/one.png); }`,
		"https://example.com/css/two.css": `@import "root.css";
.two { color: navy; }`,
		"https://example.com/css/after.css": `.after { color: green; }`,
	}
	var fetched []string
	fetchOne := func(rawURL string) fetcher.FetchedResource {
		fetched = append(fetched, rawURL)
		return fetcher.FetchedResource{
			URL:      rawURL,
			Filename: strings.TrimPrefix(rawURL, "https://example.com/css/"),
			Content:  fixtures[rawURL],
			Type:     "css",
		}
	}

	resources, binaryURLs := fetchCSSResources([]string{
		"https://example.com/css/root.css",
		"https://example.com/css/after.css",
	}, fetchOne)

	wantFetched := []string{
		"https://example.com/css/root.css",
		"https://example.com/css/nested/one.css",
		"https://example.com/css/two.css",
		"https://example.com/css/after.css",
	}
	if strings.Join(fetched, "\n") != strings.Join(wantFetched, "\n") {
		t.Fatalf("fetch order = %v, want %v", fetched, wantFetched)
	}
	if len(resources) != len(wantFetched) {
		t.Fatalf("resource count = %d, want %d", len(resources), len(wantFetched))
	}
	wantTypes := []string{"css", "css-import", "css-import", "css"}
	for i, wantType := range wantTypes {
		if resources[i].Type != wantType {
			t.Fatalf("resource %d type = %q, want %q", i, resources[i].Type, wantType)
		}
	}
	wantBinary := []string{
		"https://example.com/css/images/one.png",
		"https://example.com/images/root.png",
	}
	if strings.Join(binaryURLs, "\n") != strings.Join(wantBinary, "\n") {
		t.Fatalf("binary URLs = %v, want %v", binaryURLs, wantBinary)
	}
}

func TestCollectCSSResourcesReturnsFetchedStylesheetsAndRecordsPaths(t *testing.T) {
	capture := extractor.NewCaptureManifest("https://example.com/")
	paths := make(map[string]string)
	fetchOne := func(rawURL string) fetcher.FetchedResource {
		return fetcher.FetchedResource{
			URL: rawURL, Filename: "site.css", Content: ".page { color: navy; }", Type: "css",
		}
	}

	resources, binaryURLs := collectCSSResources(
		[]string{"https://example.com/site.css"}, fetchOne, capture, paths,
	)

	if len(resources) != 1 || resources[0].Content == "" {
		t.Fatalf("resources = %+v, want one non-empty stylesheet", resources)
	}
	if len(binaryURLs) != 0 {
		t.Fatalf("binary URLs = %v, want none", binaryURLs)
	}
	if got := paths["https://example.com/site.css"]; got != "external/css/site.css" {
		t.Fatalf("localized path = %q, want external/css/site.css", got)
	}
	if len(capture.Assets) != 1 || capture.Assets[0].Status != extractor.CaptureLocalized {
		t.Fatalf("capture assets = %+v, want one localized stylesheet", capture.Assets)
	}
}

func TestInstallWebflowLocalRuntimeCompatBracketsSourceScripts(t *testing.T) {
	doc := parseScraperDocument(t, `<html data-wf-domain="fixture.webflow.io"><head>
		<script>window.before = true</script></head><body>
		<script src="https://cdn.example.com/webflow.js"></script></body></html>`)
	base, err := url.Parse("https://fixture.webflow.io/")
	if err != nil {
		t.Fatal(err)
	}

	if !installWebflowLocalRuntimeCompat(doc, base) {
		t.Fatal("installWebflowLocalRuntimeCompat() = false, want true")
	}

	got := renderScraperDocument(t, doc)
	prepare := strings.Index(got, "__unclusterWebflowDomain")
	source := strings.Index(got, "window.before")
	runtime := strings.Index(got, "webflow.js")
	restore := strings.LastIndex(got, "__unclusterWebflowDomain")
	if prepare < 0 || source < 0 || runtime < 0 || restore < 0 || !(prepare < source && source < runtime && runtime < restore) {
		t.Fatalf("compatibility scripts do not bracket source scripts: %s", got)
	}
	if !strings.Contains(got, `data-wf-domain="fixture.webflow.io"`) {
		t.Fatalf("original Webflow domain attribute was not preserved: %s", got)
	}
}

func TestInstallWebflowLocalRuntimeCompatIgnoresNonWebflowDocuments(t *testing.T) {
	doc := parseScraperDocument(t, `<html data-wf-domain="example.com"><head></head><body></body></html>`)
	base, err := url.Parse("https://example.com/")
	if err != nil {
		t.Fatal(err)
	}

	if installWebflowLocalRuntimeCompat(doc, base) {
		t.Fatal("installWebflowLocalRuntimeCompat() = true for a non-Webflow domain")
	}
	if strings.Contains(renderScraperDocument(t, doc), "__unclusterWebflowDomain") {
		t.Fatal("non-Webflow document received compatibility scripts")
	}
}

func TestRewriteCSSURLsLocalizesQuotedAndURLImports(t *testing.T) {
	css := `@import "nested/one.css";
@import url('../shared/two.css');`

	got := rewriteCSSURLs(
		css,
		"https://example.com/css/root.css",
		"external/css/root.css",
		map[string]string{
			"https://example.com/css/nested/one.css": "external/css/one.css",
			"https://example.com/shared/two.css":     "external/css/two.css",
		},
	)

	for _, want := range []string{
		`@import "one.css"`,
		`@import url('two.css')`,
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

func TestFindAllAssetURLsPreservesCSSAndJSSourceOrder(t *testing.T) {
	doc := parseScraperDocument(t, `
		<link rel="stylesheet" href="/styles/reset.css">
		<script src="/scripts/runtime.js"></script>
		<link rel="stylesheet" href="/styles/theme.css">
		<script src="/scripts/app.js"></script>
		<link rel="stylesheet" href="/styles/reset.css">
		<script src="/scripts/runtime.js"></script>`)
	base, err := url.Parse("https://example.com/page")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	cssURLs, jsURLs, _ := findAllAssetURLs(doc, base)

	wantCSS := []string{
		"https://example.com/styles/reset.css",
		"https://example.com/styles/theme.css",
	}
	wantJS := []string{
		"https://example.com/scripts/runtime.js",
		"https://example.com/scripts/app.js",
	}
	if strings.Join(cssURLs, "\n") != strings.Join(wantCSS, "\n") {
		t.Fatalf("CSS order = %v, want %v", cssURLs, wantCSS)
	}
	if strings.Join(jsURLs, "\n") != strings.Join(wantJS, "\n") {
		t.Fatalf("JS order = %v, want %v", jsURLs, wantJS)
	}
}

func TestDataSrcLocalizationKeepsOrdinaryDataAndSrcset(t *testing.T) {
	doc := parseScraperDocument(t, `
		<div data-animation-type="lottie" data-src="https://cdn.example.com/hero.json"></div>
		<div data-src="carousel-panel"></div>
		<img data-src="/images/lazy.png" srcset="/images/small.png 1x, /images/large.png 2x">`)
	base, err := url.Parse("https://example.com/page")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	_, _, binaryURLs := findAllAssetURLs(doc, base)
	wantURLs := []string{
		"https://cdn.example.com/hero.json",
		"https://example.com/images/lazy.png",
		"https://example.com/images/small.png",
		"https://example.com/images/large.png",
	}
	if strings.Join(binaryURLs, "\n") != strings.Join(wantURLs, "\n") {
		t.Fatalf("binary URLs = %v, want %v", binaryURLs, wantURLs)
	}

	rewriteHTMLPaths(doc, map[string]string{
		"https://cdn.example.com/hero.json":    "assets/hero.json",
		"https://example.com/images/lazy.png":  "assets/lazy.png",
		"https://example.com/images/small.png": "assets/small.png",
		"https://example.com/images/large.png": "assets/large.png",
	}, base)

	got := renderScraperDocument(t, doc)
	for _, want := range []string{
		`data-src="assets/hero.json"`,
		`data-src="carousel-panel"`,
		`data-src="assets/lazy.png"`,
		`srcset="assets/small.png 1x, assets/large.png 2x"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rewriteHTMLPaths() output missing %q; got %s", want, got)
		}
	}
}

func TestFindRetainedExternalAssetsExcludesNavigationAndMetadata(t *testing.T) {
	doc := parseScraperDocument(t, `
		<link rel="preconnect" href="https://static.example.com">
		<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter">
		<a href="https://example.com/next">Next</a>
		<iframe src="https://player.example.com/embed/123"></iframe>`)
	base, err := url.Parse("https://example.com/page")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	got := findRetainedExternalAssets(doc, base)
	want := []captureReference{
		{URL: "https://fonts.googleapis.com/css2?family=Inter", Type: "css"},
		{URL: "https://player.example.com/embed/123", Type: "iframe"},
	}
	if len(got) != len(want) {
		t.Fatalf("retained assets = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("retained asset %d = %+v, want %+v", i, got[i], want[i])
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
