package analyzer

import (
	"encoding/json"
	"fmt"
	"htmlfmt/internal/ai"
	"golang.org/x/net/html"
	"log"
	"strings"
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

// AIClient is an interface for AI analysis (allows dependency injection for testing)
type AIClient interface {
	AnalyzeHTMLForComponents(htmlContent string, elementInfo string) (*ai.ComponentAnalysisResult, error)
	IsEnabled() bool
}

var globalAIClient AIClient

func SetAIClient(client AIClient) {
	globalAIClient = client
}

func AnalyzeComponents(htmlInput string) ([]ComponentSuggestion, error) {
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Collect all elements and their patterns
	elementPatterns := make(map[string]*ElementPattern)
	collectPatterns(doc, elementPatterns)

	// Generate initial suggestions based on patterns
	suggestions := generateSuggestions(elementPatterns)

	// If AI is enabled, enhance and filter suggestions
	if globalAIClient != nil && globalAIClient.IsEnabled() {
		log.Printf("🤖 Using AI to enhance component analysis...")
		enhancedSuggestions, err := enhanceWithAI(htmlInput, suggestions, elementPatterns)
		if err != nil {
			log.Printf("⚠️ AI analysis failed, using pattern-based suggestions: %v", err)
			return suggestions, nil
		}
		return enhancedSuggestions, nil
	}

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
	
	// Add id if present
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

func generateSuggestions(patterns map[string]*ElementPattern) []ComponentSuggestion {
	var suggestions []ComponentSuggestion
	
	for patternKey, pattern := range patterns {
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
	
	buf.WriteString(fmt.Sprintf("const %s = ({ ", generateComponentName(pattern.TagName, generatePatternKey(example))))
	
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
	buf.WriteString("\t\t\t\n")
	buf.WriteString(fmt.Sprintf("\t\t</%s>\n", pattern.TagName))
	buf.WriteString("\t);\n")
	buf.WriteString("};\n\n")
	buf.WriteString("export default " + generateComponentName(pattern.TagName, generatePatternKey(example)) + ";")
	
	return buf.String()
}

func enhanceWithAI(htmlInput string, suggestions []ComponentSuggestion, patterns map[string]*ElementPattern) ([]ComponentSuggestion, error) {
	if globalAIClient == nil || !globalAIClient.IsEnabled() {
		return suggestions, nil
	}

	var enhancedSuggestions []ComponentSuggestion
	analyzedCount := 0
	skippedCount := 0

	// Analyze each suggestion with AI
	for _, suggestion := range suggestions {
		// Find the pattern for this suggestion
		var pattern *ElementPattern
		for _, p := range patterns {
			if p.TagName == suggestion.TagName && p.Count == suggestion.Count {
				pattern = p
				break
			}
		}

		if pattern == nil || len(pattern.Examples) == 0 {
			enhancedSuggestions = append(enhancedSuggestions, suggestion)
			continue
		}

		// Get example HTML for this pattern
		exampleHTML := nodeToHTML(pattern.Examples[0])
		elementInfo := buildElementInfo(pattern, suggestion)

		// Ask AI if this should be a component
		aiResult, err := globalAIClient.AnalyzeHTMLForComponents(exampleHTML, elementInfo)
		if err != nil {
			log.Printf("⚠️ AI analysis failed for %s: %v", suggestion.Name, err)
			// Keep the suggestion if AI fails
			enhancedSuggestions = append(enhancedSuggestions, suggestion)
			continue
		}

		analyzedCount++

		if !aiResult.ShouldBeComponent {
			log.Printf("🚫 AI determined '%s' should NOT be a component: %s", suggestion.Name, aiResult.Reason)
			skippedCount++
			continue
		}

		// Enhance the suggestion with AI insights
		if aiResult.ComponentName != "" {
			suggestion.Name = aiResult.ComponentName
		}

		if aiResult.Reason != "" {
			suggestion.Description = fmt.Sprintf("%s (AI: %s)", suggestion.Description, aiResult.Reason)
		}

		if len(aiResult.Props) > 0 {
			suggestion.Attributes = make(map[string]string)
			for _, prop := range aiResult.Props {
				suggestion.Attributes[prop] = "{string}"
			}
		}

		// Regenerate JSX code with updated information
		if pattern != nil {
			suggestion.JSXCode = generateJSXCodeWithName(pattern, suggestion.Name)
		}

		enhancedSuggestions = append(enhancedSuggestions, suggestion)
		log.Printf("✅ AI approved component '%s' (confidence: %s)", suggestion.Name, aiResult.Confidence)
	}

	log.Printf("📊 AI Analysis Summary: %d analyzed, %d skipped, %d approved", analyzedCount, skippedCount, len(enhancedSuggestions))

	return enhancedSuggestions, nil
}

func buildElementInfo(pattern *ElementPattern, suggestion ComponentSuggestion) string {
	var info strings.Builder
	info.WriteString(fmt.Sprintf("Tag: %s\n", pattern.TagName))
	info.WriteString(fmt.Sprintf("Count: %d\n", pattern.Count))
	
	if len(pattern.Attributes) > 0 {
		info.WriteString("Attributes: ")
		attrs := make([]string, 0, len(pattern.Attributes))
		for attr := range pattern.Attributes {
			attrs = append(attrs, attr)
		}
		info.WriteString(strings.Join(attrs, ", "))
		info.WriteString("\n")
	}

	if len(pattern.Children) > 0 {
		info.WriteString("Child elements: ")
		children := make([]string, 0, len(pattern.Children))
		for child := range pattern.Children {
			children = append(children, child)
		}
		info.WriteString(strings.Join(children, ", "))
		info.WriteString("\n")
	}

	return info.String()
}

func nodeToHTML(n *html.Node) string {
	var buf strings.Builder
	renderNode(&buf, n)
	return buf.String()
}

// renderNode renders a node to HTML string
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

// generateJSXCodeWithName generates JSX code with a specific component name
func generateJSXCodeWithName(pattern *ElementPattern, componentName string) string {
	if len(pattern.Examples) == 0 {
		return ""
	}

	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("const %s = ({ ", componentName))

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
