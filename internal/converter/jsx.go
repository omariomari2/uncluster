package converter

import (
	"fmt"
	"htmlfmt/internal/analyzer"
	"htmlfmt/internal/fetcher"
	"strings"

	"golang.org/x/net/html"
)

type JSXConverter struct {
	ExternalCSS []fetcher.FetchedResource
	ExternalJS  []fetcher.FetchedResource
}

func ConvertToJSX(html, css, js string, externalCSS []fetcher.FetchedResource, externalJS []fetcher.FetchedResource) (string, error) {
	converter := &JSXConverter{
		ExternalCSS: externalCSS,
		ExternalJS:  externalJS,
	}

	jsx, err := converter.convertHTMLToJSX(html)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to JSX: %w", err)
	}

	cssImports := converter.generateCSSImports(css)
	jsCode := converter.generateJSCode(js)

	component := fmt.Sprintf(`import React from 'react'
%s

function MainComponent() {
  return (
    <>
      %s
    </>
  )
}

%s

export default MainComponent
`, cssImports, jsx, jsCode)

	return component, nil
}

func (c *JSXConverter) convertHTMLToJSX(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var buf strings.Builder
	c.renderNodeAsJSX(&buf, doc)
	return buf.String(), nil
}

func (c *JSXConverter) renderNodeAsJSX(buf *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.DocumentNode:
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			c.renderNodeAsJSX(buf, child)
		}
	case html.ElementNode:
		c.renderElementAsJSX(buf, n)
	case html.TextNode:
		c.renderTextAsJSX(buf, n)
	case html.CommentNode:
		c.renderCommentAsJSX(buf, n)
	}
}

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
	"fill-rule":                    "fillRule",
	"clip-rule":                    "clipRule",
	"clip-path":                    "clipPath",
	"stroke-width":                 "strokeWidth",
	"stroke-linecap":               "strokeLinecap",
	"stroke-linejoin":              "strokeLinejoin",
	"stroke-miterlimit":            "strokeMiterlimit",
	"stroke-dasharray":             "strokeDasharray",
	"stroke-dashoffset":            "strokeDashoffset",
	"fill-opacity":                 "fillOpacity",
	"stroke-opacity":               "strokeOpacity",
	"text-anchor":                  "textAnchor",
	"font-family":                  "fontFamily",
	"font-size":                    "fontSize",
	"font-weight":                  "fontWeight",
	"font-style":                   "fontStyle",
	"text-decoration":              "textDecoration",
	"letter-spacing":               "letterSpacing",
	"word-spacing":                 "wordSpacing",
	"dominant-baseline":            "dominantBaseline",
	"alignment-baseline":           "alignmentBaseline",
	"baseline-shift":               "baselineShift",
	"vector-effect":                "vectorEffect",
	"paint-order":                  "paintOrder",
	"shape-rendering":              "shapeRendering",
	"image-rendering":              "imageRendering",
	"color-rendering":              "colorRendering",
	"color-interpolation":          "colorInterpolation",
	"color-interpolation-filters":  "colorInterpolationFilters",
	"flood-color":                  "floodColor",
	"flood-opacity":                "floodOpacity",
	"lighting-color":               "lightingColor",
	"writing-mode":                 "writingMode",
	"pointer-events":               "pointerEvents",
	"unicode-bidi":                 "unicodeBidi",
	"stop-color":                   "stopColor",
	"stop-opacity":                 "stopOpacity",
	"marker-start":                 "markerStart",
	"marker-mid":                   "markerMid",
	"marker-end":                   "markerEnd",
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

func (c *JSXConverter) renderElementAsJSX(buf *strings.Builder, n *html.Node) {
	if skipElements[n.Data] {
		if n.Data == "html" || n.Data == "body" {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				c.renderNodeAsJSX(buf, child)
			}
		}
		return
	}

	buf.WriteString("<")
	buf.WriteString(n.Data)

	for _, attr := range n.Attr {
		key, val := c.convertAttribute(attr)
		if key != "" && val != "" {
			buf.WriteString(fmt.Sprintf(" %s=%s", key, val))
		}
	}

	if voidElements[n.Data] {
		buf.WriteString(" />")
		return
	}

	buf.WriteString(">")

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		c.renderNodeAsJSX(buf, child)
	}

	buf.WriteString("</")
	buf.WriteString(n.Data)
	buf.WriteString(">")
}

