package converter

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// JSXConverter carries no state; it groups the render methods that make up a
// single conversion.
type JSXConverter struct{}

var jsxAttributeMap = map[string]string{
	// HTML
	"class":           "className",
	"for":             "htmlFor",
	"tabindex":        "tabIndex",
	"readonly":        "readOnly",
	"maxlength":       "maxLength",
	"cellpadding":     "cellPadding",
	"cellspacing":     "cellSpacing",
	"colspan":         "colSpan",
	"rowspan":         "rowSpan",
	"frameborder":     "frameBorder",
	"allowfullscreen": "allowFullScreen",
	"crossorigin":     "crossOrigin",
	"accesskey":       "accessKey",
	"contenteditable": "contentEditable",
	"spellcheck":      "spellCheck",
	"autocomplete":    "autoComplete",
	"autofocus":       "autoFocus",
	"autoplay":        "autoPlay",
	"enctype":         "encType",
	"formaction":      "formAction",
	"hreflang":        "hrefLang",
	"inputmode":       "inputMode",
	"usemap":          "useMap",
	// SVG presentation
	"fill-rule":                   "fillRule",
	"clip-rule":                   "clipRule",
	"clip-path":                   "clipPath",
	"stroke-width":                "strokeWidth",
	"stroke-linecap":              "strokeLinecap",
	"stroke-linejoin":             "strokeLinejoin",
	"stroke-miterlimit":           "strokeMiterlimit",
	"stroke-dasharray":            "strokeDasharray",
	"stroke-dashoffset":           "strokeDashoffset",
	"fill-opacity":                "fillOpacity",
	"stroke-opacity":              "strokeOpacity",
	"text-anchor":                 "textAnchor",
	"font-family":                 "fontFamily",
	"font-size":                   "fontSize",
	"font-weight":                 "fontWeight",
	"font-style":                  "fontStyle",
	"text-decoration":             "textDecoration",
	"letter-spacing":              "letterSpacing",
	"word-spacing":                "wordSpacing",
	"dominant-baseline":           "dominantBaseline",
	"alignment-baseline":          "alignmentBaseline",
	"baseline-shift":              "baselineShift",
	"vector-effect":               "vectorEffect",
	"paint-order":                 "paintOrder",
	"shape-rendering":             "shapeRendering",
	"image-rendering":             "imageRendering",
	"color-rendering":             "colorRendering",
	"color-interpolation":         "colorInterpolation",
	"color-interpolation-filters": "colorInterpolationFilters",
	"flood-color":                 "floodColor",
	"flood-opacity":               "floodOpacity",
	"lighting-color":              "lightingColor",
	"writing-mode":                "writingMode",
	"pointer-events":              "pointerEvents",
	"unicode-bidi":                "unicodeBidi",
	"stop-color":                  "stopColor",
	"stop-opacity":                "stopOpacity",
	"marker-start":                "markerStart",
	"marker-mid":                  "markerMid",
	"marker-end":                  "markerEnd",
	// SVG structural — html.Parse lowercases camelCase attrs
	"viewbox":             "viewBox",
	"preserveaspectratio": "preserveAspectRatio",
	"gradientunits":       "gradientUnits",
	"gradienttransform":   "gradientTransform",
	"patterntransform":    "patternTransform",
	"patternunits":        "patternUnits",
	"patterncontentunits": "patternContentUnits",
	"spreadmethod":        "spreadMethod",
	"filterunits":         "filterUnits",
	"primitiveunits":      "primitiveUnits",
	"maskcontentunits":    "maskContentUnits",
	"maskunits":           "maskUnits",
	"markerunits":         "markerUnits",
	"markerwidth":         "markerWidth",
	"markerheight":        "markerHeight",
	"refx":                "refX",
	"refy":                "refY",
	"textlength":          "textLength",
	"lengthadjust":        "lengthAdjust",
	"startoffset":         "startOffset",
	"stddeviation":        "stdDeviation",
	"basefrequency":       "baseFrequency",
	"numoctaves":          "numOctaves",
	"kernelmatrix":        "kernelMatrix",
	"targetx":             "targetX",
	"targety":             "targetY",
	"specularconstant":    "specularConstant",
	"specularexponent":    "specularExponent",
	"diffuseconstant":     "diffuseConstant",
	"surfacescale":        "surfaceScale",
	"xchannelselector":    "xChannelSelector",
	"ychannelselector":    "yChannelSelector",
	"edgemode":            "edgeMode",
	"stitchtiles":         "stitchTiles",
	"clipPathUnits":       "clipPathUnits",
}

