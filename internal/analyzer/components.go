package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type ComponentSuggestion struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	TagName     string            `json:"tagName"`
	Attributes  map[string]string `json:"attributes"`
	Children    []string          `json:"children"`
	Count       int               `json:"count"`
	JSXCode     string            `json:"jsxCode"`
}

func AnalyzeComponents(htmlInput string) ([]ComponentSuggestion, error) {
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	elementPatterns := make(map[string]*ElementPattern)
	collectPatterns(doc, elementPatterns)

	return generateSuggestionsWithoutAI(elementPatterns), nil
}

type ElementPattern struct {
	TagName    string
	Attributes map[string]int
	Children   map[string]int
	Count      int
	Examples   []*html.Node
}

func collectPatterns(n *html.Node, patterns map[string]*ElementPattern) {
	if n.Type == html.ElementNode {
		patternKey := generatePatternKey(n)

		if patterns[patternKey] == nil {
			patterns[patternKey] = &ElementPattern{
				TagName:    n.Data,
				Attributes: make(map[string]int),
				Children:   make(map[string]int),
				Count:      0,
				Examples:   []*html.Node{},
			}
		}

		pattern := patterns[patternKey]
		pattern.Count++

		for _, attr := range n.Attr {
			pattern.Attributes[attr.Key]++
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				pattern.Children[c.Data]++
			}
		}

		if len(pattern.Examples) < 3 {
			pattern.Examples = append(pattern.Examples, n)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectPatterns(c, patterns)
	}
}

func generatePatternKey(n *html.Node) string {
	key := n.Data

	classes := getAttributeValue(n, "class")
	if classes != "" {
		key += "." + strings.ReplaceAll(classes, " ", ".")
	}

	id := getAttributeValue(n, "id")
	if id != "" {
		key += "#" + id
	}

	return key
}

