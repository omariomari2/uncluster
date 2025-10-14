package converter

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// ConvertToJSX takes an HTML string and converts it to JSX with proper formatting
func ConvertToJSX(htmlInput string) (string, error) {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Convert to JSX
	var buf bytes.Buffer
	err = convertNodeToJSX(&buf, doc, 0)
	if err != nil {
		return "", fmt.Errorf("failed to convert to JSX: %w", err)
	}

	return buf.String(), nil
}

// convertNodeToJSX recursively converts an HTML node to JSX
func convertNodeToJSX(buf *bytes.Buffer, n *html.Node, depth int) error {
	switch n.Type {
	case html.DocumentNode:
		// Process all children of document
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := convertNodeToJSX(buf, c, depth); err != nil {
				return err
			}
		}

	case html.ElementNode:
		// Handle self-closing/void elements
		if isVoidElement(n.Data) {
			buf.WriteString(strings.Repeat("\t", depth))
			buf.WriteString("<")
			buf.WriteString(convertTagName(n.Data))
			
			// Add attributes
			for _, attr := range n.Attr {
				buf.WriteString(" ")
				buf.WriteString(convertAttributeName(attr.Key))
				buf.WriteString("=")
				buf.WriteString(convertAttributeValue(attr.Val))
			}
			buf.WriteString(" />\n")
		} else {
			// Opening tag
			buf.WriteString(strings.Repeat("\t", depth))
			buf.WriteString("<")
			buf.WriteString(convertTagName(n.Data))
			
			// Add attributes
			for _, attr := range n.Attr {
				buf.WriteString(" ")
				buf.WriteString(convertAttributeName(attr.Key))
				buf.WriteString("=")
				buf.WriteString(convertAttributeValue(attr.Val))
			}
			buf.WriteString(">")

			// Check if element has only text content
			hasOnlyText := hasOnlyTextChildren(n)
			
			if !hasOnlyText && hasChildren(n) {
				buf.WriteString("\n")
			}

			// Process children
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := convertNodeToJSX(buf, c, depth+1); err != nil {
					return err
				}
			}

			// Closing tag
			if !hasOnlyText && hasChildren(n) {
				buf.WriteString(strings.Repeat("\t", depth))
			}
			buf.WriteString("</")
			buf.WriteString(convertTagName(n.Data))
			buf.WriteString(">\n")
		}

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			// Only add indentation if this text node is not the only child
			parent := n.Parent
			if parent != nil && !hasOnlyTextChildren(parent) {
				buf.WriteString(strings.Repeat("\t", depth))
			}
			buf.WriteString(text)
			if parent != nil && !hasOnlyTextChildren(parent) {
				buf.WriteString("\n")
			}
		}

	case html.CommentNode:
		buf.WriteString(strings.Repeat("\t", depth))
		buf.WriteString("{/* ")
		buf.WriteString(n.Data)
		buf.WriteString(" */}\n")
	}

	return nil
}

// convertTagName converts HTML tag names to JSX format
func convertTagName(tagName string) string {
	// Convert to lowercase for consistency
	tagName = strings.ToLower(tagName)
	
	// Handle special cases
	switch tagName {
	case "class":
		return "className"
	default:
		return tagName
	}
}

// convertAttributeName converts HTML attribute names to JSX format
func convertAttributeName(attrName string) string {
	// Handle special JSX attribute names
	switch strings.ToLower(attrName) {
	case "class":
		return "className"
	case "for":
		return "htmlFor"
	case "tabindex":
		return "tabIndex"
	case "readonly":
		return "readOnly"
	case "maxlength":
		return "maxLength"
	case "minlength":
		return "minLength"
	case "autocomplete":
		return "autoComplete"
	case "autofocus":
		return "autoFocus"
	case "autoplay":
		return "autoPlay"
	case "autosave":
		return "autoSave"
	case "cellpadding":
		return "cellPadding"
	case "cellspacing":
		return "cellSpacing"
	case "colspan":
		return "colSpan"
	case "datetime":
		return "dateTime"
	case "enctype":
		return "encType"
	case "formaction":
		return "formAction"
	case "formenctype":
		return "formEncType"
	case "formmethod":
		return "formMethod"
	case "formnovalidate":
		return "formNoValidate"
	case "formtarget":
		return "formTarget"
	case "frameborder":
		return "frameBorder"
	case "hreflang":
		return "hrefLang"
	case "http-equiv":
		return "httpEquiv"
	case "inputmode":
		return "inputMode"
	case "ismap":
		return "isMap"
	case "itemid":
		return "itemID"
	case "itemprop":
		return "itemProp"
	case "itemref":
		return "itemRef"
	case "itemscope":
		return "itemScope"
	case "itemtype":
		return "itemType"
	case "marginheight":
		return "marginHeight"
	case "marginwidth":
		return "marginWidth"
	case "mediagroup":
		return "mediaGroup"
	case "novalidate":
		return "noValidate"
	case "radiogroup":
		return "radioGroup"
	case "rowspan":
		return "rowSpan"
	case "spellcheck":
		return "spellCheck"
	case "srcdoc":
		return "srcDoc"
	case "srclang":
		return "srcLang"
	case "srcset":
		return "srcSet"
	case "usemap":
		return "useMap"
	default:
		// Convert kebab-case to camelCase
		return toCamelCase(attrName)
	}
}