func (c *JSXConverter) convertAttribute(attr html.Attribute) (string, string) {
	key := attr.Key
	val := attr.Val

	// xlink:href (deprecated but common in SVGs) → href
	if attr.Namespace == "xlink" && key == "href" {
		return "href", fmt.Sprintf(`"%s"`, val)
	}
	// Drop namespace attributes that React doesn't need
	if attr.Namespace != "" {
		return "", ""
	}

	if jsxKey, ok := jsxAttributeMap[key]; ok {
		key = jsxKey
	}

	if jsxEvent, ok := jsxEventMap[key]; ok {
		// Extract simple function name for the TODO comment (best-effort).
		return jsxEvent, fmt.Sprintf("{() => { %s }}", val)
	}

	if key == "style" {
		return "style", c.convertStyleToObject(val)
	}

	if key == "checked" || key == "disabled" || key == "selected" {
		if val == key || val == "true" {
			return key, "{true}"
		}
		return key, "{false}"
	}

	return key, fmt.Sprintf(`"%s"`, val)
}

func (c *JSXConverter) convertStyleToObject(style string) string {
	styles := strings.Split(style, ";")
	var jsxStyles []string

	for _, s := range styles {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		key = c.kebabToCamel(key)
		jsxStyles = append(jsxStyles, fmt.Sprintf("%s: '%s'", key, value))
	}

	return fmt.Sprintf("{%s}", strings.Join(jsxStyles, ", "))
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

func (c *JSXConverter) renderTextAsJSX(buf *strings.Builder, n *html.Node) {
	text := n.Data

	if strings.Contains(text, "<!--") && strings.Contains(text, "-->") {
		text = convertHTMLCommentsInText(text)
	}

	trimmed := strings.TrimSpace(text)
	if trimmed != "" {
		buf.WriteString(trimmed)
	}
}

func convertHTMLCommentsInText(text string) string {
	result := text
	start := 0
	for {
		commentStart := strings.Index(result[start:], "<!--")
		if commentStart == -1 {
			break
		}
		commentStart += start
		commentEnd := strings.Index(result[commentStart:], "-->")
		if commentEnd == -1 {
			break
		}
		commentEnd += commentStart + 3

		commentContent := result[commentStart+4 : commentEnd-3]

		jsxComment := "{/*" + commentContent + "*/}"
		result = result[:commentStart] + jsxComment + result[commentEnd:]
		start = commentStart + len(jsxComment)
	}
	return result
}

func (c *JSXConverter) renderCommentAsJSX(buf *strings.Builder, n *html.Node) {
	buf.WriteString("{/*")
	buf.WriteString(n.Data)
	buf.WriteString("*/}")
}

func (c *JSXConverter) generateCSSImports(css string) string {
	var imports []string

	if css != "" {
		imports = append(imports, `import '../styles/main.css'`)
	}

	for _, cssFile := range c.ExternalCSS {
		if cssFile.Error == nil {
			imports = append(imports, fmt.Sprintf(`import '../styles/external/%s'`, cssFile.Filename))
		}
	}

	return strings.Join(imports, "\n")
}

func (c *JSXConverter) generateJSCode(js string) string {
	var jsCode strings.Builder

	if js != "" {
		jsCode.WriteString("\n")
		jsCode.WriteString(js)
		jsCode.WriteString("\n")
	}

	for _, jsFile := range c.ExternalJS {
		if jsFile.Error == nil {
			jsCode.WriteString("\n")
			jsCode.WriteString(jsFile.Content)
			jsCode.WriteString("\n")
		}
	}

	return jsCode.String()
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

	// Detect repeated list patterns and generate typed component.
	if pattern := detectListPattern(body); pattern != nil {
		return buildListComponentTSX(componentName, pattern, c, body), nil
	}

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
			buf.WriteString(strings.Repeat("  ", depth) + trimmed + "\n")
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
				buf.WriteString(t)
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
		buf.WriteString(">" + textBuf.String() + "</" + n.Data + ">\n")
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

func jsxGetAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}

func jsxHasClass(n *html.Node, class string) bool {
	for _, c := range strings.Fields(jsxGetAttr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

func jsxTextContent(n *html.Node) string {
	var buf strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			buf.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(buf.String())
}

func jsxFindFirst(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if result := jsxFindFirst(child, tag); result != nil {
			return result
		}
	}
	return nil
}

func jsxFindBgURL(n *html.Node) string {
	if n.Type == html.ElementNode {
		style := jsxGetAttr(n, "style")
		if strings.Contains(style, "background-image") {
			start := strings.Index(style, "url(")
			if start >= 0 {
				rest := style[start+4:]
				rest = strings.TrimLeft(rest, "'\" ")
				end := strings.IndexAny(rest, "'\")")
				if end > 0 {
					return rest[:end]
				}
			}
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if url := jsxFindBgURL(child); url != "" {
			return url
		}
	}
	return ""
}

// =============================================================
// List pattern detection
// =============================================================

type listField struct {
	Name   string
	TSType string
	Values []string
}

type listPattern struct {
	Wrapper *html.Node
	Items   []*html.Node
	Fields  []listField
}

func detectListPattern(body *html.Node) *listPattern {
	if body == nil {
		return nil
	}
	return findListInSubtree(body, 0)
}

func findListInSubtree(n *html.Node, depth int) *listPattern {
	if n == nil || depth > 8 || n.Type != html.ElementNode {
		return nil
	}

	items := collectRepeatedItems(n)
	if len(items) >= 2 {
		fields := extractListFields(items)
		if len(fields) > 0 {
			return &listPattern{Wrapper: n, Items: items, Fields: fields}
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if p := findListInSubtree(child, depth+1); p != nil {
			return p
		}
	}
	return nil
}

func collectRepeatedItems(n *html.Node) []*html.Node {
	var children []*html.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			children = append(children, child)
		}
	}
	if len(children) < 2 {
		return nil
	}

	// Webflow w-dyn-items pattern
	class := jsxGetAttr(n, "class")
	if strings.Contains(class, "w-dyn-items") || strings.Contains(class, "w-dyn-list") {
		var items []*html.Node
		for _, child := range children {
			if jsxHasClass(child, "w-dyn-item") || jsxGetAttr(child, "role") == "listitem" {
				items = append(items, child)
			}
		}
		if len(items) >= 2 {
			return items
		}
	}

	// role="listitem" siblings
	if jsxGetAttr(children[0], "role") == "listitem" {
		var items []*html.Node
		for _, child := range children {
			if jsxGetAttr(child, "role") == "listitem" {
				items = append(items, child)
			}
		}
		if len(items) >= 2 {
			return items
		}
	}

	// ul/ol with li children
	if n.Data == "ul" || n.Data == "ol" {
		var items []*html.Node
		for _, child := range children {
			if child.Data == "li" {
				items = append(items, child)
			}
		}
		if len(items) >= 2 {
			return items
		}
	}

	// Same tag + same class (3+ for confidence)
	first := children[0]
	firstClass := jsxGetAttr(first, "class")
	if firstClass != "" {
		var items []*html.Node
		for _, child := range children {
			if child.Data == first.Data && jsxGetAttr(child, "class") == firstClass {
				items = append(items, child)
			}
		}
		if len(items) >= 3 {
			return items
		}
	}

	return nil
}

type fieldExtractor struct {
	name    string
	tsType  string
	extract func(*html.Node) string
}

func buildFieldExtractors() []fieldExtractor {
	return []fieldExtractor{
		{
			name: "href", tsType: "string",
			extract: func(n *html.Node) string {
				a := jsxFindFirst(n, "a")
				if a == nil {
					return ""
				}
				return jsxGetAttr(a, "href")
			},
		},
		{
			name: "imageSrc", tsType: "string",
			extract: func(n *html.Node) string {
				img := jsxFindFirst(n, "img")
				if img == nil {
					return ""
				}
				return jsxGetAttr(img, "src")
			},
		},
		{
			name: "imageAlt", tsType: "string",
			extract: func(n *html.Node) string {
				img := jsxFindFirst(n, "img")
				if img == nil {
					return ""
				}
				return jsxGetAttr(img, "alt")
			},
		},
		{
			name: "title", tsType: "string",
			extract: func(n *html.Node) string {
				for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
					h := jsxFindFirst(n, tag)
					if h != nil {
						t := jsxTextContent(h)
						if t != "" {
							return t
						}
					}
				}
				return ""
			},
		},
		{
			name: "description", tsType: "string",
			extract: func(n *html.Node) string {
				p := jsxFindFirst(n, "p")
				if p == nil {
					return ""
				}
				return jsxTextContent(p)
			},
		},
		{
			name: "label", tsType: "string",
			extract: func(n *html.Node) string {
				a := jsxFindFirst(n, "a")
				if a == nil {
					return ""
				}
				return jsxTextContent(a)
			},
		},
		{
			name: "backgroundImage", tsType: "string",
			extract: func(n *html.Node) string {
				return jsxFindBgURL(n)
			},
		},
	}
}

func extractListFields(items []*html.Node) []listField {
	if len(items) == 0 {
		return nil
	}

	extractors := buildFieldExtractors()
	var fields []listField
	seen := make(map[string]bool)

	for _, ext := range extractors {
		val0 := ext.extract(items[0])
		if val0 == "" {
			continue
		}

		values := make([]string, len(items))
		values[0] = val0
		allSame := true
		for i := 1; i < len(items); i++ {
			values[i] = ext.extract(items[i])
			if values[i] != val0 {
				allSame = false
			}
		}

		if allSame && len(items) > 1 {
			continue
		}
		if seen[ext.name] {
			continue
		}
		seen[ext.name] = true

		fields = append(fields, listField{
			Name:   ext.name,
			TSType: ext.tsType,
			Values: values,
		})
	}

	return fields
}

// =============================================================
// List component TSX builder
// =============================================================

func buildListComponentTSX(componentName string, pattern *listPattern, c *JSXConverter, body *html.Node) string {
	typeName := componentName + "Item"

	// value → field reference (without braces) for substitution
	fieldSubs := make(map[string]string)
	for _, field := range pattern.Fields {
		if len(field.Values) > 0 && field.Values[0] != "" {
			if field.Name == "backgroundImage" {
				fieldSubs[field.Values[0]] = "item.backgroundImage"
			} else {
				fieldSubs[field.Values[0]] = "item." + field.Name
			}
		}
	}

	// TypeScript interface
	var iface strings.Builder
	iface.WriteString(fmt.Sprintf("interface %s {\n", typeName))
	for _, f := range pattern.Fields {
		iface.WriteString(fmt.Sprintf("  %s: %s\n", f.Name, f.TSType))
	}
	iface.WriteString("}\n")

	// Data array
	var data strings.Builder
	data.WriteString(fmt.Sprintf("const items: %s[] = [\n", typeName))
	for i := range pattern.Items {
		data.WriteString("  {\n")
		for _, f := range pattern.Fields {
			val := ""
			if i < len(f.Values) {
				val = f.Values[i]
			}
			data.WriteString(fmt.Sprintf("    %s: %q,\n", f.Name, val))
		}
		data.WriteString("  },\n")
	}
	data.WriteString("]\n")

	// Outer structure with map injection at wrapper node.
	// Item depth is determined dynamically by renderWithListMap.
	roots := nonSkippedChildren(body)
	var bodyBuf strings.Builder
	for _, root := range roots {
		c.renderWithListMap(&bodyBuf, root, 2, pattern, fieldSubs)
	}
	bodyJSX := strings.TrimRight(bodyBuf.String(), "\n")

	var returnExpr string
	if len(roots) == 1 {
		returnExpr = fmt.Sprintf("(\n%s\n  )", bodyJSX)
	} else {
		returnExpr = fmt.Sprintf("(\n    <>\n%s\n    </>)", bodyJSX)
	}

	return fmt.Sprintf(`import React from 'react'

%s
%s
function %s(): JSX.Element {
  return %s
}

export default %s
`, iface.String(), data.String(), componentName, returnExpr, componentName)
}

// renderWithListMap renders the tree normally but replaces the list wrapper's
// children with a {items.map(...)} expression. The item template is rendered
// inline at depth+2 (inside the map call) for correct indentation.
func (c *JSXConverter) renderWithListMap(
	buf *strings.Builder, n *html.Node, depth int,
	pattern *listPattern, fieldSubs map[string]string,
) {
	if n == nil || n.Type != html.ElementNode {
		return
	}
	if skipElements[n.Data] {
		if n.Data == "html" || n.Data == "body" {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				c.renderWithListMap(buf, child, depth, pattern, fieldSubs)
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

	// Replace this node's children with the map expression.
	// Item renders at depth+2: depth+1 for inside map, then the element itself.
	if n == pattern.Wrapper {
		buf.WriteString(">\n")
		mapIndent := strings.Repeat("  ", depth+1)
		buf.WriteString(mapIndent + "{items.map((item, index) => (\n")
		c.renderElemWithSubs(buf, pattern.Items[0], depth+2, fieldSubs, true)
		buf.WriteString(mapIndent + "))}\n")
		buf.WriteString(indent + "</" + n.Data + ">\n")
		return
	}

	if hasElemChild(n) {
		buf.WriteString(">\n")
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			c.renderWithListMap(buf, child, depth+1, pattern, fieldSubs)
		}
		buf.WriteString(indent + "</" + n.Data + ">\n")
	} else {
		var textBuf strings.Builder
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode {
				textBuf.WriteString(strings.TrimSpace(child.Data))
			}
		}
		buf.WriteString(">" + textBuf.String() + "</" + n.Data + ">\n")
	}
}

// renderElemWithSubs renders an item element substituting dynamic field values.
func (c *JSXConverter) renderElemWithSubs(buf *strings.Builder, n *html.Node, depth int, fieldSubs map[string]string, isRoot bool) {
	if n == nil || n.Type != html.ElementNode || skipElements[n.Data] {
		return
	}

	indent := strings.Repeat("  ", depth)
	buf.WriteString(indent + "<" + n.Data)

	for _, attr := range n.Attr {
		key, val := c.convertAttrWithSubs(attr, fieldSubs)
		if key != "" && val != "" {
			buf.WriteString(fmt.Sprintf(" %s=%s", key, val))
		}
	}

	// Add key prop at the root item level.
	if isRoot {
		buf.WriteString(" key={index}")
	}

	if voidElements[n.Data] {
		buf.WriteString(" />\n")
		return
	}

	if hasElemChild(n) {
		buf.WriteString(">\n")
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			c.renderNodeWithSubs(buf, child, depth+1, fieldSubs)
		}
		buf.WriteString(indent + "</" + n.Data + ">\n")
	} else {
		var textBuf strings.Builder
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.TextNode {
				textBuf.WriteString(strings.TrimSpace(child.Data))
			}
		}
		text := textBuf.String()
		if ref, ok := fieldSubs[text]; ok {
			buf.WriteString(">{" + ref + "}</" + n.Data + ">\n")
		} else {
			buf.WriteString(">" + text + "</" + n.Data + ">\n")
		}
	}
}