func getAttributeValue(n *html.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

func convertToValidPropName(attr string) string {
	if attr == "class" {
		return "className"
	}

	if strings.HasPrefix(attr, "data-") {
		rest := strings.TrimPrefix(attr, "data-")
		if rest == "" {
			return "data"
		}
		camelRest := kebabToCamel(rest)
		if len(camelRest) > 0 {
			return "data" + strings.ToUpper(camelRest[:1]) + camelRest[1:]
		}
		return "data"
	}
	if strings.HasPrefix(attr, "aria-") {
		rest := strings.TrimPrefix(attr, "aria-")
		if rest == "" {
			return "aria"
		}
		camelRest := kebabToCamel(rest)
		if len(camelRest) > 0 {
			return "aria" + strings.ToUpper(camelRest[:1]) + camelRest[1:]
		}
		return "aria"
	}

	parts := strings.Split(attr, "-")
	if len(parts) == 1 {
		return attr
	}

	return kebabToCamel(attr)
}

func kebabToCamel(s string) string {
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

func generateSuggestionsWithoutAI(patterns map[string]*ElementPattern) []ComponentSuggestion {
	var suggestions []ComponentSuggestion

	obviousPatterns := map[string]bool{
		"card": true, "button": true, "btn": true,
		"nav-item": true, "menu-item": true, "list-item": true,
		"modal": true, "dialog": true, "popup": true,
		"form-field": true, "input-group": true,
		"tab": true, "accordion": true, "dropdown": true,
		"badge": true, "tag": true, "chip": true,
		"avatar": true, "thumbnail": true,
		"alert": true, "toast": true, "notification": true,
	}

	structuralElements := map[string]bool{
		"html": true, "head": true, "body": true, "title": true,
		"meta": true, "link": true, "script": true, "style": true,
		"base": true, "noscript": true,
	}

	for patternKey, pattern := range patterns {
		if structuralElements[pattern.TagName] {
			continue
		}

		if !matchesObviousPattern(patternKey, obviousPatterns) {
			continue
		}

		if pattern.Count < 3 {
			continue
		}

		if isStructuralElement(pattern.TagName) {
			continue
		}

		suggestion := ComponentSuggestion{
			Name:        generateComponentName(pattern.TagName, patternKey),
			Description: generateDescription(pattern),
			TagName:     pattern.TagName,
			Attributes:  make(map[string]string),
			Children:    make([]string, 0),
			Count:       pattern.Count,
			JSXCode:     generateJSXCode(pattern),
		}

		for attr, count := range pattern.Attributes {
			if count >= pattern.Count/2 {
				suggestion.Attributes[attr] = "{string}"
			}
		}

		for childTag, count := range pattern.Children {
			if count >= pattern.Count/2 {
				suggestion.Children = append(suggestion.Children, childTag)
			}
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

func matchesObviousPattern(patternKey string, patterns map[string]bool) bool {
	lowerKey := strings.ToLower(patternKey)
	for pattern := range patterns {
		if strings.Contains(lowerKey, pattern) {
			return true
		}
	}
	return false
}

func isStructuralElement(tagName string) bool {
	structural := map[string]bool{
		"div": true, "span": true, "section": true, "article": true,
		"header": true, "footer": true, "main": true, "aside": true,
		"p": true, "a": true, "ul": true, "ol": true, "li": true,
	}
	return structural[tagName]
}

func generateComponentName(tagName, patternKey string) string {
	name := strings.Title(tagName)

	if strings.Contains(patternKey, "card") {
		name += "Card"
	} else if strings.Contains(patternKey, "button") {
		name += "Button"
	} else if strings.Contains(patternKey, "nav") {
		name += "Item"
	} else if strings.Contains(patternKey, "list") {
		name += "Item"
	} else if strings.Contains(patternKey, "form") {
		name += "Field"
	} else {
		name += "Component"
	}

	return name
}

func generateDescription(pattern *ElementPattern) string {
	desc := fmt.Sprintf("A reusable %s component", pattern.TagName)

	if pattern.Count > 1 {
		desc += fmt.Sprintf(" (appears %d times)", pattern.Count)
	}

	if len(pattern.Attributes) > 0 {
		desc += " with configurable attributes"
	}

	if len(pattern.Children) > 0 {
		desc += " and child elements"
	}

	return desc
}

func generateJSXCode(pattern *ElementPattern) string {
	if len(pattern.Examples) == 0 {
		return ""
	}

	example := pattern.Examples[0]
	var buf strings.Builder

	componentName := generateComponentName(pattern.TagName, generatePatternKey(example))
	buf.WriteString(fmt.Sprintf("const %s = ({ ", componentName))

	props := []string{}
	propMap := make(map[string]string)
	for attr, count := range pattern.Attributes {
		if count >= pattern.Count/2 {
			propName := convertToValidPropName(attr)
			props = append(props, propName)
			propMap[attr] = propName
		}
	}

	if len(props) > 0 {
		buf.WriteString(strings.Join(props, ", "))
	}

	buf.WriteString(" }) => {\n")
	buf.WriteString("\treturn (\n")

	buf.WriteString(fmt.Sprintf("\t\t<%s", pattern.TagName))

	for attr, count := range pattern.Attributes {
		if count >= pattern.Count/2 {
			propName := propMap[attr]
			jsxAttr := attr
			if attr == "class" {
				jsxAttr = "className"
			}
			buf.WriteString(fmt.Sprintf(" %s={%s}", jsxAttr, propName))
		}
	}

	buf.WriteString(">\n")
	buf.WriteString("\t\t\t\n")
	buf.WriteString(fmt.Sprintf("\t\t</%s>\n", pattern.TagName))
	buf.WriteString("\t);\n")
	buf.WriteString("};\n\n")
	buf.WriteString("export default " + generateComponentName(pattern.TagName, generatePatternKey(example)) + ";")

	return buf.String()
}

func nodeToHTML(n *html.Node) string {
	var buf strings.Builder
	renderNode(&buf, n)
	return buf.String()
}

func renderNode(buf *strings.Builder, n *html.Node) {
	if n == nil {
		return
	}

	switch n.Type {
	case html.ElementNode:
		buf.WriteString("<")
		buf.WriteString(n.Data)

		for _, attr := range n.Attr {
			buf.WriteString(" ")
			buf.WriteString(attr.Key)
			if attr.Val != "" {
				buf.WriteString(`="`)
				buf.WriteString(attr.Val)
				buf.WriteString(`"`)
			}
		}

		if isVoidElement(n.Data) {
			buf.WriteString(" />")
			return
		}

		buf.WriteString(">")

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			renderNode(buf, c)
		}

		buf.WriteString("</")
		buf.WriteString(n.Data)
		buf.WriteString(">")

	case html.TextNode:
		buf.WriteString(n.Data)
	}
}

func isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return voidElements[strings.ToLower(tagName)]
}

func generateJSXCodeWithName(pattern *ElementPattern, componentName string) string {
	if len(pattern.Examples) == 0 {
		return ""
	}

	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("const %s = ({ ", componentName))

	props := []string{}
	propMap := make(map[string]string)
	for attr, count := range pattern.Attributes {
		if count >= pattern.Count/2 {
			propName := convertToValidPropName(attr)
			props = append(props, propName)
			propMap[attr] = propName
		}
	}

	if len(props) > 0 {
		buf.WriteString(strings.Join(props, ", "))
	}

	buf.WriteString(" }) => {\n")
	buf.WriteString("\treturn (\n")

	buf.WriteString(fmt.Sprintf("\t\t<%s", pattern.TagName))

	for attr, count := range pattern.Attributes {
		if count >= pattern.Count/2 {
			propName := propMap[attr]
			jsxAttr := attr
			if attr == "class" {
				jsxAttr = "className"
			}
			buf.WriteString(fmt.Sprintf(" %s={%s}", jsxAttr, propName))
		}
	}

	buf.WriteString(">\n")
	buf.WriteString("\t\t\t\n")
	buf.WriteString(fmt.Sprintf("\t\t</%s>\n", pattern.TagName))
	buf.WriteString("\t);\n")
	buf.WriteString("};\n\n")
	buf.WriteString("export default " + componentName + ";")

	return buf.String()
}

func GetSuggestionsJSON(htmlInput string) (string, error) {
	suggestions, err := AnalyzeComponents(htmlInput)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(suggestions, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal suggestions to JSON: %w", err)
	}

	return string(jsonData), nil
}
