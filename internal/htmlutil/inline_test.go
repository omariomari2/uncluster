package htmlutil

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseDocument(t *testing.T, source string) *html.Node {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatalf("html.Parse() error = %v", err)
	}
	return doc
}

func renderDocument(t *testing.T, doc *html.Node) string {
	t.Helper()

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		t.Fatalf("html.Render() error = %v", err)
	}
	return buf.String()
}

func collect(t *testing.T, source string) (*InlineCollector, string) {
	t.Helper()

	doc := parseDocument(t, source)
	var c InlineCollector
	c.Walk(doc)
	return &c, renderDocument(t, doc)
}

func TestInlineCollectorPreservesDataScripts(t *testing.T) {
	c, got := collect(t, `
		<script type="application/ld+json">{"name":"Uncluster"}</script>
		<script>window.ready = true</script>`)

	if len(c.InlineJS) != 1 {
		t.Fatalf("extracted %d scripts; want 1 executable script", len(c.InlineJS))
	}
	for _, want := range []string{`type="application/ld+json"`, `{"name":"Uncluster"}`} {
		if !strings.Contains(got, want) {
			t.Errorf("data script content %q was removed; got %s", want, got)
		}
	}
}

func TestInlineCollectorUsesPortableRelativePaths(t *testing.T) {
	_, got := collect(t, `
		<style>body { color: red; }</style>
		<script>window.ready = true</script>`)

	for _, want := range []string{`href="inline/style-1.css"`, `src="inline/script-1.js"`} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q; got %s", want, got)
		}
	}
}

// The scraper's old private copy built its replacement nodes from scratch and
// silently dropped every other attribute.
func TestInlineCollectorKeepsOtherAttributes(t *testing.T) {
	_, got := collect(t, `
		<style media="print" data-id="a">body { color: red; }</style>
		<script nonce="abc123" data-role="boot">window.ready = true</script>`)

	for _, want := range []string{
		`media="print"`,
		`data-id="a"`,
		`nonce="abc123"`,
		`data-role="boot"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("attribute %q was dropped during extraction; got %s", want, got)
		}
	}
}

// The scraper's old private copy did not strip MIME parameters and so treated
// a perfectly ordinary script as data.
func TestInlineCollectorExtractsScriptsWithMIMEParameters(t *testing.T) {
	c, _ := collect(t, `<script type="text/javascript; charset=utf-8">window.ready = true</script>`)

	if len(c.InlineJS) != 1 {
		t.Fatalf("extracted %d scripts; want 1", len(c.InlineJS))
	}
}

func TestInlineCollectorSkipsEmptyAndExternal(t *testing.T) {
	c, got := collect(t, `
		<style>   </style>
		<script src="/vendor.js"></script>`)

	if len(c.InlineCSS) != 0 || len(c.InlineJS) != 0 {
		t.Fatalf("extracted %d styles and %d scripts; want none", len(c.InlineCSS), len(c.InlineJS))
	}
	if !strings.Contains(got, `src="/vendor.js"`) {
		t.Errorf("external script reference was altered; got %s", got)
	}
}

func TestInlineCollectorNumbersSequentially(t *testing.T) {
	c, _ := collect(t, `
		<style>a{}</style><style>b{}</style>
		<script>one()</script><script>two()</script>`)

	if len(c.InlineCSS) != 2 || len(c.InlineJS) != 2 {
		t.Fatalf("got %d styles and %d scripts; want 2 each", len(c.InlineCSS), len(c.InlineJS))
	}
	for i, want := range []string{"inline/style-1.css", "inline/style-2.css"} {
		if c.InlineCSS[i].Path != want {
			t.Errorf("InlineCSS[%d].Path = %q, want %q", i, c.InlineCSS[i].Path, want)
		}
	}
	if !strings.Contains(c.CSS.String(), "a{}") || !strings.Contains(c.CSS.String(), "b{}") {
		t.Errorf("aggregated CSS lost content: %q", c.CSS.String())
	}
}

func TestIsJavaScriptType(t *testing.T) {
	javascript := []string{"", "text/javascript", "  TEXT/JavaScript  ", "module", "text/javascript; charset=utf-8"}
	data := []string{"application/ld+json", "application/json", "text/template"}

	for _, v := range javascript {
		if !IsJavaScriptType(v) {
			t.Errorf("IsJavaScriptType(%q) = false, want true", v)
		}
	}
	for _, v := range data {
		if IsJavaScriptType(v) {
			t.Errorf("IsJavaScriptType(%q) = true, want false", v)
		}
	}
}

func TestAttrHelpersAreCaseInsensitive(t *testing.T) {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "img",
		Attr: []html.Attribute{{Key: "SrcSet", Val: "a.png"}},
	}

	if got := GetAttr(n, "srcset"); got != "a.png" {
		t.Errorf("GetAttr() = %q, want %q", got, "a.png")
	}
	if !HasAttr(n, "SRCSET") {
		t.Error("HasAttr() = false, want true")
	}

	SetAttr(n, "srcset", "b.png")
	if len(n.Attr) != 1 {
		t.Fatalf("SetAttr() appended a duplicate: %+v", n.Attr)
	}
	if got := GetAttr(n, "SrcSet"); got != "b.png" {
		t.Errorf("after SetAttr, GetAttr() = %q, want %q", got, "b.png")
	}
}