func (c *JSXConverter) renderNodeWithSubs(buf *strings.Builder, n *html.Node, depth int, fieldSubs map[string]string) {
	switch n.Type {
	case html.ElementNode:
		c.renderElemWithSubs(buf, n, depth, fieldSubs, false)
	case html.TextNode:
		trimmed := strings.TrimSpace(n.Data)
		if trimmed == "" {
			return
		}
		indent := strings.Repeat("  ", depth)
		if ref, ok := fieldSubs[trimmed]; ok {
			buf.WriteString(indent + "{" + ref + "}\n")
		} else {
			buf.WriteString(indent + trimmed + "\n")
		}
	}
}

// convertAttrWithSubs converts an attribute, substituting known field values.
func (c *JSXConverter) convertAttrWithSubs(attr html.Attribute, fieldSubs map[string]string) (string, string) {
	key := attr.Key
	rawVal := attr.Val

	if jsxKey, ok := jsxAttributeMap[key]; ok {
		key = jsxKey
	}

	if jsxEvent, ok := jsxEventMap[key]; ok {
		return jsxEvent, fmt.Sprintf("{() => { %s }}", rawVal)
	}

	if key == "style" {
		return "style", c.convertStyleWithSubs(rawVal, fieldSubs)
	}

	if key == "checked" || key == "disabled" || key == "selected" {
		if rawVal == key || rawVal == "true" {
			return key, "{true}"
		}
		return key, "{false}"
	}

	if ref, ok := fieldSubs[rawVal]; ok {
		return key, "{" + ref + "}"
	}

	return key, fmt.Sprintf(`"%s"`, rawVal)
}

