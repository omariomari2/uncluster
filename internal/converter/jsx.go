package converter

import (
	"encoding/json"
	"fmt"
	"htmlfmt/internal/analyzer"
	"htmlfmt/internal/fetcher"
	"os"
	"strings"
	"time"

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

	// Convert HTML to JSX
	jsx, err := converter.convertHTMLToJSX(html)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to JSX: %w", err)
	}

	// Add CSS imports
	cssImports := converter.generateCSSImports(css)

	// Add JS functionality
	jsCode := converter.generateJSCode(js)

	// Combine everything
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
	// #region agent log
	logDebugJSX("convertHTMLToJSX", "entry", map[string]interface{}{"htmlLength": len(htmlContent)}, "B")
	// #endregion
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// #region agent log
		logDebugJSX("convertHTMLToJSX", "parse_error", map[string]interface{}{"error": err.Error()}, "B")
		// #endregion
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var buf strings.Builder
	c.renderNodeAsJSX(&buf, doc)
	result := buf.String()
	// #region agent log
	hasHTMLComments := strings.Contains(result, "<!--")
	hasClassAttr := strings.Contains(result, `class="`)
	hasDoubleSlash := strings.Contains(result, "/ />")
	logDebugJSX("convertHTMLToJSX", "result", map[string]interface{}{
		"resultLength":   len(result),
		"hasHTMLComments": hasHTMLComments,
		"hasClassAttr":    hasClassAttr,
		"hasDoubleSlash":  hasDoubleSlash,
		"first300Chars":   safeSubstringJSX(result, 0, 300),
	}, "B")
	// #endregion
	return result, nil
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
	"class":          "className",
	"for":            "htmlFor",
	"tabindex":       "tabIndex",
	"readonly":       "readOnly",
	"maxlength":      "maxLength",
	"cellpadding":    "cellPadding",
	"cellspacing":    "cellSpacing",
	"colspan":        "colSpan",
	"rowspan":        "rowSpan",
	"frameborder":    "frameBorder",
	"allowfullscreen": "allowFullScreen",
	"fill-rule":      "fillRule",
	"clip-rule":      "clipRule",
	"stroke-width":   "strokeWidth",
	"stroke-linecap":  "strokeLinecap",
	"stroke-linejoin": "strokeLinejoin",
	"fill-opacity":   "fillOpacity",
	"stroke-opacity": "strokeOpacity",
	"text-anchor":    "textAnchor",
	"font-family":    "fontFamily",
	"font-size":      "fontSize",
	"font-weight":    "fontWeight",
	"text-decoration": "textDecoration",
}

var jsxEventMap = map[string]string{
	"onclick":   "onClick",
	"onchange":  "onChange",
	"onsubmit":  "onSubmit",
	"onload":    "onLoad",
	"onerror":   "onError",
	"onkeydown": "onKeyDown",
	"onkeyup":   "onKeyUp",
	"onkeypress": "onKeyPress",
	"onfocus":   "onFocus",
	"onblur":    "onBlur",
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

	if jsxKey, ok := jsxAttributeMap[key]; ok {
		key = jsxKey
	}

	if jsxEvent, ok := jsxEventMap[key]; ok {
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

// kebabToCamel converts kebab-case to camelCase
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
	// #region agent log
	logDebugJSX("renderTextAsJSX", "entry", map[string]interface{}{
		"textLength": len(text),
		"hasHTMLComment": strings.Contains(text, "<!--"),
		"first100Chars": safeSubstringJSX(text, 0, 100),
	}, "B")
	// #endregion
	
	// Check if text contains HTML comments and convert them
	if strings.Contains(text, "<!--") && strings.Contains(text, "-->") {
		// Convert HTML comments to JSX comments
		text = convertHTMLCommentsInText(text)
		// #region agent log
		logDebugJSX("renderTextAsJSX", "converted_comments", map[string]interface{}{
			"convertedText": safeSubstringJSX(text, 0, 200),
		}, "B")
		// #endregion
	}
	
	trimmed := strings.TrimSpace(text)
	if trimmed != "" {
		buf.WriteString(trimmed)
	}
}

// convertHTMLCommentsInText converts HTML comments <!-- --> to JSX comments {/* */}
func convertHTMLCommentsInText(text string) string {
	// Use regex to find and replace HTML comments
	// Pattern: <!-- ... -->
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
		
		// Extract comment content (between <!-- and -->)
		commentContent := result[commentStart+4 : commentEnd-3]
		
		// Replace with JSX comment
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

// AnalyzeAndConvert analyzes HTML and converts to optimized JSX components
func AnalyzeAndConvert(html string) ([]string, error) {
	// Use existing analyzer to get component suggestions
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

		// Fallback: create basic JSX from the component info
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

func logDebugJSX(location, message string, data map[string]interface{}, hypothesisId string) {
	logEntry := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": hypothesisId,
		"location":     "jsx.go:" + location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	jsonData, _ := json.Marshal(logEntry)
	logPath := ".cursor/debug.log"
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintln(f, string(jsonData))
		f.Close()
	}
}

func safeSubstringJSX(s string, start, end int) string {
	if start >= len(s) {
		return ""
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}
