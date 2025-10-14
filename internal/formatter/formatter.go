package formatter

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// Format takes a clustered HTML string and returns properly formatted HTML with tab indentation
func Format(htmlInput string) (string, error) {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Format the parsed HTML
	var buf bytes.Buffer
	err = formatNode(&buf, doc, 0)
	if err != nil {
		return "", fmt.Errorf("failed to format HTML: %w", err)
	}

	return buf.String(), nil
}

// formatNode recursively formats an HTML node with proper indentation
func formatNode(buf *bytes.Buffer, n *html.Node, depth int) error {
	switch n.Type {
	case html.DocumentNode:
		// Process all children of document
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := formatNode(buf, c, depth); err != nil {
				return err
			}
		}
	case html.ElementNode:
		// Handle self-closing/void elements
		if isVoidElement(n.Data) {
			buf.WriteString(strings.Repeat("\t", depth))
			buf.WriteString("<")
			buf.WriteString(n.Data)
			
			// Add attributes
			for _, attr := range n.Attr {
				buf.WriteString(" ")
				buf.WriteString(attr.Key)
				if attr.Val != "" {
					buf.WriteString(`="`)
					buf.WriteString(attr.Val)
					buf.WriteString(`"`)
				}
			}
			buf.WriteString(" />\n")
		} else {
			// Opening tag
			buf.WriteString(strings.Repeat("\t", depth))
			buf.WriteString("<")
			buf.WriteString(n.Data)
			
			// Add attributes
			for _, attr := range n.Attr {
				buf.WriteString(" ")
				buf.WriteString(attr.Key)
				if attr.Val != "" {
					buf.WriteString(`="`)
					buf.WriteString(attr.Val)
					buf.WriteString(`"`)
				}
			}
			buf.WriteString(">")

			// Check if element has only text content
			hasOnlyText := hasOnlyTextChildren(n)
			
			if !hasOnlyText && hasChildren(n) {
				buf.WriteString("\n")
			}

			// Process children
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err := formatNode(buf, c, depth+1); err != nil {
					return err
				}
			}

			// Closing tag
			if !hasOnlyText && hasChildren(n) {
				buf.WriteString(strings.Repeat("\t", depth))
			}
			buf.WriteString("</")
			buf.WriteString(n.Data)
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
		buf.WriteString("<!--")
		buf.WriteString(n.Data)
		buf.WriteString("-->\n")

	case html.DoctypeNode:
		buf.WriteString("<!DOCTYPE ")
		buf.WriteString(n.Data)
		buf.WriteString(">\n")
	}

	return nil
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
