package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// Golden-DOM fidelity harness.
//
// The promise of the TSX conversion is that the generated components render the
// scraped page indistinguishably from the original. Because components preserve
// the DOM verbatim, that promise is checkable without executing React: invert
// the converter's transforms on the emitted JSX, parse the result as HTML, and
// compare it with the source tree.
//
// What this proves: no node, attribute, or text is dropped, duplicated,
// reordered or corrupted, and escaping round-trips.
//
// What it does NOT prove: that the mapping tables are themselves correct. The
// inverse is derived from the same tables, so a wrong entry cancels out. Table
// correctness is a separate concern from structural fidelity — see
// TestAttributeMapIsInvertible for the part of that which is checkable here.
// CSS cascade, runtime behaviour and hydration need a browser and remain out of
// scope.

func TestFidelityAgainstFixtures(t *testing.T) {
	fixtures, err := filepath.Glob(filepath.Join("testdata", "*.html"))
	if err != nil {
		t.Fatalf("glob testdata: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixtures found in testdata/")
	}

	for _, path := range fixtures {
		t.Run(strings.TrimSuffix(filepath.Base(path), ".html"), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}

			tsx, err := ConvertSectionToTSX(string(source), "Fixture")
			if err != nil {
				t.Fatalf("ConvertSectionToTSX() error = %v", err)
			}

			assertJSXStructurallyValid(t, tsx)

			jsx, err := extractJSX(tsx)
			if err != nil {
				t.Fatalf("%v\n--- generated ---\n%s", err, tsx)
			}

			want := normalizeTree(t, parseFragment(t, string(source)))
			generated := parseFragment(t, jsxToHTML(jsx))
			applyReactInverse(generated)
			got := normalizeTree(t, generated)

			if want != got {
				t.Errorf("DOM fidelity lost\n\nwant: %s\n\ngot:  %s\n\n--- generated TSX ---\n%s",
					want, got, tsx)
			}
		})
	}
}

// TestAttributeMapIsInvertible guards the one table property the fidelity test
// cannot: that no two HTML attributes map onto the same JSX name, which would
// make the transform lossy.
func TestAttributeMapIsInvertible(t *testing.T) {
	seen := map[string]string{}
	for htmlName, jsxName := range jsxAttributeMap {
		key := strings.ToLower(jsxName)
		if prev, clash := seen[key]; clash {
			t.Errorf("%q and %q both map to %q", prev, htmlName, jsxName)
		}
		seen[key] = htmlName
	}
}

// ---------- extracting the JSX ----------

// extractJSX returns the markup inside the component's return statement.
func extractJSX(tsx string) (string, error) {
	const open = "  return (\n"
	start := strings.Index(tsx, open)
	if start == -1 {
		return "", fmt.Errorf("no return statement in generated component")
	}
	body := tsx[start+len(open):]

	end := strings.Index(body, "\n  )\n")
	if end == -1 {
		return "", fmt.Errorf("unterminated return statement in generated component")
	}
	return body[:end], nil
}

// ---------- inverting the transforms ----------

var (
	// jsxStylePattern matches style={{ ... }} including the outer braces.
	jsxStylePattern = regexp.MustCompile(`style=\{\{([^}]*)\}\}`)
	// jsxHandlerPattern matches onClick={() => { body }}.
	jsxHandlerPattern = regexp.MustCompile(`(on[A-Z][A-Za-z]*)=\{\(\) => \{ (.*?) \}\}`)
	// jsxBoolPattern matches disabled={true}.
	jsxBoolPattern = regexp.MustCompile(`([A-Za-z-]+)=\{true\}`)
	// jsxDefaultValuePattern matches defaultValue={"x"} and defaultValue={["a","b"]}.
	jsxDefaultValuePattern = regexp.MustCompile(`defaultValue=\{(\[[^\]]*\]|"(?:[^"\\]|\\.)*")\}`)
	// jsStringLiteralPattern matches one double-quoted JS string literal.
	jsStringLiteralPattern = regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
)

// inverseAttributeMap maps the lowercased JSX attribute name back to its HTML
// name. It is derived from jsxAttributeMap so the two cannot drift apart. Most
// entries are pure camelCasing, which html.Parse already undoes by lowercasing;
// only genuinely renamed attributes (class/className, fill-rule/fillRule) end
// up here.
var inverseAttributeMap = func() map[string]string {
	m := map[string]string{}
	for htmlName, jsxName := range jsxAttributeMap {
		if lower := strings.ToLower(jsxName); lower != htmlName {
			m[lower] = htmlName
		}
	}
	return m
}()

