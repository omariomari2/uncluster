package analyzer

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// ComponentSuggestion represents a suggested component extraction
type ComponentSuggestion struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	TagName     string            `json:"tagName"`
	Attributes  map[string]string `json:"attributes"`
	Children    []string          `json:"children"`
	Count       int               `json:"count"`
	JSXCode     string            `json:"jsxCode"`
}

// AnalyzeComponents analyzes HTML and returns component suggestions
func AnalyzeComponents(htmlInput string) ([]ComponentSuggestion, error) {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Collect all elements and their patterns
	elementPatterns := make(map[string]*ElementPattern)
	collectPatterns(doc, elementPatterns)

	// Generate suggestions based on patterns
	suggestions := generateSuggestions(elementPatterns)

	return suggestions, nil
}

// ElementPattern represents a pattern found in the HTML
type ElementPattern struct {
	TagName    string
	Attributes map[string]int
	Children   map[string]int
	Count      int
	Examples   []*html.Node
}

// collectPatterns recursively collects element patterns from the DOM
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
		
		// Collect attributes
		for _, attr := range n.Attr {
			pattern.Attributes[attr.Key]++
		}
		
		// Collect child elements
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode {
				pattern.Children[c.Data]++
			}
		}
		
		// Keep examples (limit to 3)
		if len(pattern.Examples) < 3 {
			pattern.Examples = append(pattern.Examples, n)
		}
	}
	
	// Recursively process children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectPatterns(c, patterns)
	}
}

// generatePatternKey creates a unique key for an element pattern
func generatePatternKey(n *html.Node) string {
	// Use tag name as base
	key := n.Data
	
	// Add class information if present
	classes := getAttributeValue(n, "class")
	if classes != "" {
		key += "." + strings.ReplaceAll(classes, " ", ".")
	}
	
	// Add id if present
	id := getAttributeValue(n, "id")
	if id != "" {
		key += "#" + id
	}
	
	return key
}

// getAttributeValue gets the value of an attribute from a node
func getAttributeValue(n *html.Node, attrName string) string {
	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// generateSuggestions creates component suggestions from patterns
func generateSuggestions(patterns map[string]*ElementPattern) []ComponentSuggestion {
	var suggestions []ComponentSuggestion
	
	for patternKey, pattern := range patterns {
		// Only suggest components for elements that appear multiple times or have significant structure
		if pattern.Count < 2 && len(pattern.Children) < 2 {
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
		
		// Add common attributes as props
		for attr, count := range pattern.Attributes {
			if count >= pattern.Count/2 { // Attribute appears in at least half of instances
				suggestion.Attributes[attr] = "{string}" // Default to string type
			}
		}
		
		// Add child element types
		for childTag, count := range pattern.Children {
			if count >= pattern.Count/2 {
				suggestion.Children = append(suggestion.Children, childTag)
			}
		}
		
		suggestions = append(suggestions, suggestion)
	}
	
	return suggestions
}

// generateComponentName creates a component name from tag and pattern
func generateComponentName(tagName, patternKey string) string {
	// Convert to PascalCase
	name := strings.Title(tagName)
	
	// Add descriptive suffix based on common patterns
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

// generateDescription creates a description for the component
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

// generateJSXCode creates example JSX code for the component
func generateJSXCode(pattern *ElementPattern) string {
	if len(pattern.Examples) == 0 {
		return ""
	}
	
	example := pattern.Examples[0]
	var buf strings.Builder
	
	// Component definition
	buf.WriteString(fmt.Sprintf("const %s = ({ ", generateComponentName(pattern.TagName, generatePatternKey(example))))
	
	// Add props based on common attributes
	props := []string{}
	for attr, count := range pattern.Attributes {
		if count >= pattern.Count/2 {
			props = append(props, attr+"=\"{string}\"")
		}
	}
	
	if len(props) > 0 {
		buf.WriteString(strings.Join(props, ", "))
	}
	
	buf.WriteString(" }) => {\n")
	buf.WriteString("\treturn (\n")
	
	// Generate JSX element
	buf.WriteString(fmt.Sprintf("\t\t<%s", pattern.TagName))
	
	// Add props
	for attr, count := range pattern.Attributes {
		if count >= pattern.Count/2 {
			buf.WriteString(fmt.Sprintf(" %s={%s}", attr, attr))
		}
	}
	
	buf.WriteString(">\n")
	buf.WriteString("\t\t\t{/* Add your content here */}\n")
	buf.WriteString(fmt.Sprintf("\t\t</%s>\n", pattern.TagName))
	buf.WriteString("\t);\n")
	buf.WriteString("};\n\n")
	buf.WriteString("export default " + generateComponentName(pattern.TagName, generatePatternKey(example)) + ";")
	
	return buf.String()
}

// GetSuggestionsJSON returns component suggestions as JSON
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