// convertStyleWithSubs converts a CSS style string, substituting field values.
func (c *JSXConverter) convertStyleWithSubs(style string, fieldSubs map[string]string) string {
	styles := strings.Split(style, ";")
	var jsxStyles []string

	for _, s := range styles {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, ":", 2)
		if len(parts) != 2 {
			continue
		}
		cssKey := strings.TrimSpace(parts[0])
		cssVal := strings.TrimSpace(parts[1])
		camelKey := c.kebabToCamel(cssKey)

		if cssKey == "background-image" {
			start := strings.Index(cssVal, "url(")
			if start >= 0 {
				rest := cssVal[start+4:]
				rest = strings.TrimLeft(rest, "'\" ")
				end := strings.IndexAny(rest, "'\")")
				if end > 0 {
					urlVal := rest[:end]
					if _, ok := fieldSubs[urlVal]; ok {
						jsxStyles = append(jsxStyles, camelKey+": `url(${item.backgroundImage})`")
						continue
					}
				}
			}
		}

		substituted := false
		for origVal, ref := range fieldSubs {
			if cssVal == origVal {
				jsxStyles = append(jsxStyles, fmt.Sprintf("%s: {%s}", camelKey, ref))
				substituted = true
				break
			}
		}
		if !substituted {
			jsxStyles = append(jsxStyles, fmt.Sprintf("%s: '%s'", camelKey, cssVal))
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(jsxStyles, ", "))
}

func AnalyzeAndConvert(html string) ([]string, error) {
	suggestions, err := analyzer.AnalyzeComponents(html)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze HTML: %w", err)
	}

	var components []string

	for _, suggestion := range suggestions {
		componentName := suggestion.Name
		componentName = strings.Title(strings.ReplaceAll(componentName, "-", " "))
		componentName = strings.ReplaceAll(componentName, " ", "")

		if suggestion.JSXCode != "" {
			component := fmt.Sprintf(`import React from 'react'

%s`, suggestion.JSXCode)
			components = append(components, component)
			continue
		}

		jsx := fmt.Sprintf(`<div className="%s">
  {/* %s */}
</div>`, suggestion.TagName, suggestion.Description)

		component := fmt.Sprintf(`import React from 'react'

interface %sProps {
}

function %s(props: %sProps) {
  return (
    <>
      %s
    </>
  )
}

export default %s
`, componentName, componentName, componentName, jsx, componentName)

		components = append(components, component)
	}

	return components, nil
}