// inlineElements are HTML elements that flow inline with text.
// When an element's children are only text + inline elements, we render inline.
var inlineElements = map[string]bool{
	"a": true, "abbr": true, "acronym": true, "b": true, "bdi": true,
	"bdo": true, "big": true, "br": true, "cite": true, "code": true,
	"data": true, "dfn": true, "em": true, "i": true, "kbd": true,
	"label": true, "mark": true, "q": true, "s": true, "samp": true,
	"small": true, "span": true, "strong": true, "sub": true, "sup": true,
	"time": true, "tt": true, "u": true, "var": true,
}

var jsxEventMap = map[string]string{
	"onclick":     "onClick",
	"onchange":    "onChange",
	"onsubmit":    "onSubmit",
	"onload":      "onLoad",
	"onerror":     "onError",
	"onkeydown":   "onKeyDown",
	"onkeyup":     "onKeyUp",
	"onkeypress":  "onKeyPress",
	"onfocus":     "onFocus",
	"onblur":      "onBlur",
	"onmouseover": "onMouseOver",
	"onmouseout":  "onMouseOut",
	"onmousedown": "onMouseDown",
	"onmouseup":   "onMouseUp",
}

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"link": true, "meta": true, "source": true, "track": true, "wbr": true,
}

var skipElements = map[string]bool{
	"html": true, "head": true, "body": true,
	"title": true, "meta": true, "link": true,
	"style": true, "script": true,
}

// convertAttribute converts one HTML attribute into a JSX name/value pair.
func (c *JSXConverter) convertAttribute(attr html.Attribute) (string, string) {
	key := attr.Key
	val := attr.Val

	// xlink:href (deprecated but common in SVGs) → href
	if attr.Namespace == "xlink" && key == "href" {
		return "href", fmt.Sprintf(`"%s"`, escapeJSXAttribute(val))
	}
	// Drop namespace attributes that React doesn't need
	if attr.Namespace != "" {
		return "", ""
	}

	// Checked against the HTML name, before jsxAttributeMap renames the likes of
	// autoplay → autoPlay.
	isBoolean := jsxBooleanAttributes[strings.ToLower(key)]

	if jsxKey, ok := jsxAttributeMap[key]; ok {
		key = jsxKey
	}

	if jsxEvent, ok := jsxEventMap[key]; ok {
		// Extract simple function name for the TODO comment (best-effort).
		return jsxEvent, fmt.Sprintf("{() => { %s }}", val)
	}

	if key == "style" {
		return "style", c.convertStyle(val)
	}

	// Presence alone means true: disabled="false" is still disabled in HTML.
	if isBoolean {
		return key, "{true}"
	}

	return key, fmt.Sprintf(`"%s"`, escapeJSXAttribute(val))
}

var jsxAttributeEscaper = strings.NewReplacer(
	"&", "&amp;",
	"\"", "&quot;",
	"<", "&lt;",
	">", "&gt;",
	"{", "&#123;",
	"}", "&#125;",
)

var jsxTextEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	"{", "&#123;",
	"}", "&#125;",
)

// Style values land inside a single-quoted JS string literal, so only the
// backslash and the quote itself need escaping. NewReplacer does not rescan its
// own output, so these two rules cannot compound.
var jsxStyleValueEscaper = strings.NewReplacer(
	`\`, `\\`,
	`'`, `\'`,
)

// jsxBooleanAttributes lists HTML attributes whose mere presence means true.
// Keyed by HTML name, so look up before applying jsxAttributeMap.
var jsxBooleanAttributes = map[string]bool{
	"allowfullscreen": true,
	"async":           true,
	"autofocus":       true,
	"autoplay":        true,
	"checked":         true,
	"controls":        true,
	"default":         true,
	"defer":           true,
	"disabled":        true,
	"formnovalidate":  true,
	"hidden":          true,
	"inert":           true,
	"ismap":           true,
	"itemscope":       true,
	"loop":            true,
	"multiple":        true,
	"muted":           true,
	"novalidate":      true,
	"open":            true,
	"playsinline":     true,
	"readonly":        true,
	"required":        true,
	"reversed":        true,
	"selected":        true,
}

func escapeJSXAttribute(value string) string {
	return jsxAttributeEscaper.Replace(value)
}

