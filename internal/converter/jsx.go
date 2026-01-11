package converter

import (
	"fmt"
	"htmlfmt/internal/analyzer"
	"htmlfmt/internal/fetcher"
	"regexp"
	"strings"
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

func (c *JSXConverter) convertHTMLToJSX(html string) (string, error) {
	html = c.cleanHTML(html)

	jsx := c.convertAttributes(html)

	jsx = c.convertSelfClosingTags(jsx)

	jsx = c.convertClassToClassName(jsx)

	jsx = c.convertStyleAttributes(jsx)

	jsx = c.convertEventHandlers(jsx)

	jsx = c.convertExternalResources(jsx)

	return jsx, nil
}

// cleanHTML removes unnecessary HTML structure
func (c *JSXConverter) cleanHTML(html string) string {
	// Remove DOCTYPE
	html = regexp.MustCompile(`<!DOCTYPE[^>]*>`).ReplaceAllString(html, "")

	// Remove html, head, body tags but keep their content
	html = regexp.MustCompile(`<html[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</html>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<head[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</head>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<body[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</body>`).ReplaceAllString(html, "")

	// Remove elements that shouldn't be in React component body
	html = regexp.MustCompile(`<title[^>]*>.*?</title>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<meta[^>]*/>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<meta[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<style[^>]*>.*?</style>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<script[^>]*>.*?</script>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<link[^>]*/>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<link[^>]*>`).ReplaceAllString(html, "")

	return strings.TrimSpace(html)
}

func (c *JSXConverter) convertAttributes(html string) string {
	html = regexp.MustCompile(`for="([^"]*)"`).ReplaceAllString(html, `htmlFor="$1"`)

	html = regexp.MustCompile(`tabindex="([^"]*)"`).ReplaceAllString(html, `tabIndex="$1"`)

	html = regexp.MustCompile(`readonly`).ReplaceAllString(html, `readOnly`)

	html = regexp.MustCompile(`checked="([^"]*)"`).ReplaceAllString(html, `checked={$1 === "checked"}`)
	html = regexp.MustCompile(`disabled="([^"]*)"`).ReplaceAllString(html, `disabled={$1 === "disabled"}`)
	html = regexp.MustCompile(`selected="([^"]*)"`).ReplaceAllString(html, `selected={$1 === "selected"}`)

	return html
}

func (c *JSXConverter) convertSelfClosingTags(html string) string {
	selfClosingTags := []string{"br", "hr", "img", "input", "meta", "link", "area", "base", "col", "embed", "source", "track", "wbr"}

	for _, tag := range selfClosingTags {
		pattern := fmt.Sprintf(`<%s([^>]*)>`, tag)
		replacement := fmt.Sprintf(`<%s$1 />`, tag)
		html = regexp.MustCompile(pattern).ReplaceAllString(html, replacement)
	}

	return html
}

// convertClassToClassName converts class attributes to className
func (c *JSXConverter) convertClassToClassName(html string) string {
	return regexp.MustCompile(`class="([^"]*)"`).ReplaceAllString(html, `className="$1"`)
}

func (c *JSXConverter) convertStyleAttributes(html string) string {
	stylePattern := `style="([^"]*)"`
	html = regexp.MustCompile(stylePattern).ReplaceAllStringFunc(html, func(match string) string {
		styleContent := regexp.MustCompile(`style="([^"]*)"`).FindStringSubmatch(match)[1]
		jsxStyle := c.convertStyleString(styleContent)
		return fmt.Sprintf(`style={%s}`, jsxStyle)
	})

	return html
}

// convertStyleString converts CSS style string to JSX style object
func (c *JSXConverter) convertStyleString(style string) string {
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

func (c *JSXConverter) convertEventHandlers(html string) string {
	html = regexp.MustCompile(`onclick="([^"]*)"`).ReplaceAllString(html, `onClick={() => { $1 }}`)
	html = regexp.MustCompile(`onchange="([^"]*)"`).ReplaceAllString(html, `onChange={() => { $1 }}`)
	html = regexp.MustCompile(`onsubmit="([^"]*)"`).ReplaceAllString(html, `onSubmit={() => { $1 }}`)
	html = regexp.MustCompile(`onload="([^"]*)"`).ReplaceAllString(html, `onLoad={() => { $1 }}`)

	return html
}

func (c *JSXConverter) convertExternalResources(html string) string {
	html = regexp.MustCompile(`<link[^>]*rel="stylesheet"[^>]*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<script[^>]*src="[^"]*"[^>]*></script>`).ReplaceAllString(html, "")

	return html
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
			jsCode.WriteString(fmt.Sprintf("\n", jsFile.Filename))
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
