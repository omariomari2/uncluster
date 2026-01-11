package formatter

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

func Format(htmlInput string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var buf bytes.Buffer
	err = formatNode(&buf, doc, 0)
	if err != nil {
		return "", fmt.Errorf("failed to format HTML: %w", err)
	}

	return buf.String(), nil
}

func formatNode(buf *bytes.Buffer, n *html.Node, depth int) error {
	switch n.Type {
	case html.DocumentNode:
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

func hasOnlyTextChildren(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			return false
		}
	}
	return true
}
