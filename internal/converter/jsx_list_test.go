package converter

import (
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
	got := c.convertStyle(`font-family: 'Times New Roman'`, nil)
	want := `{{fontFamily: '\'Times New Roman\''}}`

	if got != want {
		t.Fatalf("convertStyle() = %s, want %s", got, want)
	}
}

func TestConvertStyleEmitsSubstitutionsAsBareExpressions(t *testing.T) {
	c := &JSXConverter{}
	subs := map[string]string{"12px": "item.fontSize"}

	got := c.convertStyle("font-size: 12px", subs)
	want := "{{fontSize: item.fontSize}}"

	if got != want {
		t.Fatalf("convertStyle() = %s, want %s", got, want)
	}
}

func TestConvertStyleSubstitutesBackgroundImageURL(t *testing.T) {
	c := &JSXConverter{}
	subs := map[string]string{"/img/hero.png": "item.backgroundImage"}

	got := c.convertStyle(`background-image: url('/img/hero.png')`, subs)
	want := "{{backgroundImage: `url(${item.backgroundImage})`}}"

	if got != want {
		t.Fatalf("convertStyle() = %s, want %s", got, want)
	}
}

func TestConvertAttributeDropsNamespacedAttributesWithSubstitutions(t *testing.T) {
	c := &JSXConverter{}
	attr := html.Attribute{Namespace: "xlink", Key: "title", Val: "x"}

	key, val := c.convertAttribute(attr, map[string]string{"x": "item.a"})
	if key != "" || val != "" {
		t.Fatalf("convertAttribute() namespaced = (%q, %q), want both empty", key, val)
	}
}

// assertJSXStructurallyValid catches the syntax errors the old duplicate
// converters produced: unbalanced braces, unescaped quotes inside attribute
// values, and single-braced style objects. It is a structural check, not a real
// parse; a genuine parse would need Node or esbuild.
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
