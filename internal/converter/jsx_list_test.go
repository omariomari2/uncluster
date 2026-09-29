package converter

import (
	"regexp"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// The list-detection path (ConvertSectionToTSX -> detectListPattern) once had its
// own copy of the attribute and style converters, and every escaping fix landed
// only on the plain ConvertToJSX path. These tests mirror the plain-path
// assertions through the list path so the two cannot drift apart again.

// listFixture wraps markup in three identical siblings, which is the threshold
// that trips detectListPattern.
func listFixture(item string) string {
	return "<body><ul class=\"cards\">" +
		strings.Repeat("<li class=\"card\">"+item+"</li>", 3) +
		"</ul></body>"
}

func convertListForTest(t *testing.T, item string) string {
	t.Helper()

	got, err := ConvertSectionToTSX(listFixture(item), "Cards")
	if err != nil {
		t.Fatalf("ConvertSectionToTSX() error = %v", err)
	}
	assertJSXStructurallyValid(t, got)
	return got
}

func TestConvertSectionToTSXWrapsInlineStyleInAJSXObject(t *testing.T) {
	got := convertListForTest(t, `<span style="color: red; font-size: 12px">x</span>`)
	want := `style={{color: 'red', fontSize: '12px'}}`

	if !strings.Contains(got, want) {
		t.Fatalf("list-path style output missing %q; got:\n%s", want, got)
	}
}

func TestConvertSectionToTSXTreatsPresenceOnlyBooleanAttributesAsTrue(t *testing.T) {
	got := convertListForTest(t, `<input disabled><input required><input readonly><input multiple>`)

	for _, want := range []string{
		"disabled={true}",
		"required={true}",
		"readOnly={true}",
		"multiple={true}",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("list-path boolean output missing %q; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, `=""`) {
		t.Errorf("list-path emitted an empty-string boolean attribute; got:\n%s", got)
	}
}

func TestConvertSectionToTSXUsesReactSrcSetAttribute(t *testing.T) {
	got := convertListForTest(t, `<img srcset="small.png 1x, large.png 2x">`)

	if !strings.Contains(got, `srcSet="small.png 1x, large.png 2x"`) {
		t.Fatalf("srcset was not emitted as React's srcSet property; got:\n%s", got)
	}
}

func TestConvertSectionToTSXDoesNotImportUnusedReactBinding(t *testing.T) {
	got := convertListForTest(t, `<span>content</span>`)

	if strings.Contains(got, "import React") {
		t.Fatalf("generated component contains an unused React import; got:\n%s", got)
	}
}

func TestConvertSectionToTSXEscapesAttributeAndTextSyntax(t *testing.T) {
	got := convertListForTest(t, `<span title="say &quot;hi&quot; &amp; {x}">a &amp; b {y}</span>`)

	for _, want := range []string{
		`title="say &quot;hi&quot; &amp; &#123;x&#125;"`,
		`a &amp; b &#123;y&#125;`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("list-path escaped output missing %q; got:\n%s", want, got)
		}
	}
}

func TestConvertStyleEscapesQuotesInValues(t *testing.T) {
	c := &JSXConverter{}
	got := c.convertStyle(`font-family: 'Times New Roman'`)
	want := `{{fontFamily: '\'Times New Roman\''}}`

	if got != want {
		t.Fatalf("convertStyle() = %s, want %s", got, want)
	}
}

func TestConvertStyleUsesReactCasingForMSVendorPrefix(t *testing.T) {
	c := &JSXConverter{}
	got := c.convertStyle(`-ms-grid-row: 2; -webkit-line-clamp: 3`)
	want := `{{msGridRow: '2', WebkitLineClamp: '3'}}`

	if got != want {
		t.Fatalf("convertStyle() = %s, want %s", got, want)
	}
}

func TestConvertAttributeDropsNamespacedAttributes(t *testing.T) {
	c := &JSXConverter{}
	attr := html.Attribute{Namespace: "xlink", Key: "title", Val: "x"}

	key, val := c.convertAttribute(attr)
	if key != "" || val != "" {
		t.Fatalf("convertAttribute() namespaced = (%q, %q), want both empty", key, val)
	}
}

// assertJSXStructurallyValid catches the syntax errors the old duplicate
// converters produced: unbalanced braces, unescaped quotes inside attribute
// values, single-braced style objects, and raw braces left in text position.
// It is a structural check, not a real parse; a genuine parse would need Node
// or esbuild.
func assertJSXStructurallyValid(t *testing.T, src string) {
	t.Helper()

	for _, line := range strings.Split(src, "\n") {
		if idx := strings.Index(line, "style="); idx != -1 {
			if !strings.HasPrefix(line[idx+len("style="):], "{{") {
				t.Errorf("style attribute is not a JSX object literal: %s", strings.TrimSpace(line))
			}
		}
		if bad := findUnescapedAttributeQuote(line); bad != "" {
			t.Errorf("attribute value contains an unescaped quote: %s", bad)
		}
	}

	if depth := braceDepth(src); depth != 0 {
		t.Errorf("unbalanced braces in generated JSX (net depth %d):\n%s", depth, src)
	}

	// Brace-in-text only makes sense inside the returned markup; the surrounding
	// TypeScript legitimately uses braces.
	jsx := src
	if body, err := extractJSX(src); err == nil {
		jsx = body
	}
	if stray := strayTextBraces(jsx); stray != "" {
		t.Errorf("raw brace left in JSX text position (should be escaped to &#123;/&#125;): %s", stray)
	}
}

var (
	jsxExprAttrPattern = regexp.MustCompile(`[A-Za-z-]+=\{(?:[^{}]|\{[^{}]*\})*\}`)
	jsxQuotedPattern   = regexp.MustCompile(`[A-Za-z-]+="[^"]*"`)
	jsxCommentPattern  = regexp.MustCompile(`\{/\*.*?\*/\}`)
)

// strayTextBraces reports a brace sitting in text position. Every legitimate
// brace this converter emits belongs to an expression attribute, a quoted
// value, or a JSX comment; anything left after removing those is unescaped text
// that would be parsed as an expression.
func strayTextBraces(src string) string {
	for _, line := range strings.Split(src, "\n") {
		stripped := jsxCommentPattern.ReplaceAllString(line, "")
		stripped = jsxExprAttrPattern.ReplaceAllString(stripped, "")
		stripped = jsxQuotedPattern.ReplaceAllString(stripped, "")
		if strings.ContainsAny(stripped, "{}") {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// findUnescapedAttributeQuote reports an attr="..." whose value contains a bare
// double quote, which terminates the attribute early and breaks the parse.
// The converter escapes " to &quot;, so a value's closing quote must always be
// followed by a separator rather than more value text.
func findUnescapedAttributeQuote(line string) string {
	rest := line
	for {
		eq := strings.Index(rest, `="`)
		if eq == -1 {
			return ""
		}
		value := rest[eq+2:]
		end := strings.Index(value, `"`)
		if end == -1 {
			return ""
		}
		switch after := value[end+1:]; {
		case after == "",
			strings.HasPrefix(after, " "),
			strings.HasPrefix(after, ">"),
			strings.HasPrefix(after, "/"):
		default:
			return strings.TrimSpace(line)
		}
		rest = value[end+1:]
	}
}

// braceDepth counts net { } nesting outside of quoted strings.
func braceDepth(src string) int {
	depth := 0
	var quote rune
	var prev rune

	for _, r := range src {
		switch {
		case quote != 0:
			if r == quote && prev != '\\' {
				quote = 0
			}
		case r == '\'' || r == '"' || r == '`':
			quote = r
		case r == '{':
			depth++
		case r == '}':
			depth--
		}
		prev = r
	}
	return depth
}

// Ported from the deleted ConvertToJSX suite: whitespace between inline
// elements is meaningful and must survive conversion (DEF-009).
func TestConvertSectionToTSXPreservesSpacesAroundInlineElements(t *testing.T) {
	got, err := ConvertSectionToTSX(`<body><p>Hello <strong>world</strong> again</p></body>`, "Copy")
	if err != nil {
		t.Fatalf("ConvertSectionToTSX() error = %v", err)
	}
	assertJSXStructurallyValid(t, got)

	want := `<p>Hello <strong>world</strong> again</p>`
	if !strings.Contains(got, want) {
		t.Fatalf("inline text missing %q; got:\n%s", want, got)
	}
}