// convertAttributeValue converts HTML attribute values to JSX format
func convertAttributeValue(value string) string {
	// Handle style attribute specially
	if strings.Contains(value, ":") && strings.Contains(value, ";") {
		return convertStyleToJSX(value)
	}
	
	// Handle boolean attributes
	booleanAttrs := map[string]bool{
		"checked": true, "disabled": true, "readonly": true, "required": true,
		"selected": true, "defer": true, "reversed": true, "autofocus": true,
		"autoplay": true, "controls": true, "loop": true, "muted": true,
		"default": true, "novalidate": true, "formnovalidate": true,
	}
	
	if booleanAttrs[strings.ToLower(value)] {
		return value
	}
	
	// Wrap in quotes for string values
	return `"` + value + `"`
}

// convertStyleToJSX converts inline CSS styles to JSX style objects
func convertStyleToJSX(style string) string {
	// Remove extra whitespace
	style = strings.TrimSpace(style)
	
	// Split by semicolon
	stylePairs := strings.Split(style, ";")
	var jsxStyles []string
	
	for _, pair := range stylePairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		// Split by colon
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		
		prop := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Convert CSS property to camelCase
		jsxProp := toCamelCase(prop)
		
		// Handle numeric values (remove quotes if it's a number)
		if isNumeric(value) {
			jsxStyles = append(jsxStyles, jsxProp+": "+value)
		} else {
			jsxStyles = append(jsxStyles, jsxProp+`: "`+value+`"`)
		}
	}
	
	return `{{` + strings.Join(jsxStyles, ", ") + `}}`
}

// toCamelCase converts kebab-case strings to camelCase
func toCamelCase(s string) string {
	// Handle empty string
	if s == "" {
		return s
	}
	
	// Split by hyphens
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		return s
	}
	
	// First part stays lowercase, rest get capitalized
	result := parts[0]
	for _, part := range parts[1:] {
		if part != "" {
			result += strings.Title(part)
		}
	}
	
	return result
}

// isNumeric checks if a string represents a numeric value
func isNumeric(s string) bool {
	// Remove common CSS units and check if the rest is numeric
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	
	// Check for CSS units
	units := []string{"px", "em", "rem", "%", "vh", "vw", "pt", "pc", "in", "cm", "mm"}
	for _, unit := range units {
		if strings.HasSuffix(s, unit) {
			// Remove unit and check if the rest is numeric
			num := strings.TrimSuffix(s, unit)
			if num == "" {
				return false
			}
			// Simple numeric check (could be improved with proper regex)
			for _, r := range num {
				if r < '0' || r > '9' {
					if r != '.' && r != '-' {
						return false
					}
				}
			}
			return true
		}
	}
	
	// Check if it's a pure number
	for _, r := range s {
		if r < '0' || r > '9' {
			if r != '.' && r != '-' {
				return false
			}
		}
	}
	return true
}

// isVoidElement checks if an element is void (self-closing)
func isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return voidElements[strings.ToLower(tagName)]
}

// hasChildren checks if a node has child nodes
func hasChildren(n *html.Node) bool {
	return n.FirstChild != nil
}

// hasOnlyTextChildren checks if a node has only text children (no element children)
func hasOnlyTextChildren(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			return false
		}
	}
	return true
}