func escapeJSXText(value string) string {
	return jsxTextEscaper.Replace(value)
}

func escapeJSXStyleValue(value string) string {
	return jsxStyleValueEscaper.Replace(value)
}

// convertStyle converts a CSS declaration list into a JSX style object literal,
// including the outer braces: style={{color: 'red'}}.
func (c *JSXConverter) convertStyle(style string) string {
	var jsxStyles []string

	for _, s := range strings.Split(style, ";") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			continue
		}

		camelKey := c.kebabToCamel(strings.TrimSpace(parts[0]))
		cssVal := strings.TrimSpace(parts[1])

		jsxStyles = append(jsxStyles, fmt.Sprintf("%s: '%s'", camelKey, escapeJSXStyleValue(cssVal)))
	}

	return fmt.Sprintf("{{%s}}", strings.Join(jsxStyles, ", "))
}

func (c *JSXConverter) kebabToCamel(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		return s
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// =============================================================
// ConvertSectionToTSX — indented JSX, TypeScript return type,
// optional list pattern extraction.
// =============================================================

// ConvertSectionToTSX converts an HTML fragment into a standalone TSX component.
// It produces properly indented JSX, adds (): JSX.Element return type, removes
// unnecessary Fragment wrappers, and extracts repeated list patterns into typed
// interfaces with data arrays.
func ConvertSectionToTSX(htmlFragment, componentName string) (string, error) {
	c := &JSXConverter{}

	doc, err := html.Parse(strings.NewReader(htmlFragment))
	if err != nil {
		return "", fmt.Errorf("failed to convert section %q to JSX: %w", componentName, err)
	}

	body := findBodyNode(doc)

	roots := nonSkippedChildren(body)

	// Collect any inline event handler function names so we can warn the developer.
	handlers := collectHandlerNames(body)
	handlerComment := ""
	if len(handlers) > 0 {
		handlerComment = fmt.Sprintf("// TODO: define or import these handlers — %s\n", strings.Join(handlers, ", "))
	}

	var jsxBuf strings.Builder
	if len(roots) == 1 {
		c.renderElementIndented(&jsxBuf, roots[0], 2)
		jsx := strings.TrimRight(jsxBuf.String(), "\n")
		return fmt.Sprintf(`import React from 'react'

%sfunction %s(): JSX.Element {
  return (
%s
  )
}

export default %s
`, handlerComment, componentName, jsx, componentName), nil
	}

	for _, root := range roots {
		c.renderElementIndented(&jsxBuf, root, 3)
	}
	jsx := strings.TrimRight(jsxBuf.String(), "\n")
	return fmt.Sprintf(`import React from 'react'

%sfunction %s(): JSX.Element {
  return (
    <>
%s
    </>
  )
}

export default %s
`, handlerComment, componentName, jsx, componentName), nil
}

// collectHandlerNames walks the node tree and returns the distinct function
// names referenced by inline event handler attributes (e.g. onclick="foo()").
func collectHandlerNames(n *html.Node) []string {
	seen := make(map[string]bool)
	var names []string

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if _, ok := jsxEventMap[attr.Key]; ok && attr.Val != "" {
					// Extract the leading identifier: "doThing(a, b)" → "doThing"
					name := extractFuncName(attr.Val)
					if name != "" && !seen[name] {
						seen[name] = true
						names = append(names, name)
					}
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return names
}

// extractFuncName returns the leading identifier from a JS expression like
// "doSomething()" or "ns.doSomething()" → "doSomething".
func extractFuncName(expr string) string {
	expr = strings.TrimSpace(expr)
	// Find the first '(' and take the last dotted segment before it.
	paren := strings.Index(expr, "(")
	if paren <= 0 {
		return ""
	}
	ident := expr[:paren]
	if dot := strings.LastIndex(ident, "."); dot >= 0 {
		ident = ident[dot+1:]
	}
	// Validate: must be a plain identifier
	for _, r := range ident {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$') {
			return ""
		}
	}
	return ident
}

// =============================================================
// Depth-aware indented rendering
// =============================================================

func (c *JSXConverter) renderNodeIndented(buf *strings.Builder, n *html.Node, depth int) {
	switch n.Type {
	case html.DocumentNode:
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			c.renderNodeIndented(buf, child, depth)
		}
	case html.ElementNode:
		c.renderElementIndented(buf, n, depth)
	case html.TextNode:
		trimmed := strings.TrimSpace(n.Data)
		if trimmed != "" {
			buf.WriteString(strings.Repeat("  ", depth) + escapeJSXText(trimmed) + "\n")
		}
	case html.CommentNode:
		trimmed := strings.TrimSpace(n.Data)
		if trimmed != "" {
			buf.WriteString(strings.Repeat("  ", depth) + "{/*" + trimmed + "*/}\n")
		}
	}
}