// jsxToHTML rewrites emitted JSX into markup that html.Parse can read, undoing
// exactly the transforms the converter applied.
func jsxToHTML(jsx string) string {
	out := strings.ReplaceAll(jsx, "<>", "")
	out = strings.ReplaceAll(out, "</>", "")

	out = jsxStylePattern.ReplaceAllStringFunc(out, func(match string) string {
		inner := jsxStylePattern.FindStringSubmatch(match)[1]
		return fmt.Sprintf("style=%q", styleObjectToCSS(inner))
	})

	out = jsxHandlerPattern.ReplaceAllStringFunc(out, func(match string) string {
		parts := jsxHandlerPattern.FindStringSubmatch(match)
		return fmt.Sprintf("%s=%q", strings.ToLower(parts[1]), parts[2])
	})

	// A boolean attribute round-trips as bare presence.
	out = jsxBoolPattern.ReplaceAllString(out, `$1=""`)

	// defaultValue holds a JS string literal, which html.Parse cannot read as an
	// attribute value. Flatten it to a plain attribute for applyReactInverse to
	// turn back into DOM.
	out = jsxDefaultValuePattern.ReplaceAllStringFunc(out, func(match string) string {
		inner := jsxDefaultValuePattern.FindStringSubmatch(match)[1]
		var values []string
		for _, lit := range jsStringLiteralPattern.FindAllString(inner, -1) {
			if v, err := strconv.Unquote(lit); err == nil {
				values = append(values, v)
			}
		}
		// Written as an HTML attribute, not re-quoted: HTML does not decode
		// backslash escapes, so the value must carry real newlines and tabs.
		return `defaultvalue="` + htmlAttrEscape(strings.Join(values, defaultValueSep)) + `"`
	})

	// Self-closing syntax is only valid for void elements in HTML. <textarea />
	// is RCDATA, so an unclosed tag swallows the rest of the document.
	out = selfClosedTextareaPattern.ReplaceAllString(out, `$1></textarea>`)

	return out
}

var selfClosedTextareaPattern = regexp.MustCompile(`(<textarea\b[^>]*?)\s*/>`)

var htmlAttrEscaper = strings.NewReplacer(
	"&", "&amp;",
	`"`, "&quot;",
	"<", "&lt;",
	">", "&gt;",
)

func htmlAttrEscape(s string) string { return htmlAttrEscaper.Replace(s) }

// defaultValueSep joins a multiple-select's values inside a single attribute.
const defaultValueSep = "\x1f"

// applyReactInverse undoes the divergences React requires, so the generated
// tree can be compared against the source DOM. React reconstructs exactly this
// at runtime: textarea content from defaultValue, and option selection from the
// enclosing select's defaultValue.
func applyReactInverse(n *html.Node) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "textarea":
			if value, ok := takeAttr(n, "defaultvalue"); ok {
				for n.FirstChild != nil {
					n.RemoveChild(n.FirstChild)
				}
				n.AppendChild(&html.Node{Type: html.TextNode, Data: value})
			}
		case "select":
			if value, ok := takeAttr(n, "defaultvalue"); ok {
				chosen := map[string]bool{}
				for _, v := range strings.Split(value, defaultValueSep) {
					chosen[v] = true
				}
				markSelectedOptions(n, chosen)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		applyReactInverse(c)
	}
}

func markSelectedOptions(sel *html.Node, chosen map[string]bool) {
	for c := sel.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if c.Data == "option" && chosen[optionValue(c)] {
			c.Attr = append(c.Attr, html.Attribute{Key: "selected", Val: ""})
		}
		markSelectedOptions(c, chosen)
	}
}

// takeAttr removes an attribute and returns its value.
func takeAttr(n *html.Node, key string) (string, bool) {
	for i, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
			return a.Val, true
		}
	}
	return "", false
}