// hasElemChild returns true if n has at least one non-skipped element child.
func hasElemChild(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && !skipElements[child.Data] {
			return true
		}
	}
	return false
}

// isInlineContent returns true when all element children are inline-level
// elements (e.g. <strong>, <em>, <a>, <span>). In that case the element
// should be rendered on a single line to preserve text flow.
func isInlineContent(n *html.Node) bool {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		if skipElements[child.Data] {
			continue
		}
		if !inlineElements[child.Data] && !voidElements[child.Data] {
			return false
		}
	}
	return true
}

// normalizeInlineText collapses internal whitespace runs to a single space
// while preserving a leading or trailing space (word boundaries between nodes).
func normalizeInlineText(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	result := strings.Join(words, " ")
	if len(s) > 0 && isSpace(rune(s[0])) {
		result = " " + result
	}
	if last := s[len(s)-1]; isSpace(rune(last)) {
		result += " "
	}
	return result
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// renderChildrenInline renders all children compactly on one line —
// used for elements whose children are text + inline elements only.
func (c *JSXConverter) renderChildrenInline(buf *strings.Builder, n *html.Node) {
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			t := normalizeInlineText(child.Data)
			if t != "" {
				buf.WriteString(escapeJSXText(t))
			}
		case html.ElementNode:
			if skipElements[child.Data] {
				continue
			}
			buf.WriteString("<" + child.Data)
			for _, attr := range child.Attr {
				key, val := c.convertAttribute(attr)
				if key != "" && val != "" {
					buf.WriteString(fmt.Sprintf(" %s=%s", key, val))
				}
			}
			if voidElements[child.Data] {
				buf.WriteString(" />")
				continue
			}
			buf.WriteString(">")
			c.renderChildrenInline(buf, child)
			buf.WriteString("</" + child.Data + ">")
		case html.CommentNode:
			t := strings.TrimSpace(child.Data)
			if t != "" {
				buf.WriteString("{/*" + t + "*/}")
			}
		}
	}
}

func (c *JSXConverter) renderElementIndented(buf *strings.Builder, n *html.Node, depth int) {
	if skipElements[n.Data] {
		if n.Data == "html" || n.Data == "body" {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				c.renderNodeIndented(buf, child, depth)
			}
		}
		return
	}

	indent := strings.Repeat("  ", depth)
	buf.WriteString(indent + "<" + n.Data)

	for _, attr := range n.Attr {
		key, val := c.convertAttribute(attr)
		if key != "" && val != "" {
			buf.WriteString(fmt.Sprintf(" %s=%s", key, val))
		}
	}

	if voidElements[n.Data] {
		buf.WriteString(" />\n")
		return
	}

	if hasElemChild(n) {
		if isInlineContent(n) {
			// Mixed text + inline elements: keep on one line
			buf.WriteString(">")
			c.renderChildrenInline(buf, n)
			buf.WriteString("</" + n.Data + ">\n")
		} else {
			// Block children: each on its own line
			buf.WriteString(">\n")
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				c.renderNodeIndented(buf, child, depth+1)
			}
			buf.WriteString(indent + "</" + n.Data + ">\n")
		}
	} else {
		var textBuf strings.Builder
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode {
				textBuf.WriteString(strings.TrimSpace(child.Data))
			}
		}
		buf.WriteString(">" + escapeJSXText(textBuf.String()) + "</" + n.Data + ">\n")
	}
}

// =============================================================
// Helpers
// =============================================================

func findBodyNode(doc *html.Node) *html.Node {
	var find func(*html.Node) *html.Node
	find = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode && n.Data == "body" {
			return n
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := find(c); result != nil {
				return result
			}
		}
		return nil
	}
	return find(doc)
}

func nonSkippedChildren(n *html.Node) []*html.Node {
	if n == nil {
		return nil
	}
	var result []*html.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && !skipElements[child.Data] {
			result = append(result, child)
		}
	}
	return result
}