// styleObjectToCSS turns {fontSize: '12px', color: 'red'} back into
// "font-size: 12px; color: red".
func styleObjectToCSS(inner string) string {
	var decls []string
	for _, part := range splitTopLevel(inner) {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := camelToKebab(strings.TrimSpace(kv[0]))
		val := strings.TrimSpace(kv[1])
		val = strings.TrimPrefix(val, "'")
		val = strings.TrimSuffix(val, "'")
		val = strings.ReplaceAll(val, `\'`, `'`)
		val = strings.ReplaceAll(val, `\\`, `\`)
		decls = append(decls, key+": "+val)
	}
	return strings.Join(decls, "; ")
}

// splitTopLevel splits on commas that are not inside quotes.
func splitTopLevel(s string) []string {
	var parts []string
	var buf strings.Builder
	var quote rune
	var prev rune

	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote && prev != '\\' {
				quote = 0
			}
			buf.WriteRune(r)
		case r == '\'' || r == '"':
			quote = r
			buf.WriteRune(r)
		case r == ',':
			parts = append(parts, buf.String())
			buf.Reset()
		default:
			buf.WriteRune(r)
		}
		prev = r
	}
	parts = append(parts, buf.String())
	return parts
}

func camelToKebab(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// ---------- normalising ----------

func parseFragment(t *testing.T, markup string) *html.Node {
	t.Helper()

	doc, err := html.Parse(strings.NewReader("<body>" + markup + "</body>"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	body := findBodyNode(doc)
	if body == nil {
		t.Fatalf("no body in parsed fragment")
	}
	return body
}

// normalizeTree renders a canonical form: attributes sorted and lowercased,
// boolean attributes reduced to presence, whitespace runs collapsed, comments
// and skipped elements dropped.
func normalizeTree(t *testing.T, n *html.Node) string {
	t.Helper()

	var b strings.Builder
	writeNormalized(&b, n)
	return b.String()
}

func writeNormalized(b *strings.Builder, n *html.Node) {
	writeNormalizedIn(b, n, false)
}

// writeNormalizedIn renders n. Inside a whitespace-preserving element every
// space, tab and newline is significant, so text is compared byte for byte
// rather than collapsed.
func writeNormalizedIn(b *strings.Builder, n *html.Node, preserve bool) {
	switch n.Type {
	case html.TextNode:
		text := n.Data
		if !preserve {
			text = collapseSpace(text)
		}
		if text != "" {
			b.WriteString("#text(" + text + ")")
		}
		return
	case html.CommentNode, html.DoctypeNode:
		return
	}

	if n.Type == html.ElementNode {
		if skipElements[n.Data] && n.Data != "body" {
			return
		}
		if n.Data != "body" {
			b.WriteString("<" + n.Data)
			for _, a := range normalizedAttrs(n) {
				b.WriteString(" " + a)
			}
			b.WriteString(">")
		}
		if preserveWhitespaceElements[n.Data] {
			preserve = true
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		writeNormalizedIn(b, c, preserve)
	}

	if n.Type == html.ElementNode && n.Data != "body" {
		b.WriteString("</" + n.Data + ">")
	}
}

func normalizedAttrs(n *html.Node) []string {
	var attrs []string
	for _, a := range n.Attr {
		name := strings.ToLower(a.Key)
		if orig, ok := inverseAttributeMap[name]; ok {
			name = orig
		}

		val := a.Val
		if jsxBooleanAttributes[name] {
			// disabled, disabled="disabled" and disabled="" are the same thing.
			val = ""
		}
		if name == "style" {
			val = canonicalCSS(val)
		}
		attrs = append(attrs, fmt.Sprintf("%s=%q", name, val))
	}
	sort.Strings(attrs)
	return attrs
}

// canonicalCSS normalises spacing and drops a trailing semicolon so that
// "color:red;" and "color: red" compare equal.
func canonicalCSS(style string) string {
	var decls []string
	for _, part := range strings.Split(style, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		decls = append(decls, strings.TrimSpace(kv[0])+": "+collapseSpace(kv[1]))
	}
	return strings.Join(decls, "; ")
}

func collapseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// React imposes rules that plain HTML does not. These are the cases where the
// generated JSX deliberately differs from the source DOM; applyReactInverse
// encodes what React reconstructs at runtime, so TestFidelityAgainstFixtures
// still holds the output to the source.
func TestGeneratedJSXIsReactValid(t *testing.T) {
	cases := []struct {
		name     string
		markup   string
		wantAny  []string
		wantNone []string
	}{
		{
			name:     "textarea content becomes defaultValue",
			markup:   `<form><textarea name="bio">hello</textarea><p>x</p></form>`,
			wantAny:  []string{`defaultValue={"hello"}`},
			wantNone: []string{`<textarea name="bio">hello</textarea>`},
		},
		{
			name:     "textarea newlines survive as escapes",
			markup:   "<form><textarea>a\nb</textarea><p>x</p></form>",
			wantAny:  []string{`defaultValue={"a\nb"}`},
			wantNone: []string{},
		},
		{
			name:     "select selection moves off the option",
			markup:   `<form><select name="t"><option value="a">A</option><option value="b" selected>B</option></select><p>x</p></form>`,
			wantAny:  []string{`defaultValue={"b"}`},
			wantNone: []string{"selected={true}", "selected="},
		},
		{
			name:     "option without value falls back to its text",
			markup:   `<form><select><option selected>Alpha</option><option>Beta</option></select><p>x</p></form>`,
			wantAny:  []string{`defaultValue={"Alpha"}`},
			wantNone: []string{"selected="},
		},
		{
			name:     "multiple select yields an array",
			markup:   `<form><select multiple><option selected>A</option><option selected>B</option></select><p>x</p></form>`,
			wantAny:  []string{`defaultValue={["A", "B"]}`},
			wantNone: []string{"selected="},
		},
		{
			name:     "select with no selection gets no defaultValue",
			markup:   `<form><select><option>A</option></select><p>x</p></form>`,
			wantAny:  []string{},
			wantNone: []string{"defaultValue"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConvertSectionToTSX("<body>"+tc.markup+"</body>", "Fixture")
			if err != nil {
				t.Fatalf("ConvertSectionToTSX() error = %v", err)
			}
			for _, want := range tc.wantAny {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q; got:\n%s", want, got)
				}
			}
			for _, unwanted := range tc.wantNone {
				if strings.Contains(got, unwanted) {
					t.Errorf("output should not contain %q; got:\n%s", unwanted, got)
				}
			}
		})
	}
}
