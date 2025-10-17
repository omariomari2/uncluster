package extractor

import (
	"bytes"
	"fmt"
	"htmlfmt/internal/formatter"
	"strings"

	"golang.org/x/net/html"
)

// ExtractedContent represents the separated HTML, CSS, and JS content
type ExtractedContent struct {
	HTML string
	CSS  string
	JS   string
}

// Extract separates CSS and JS from HTML and returns cleaned HTML with proper linking
func Extract(htmlContent string) (*ExtractedContent, error) {
	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var cssContent strings.Builder
	var jsContent strings.Builder

	// Extract CSS and JS content
	extractStylesAndScripts(doc, &cssContent, &jsContent)

	// Remove style and script tags from the document
	removeStyleAndScriptTags(doc)

	// Add link and script tags to the document
	addLinksToDocument(doc)

	// Convert the modified document back to HTML
	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to render HTML: %w", err)
	}

	// Format the HTML using the existing formatter
	formattedHTML, err := formatter.Format(buf.String())
	if err != nil {
		return nil, fmt.Errorf("failed to format HTML: %w", err)
	}

	return &ExtractedContent{
		HTML: formattedHTML,
		CSS:  cssContent.String(),
		JS:   jsContent.String(),
	}, nil
}

// extractStylesAndScripts recursively extracts content from style and script tags
func extractStylesAndScripts(n *html.Node, cssContent, jsContent *strings.Builder) {
	if n.Type == html.ElementNode {
		if n.Data == "style" {
			// Extract CSS content
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					cssContent.WriteString(c.Data)
					cssContent.WriteString("\n")
				}
			}
		} else if n.Data == "script" {
			// Only extract inline scripts (no src attribute)
			hasSrc := false
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					hasSrc = true
					break
				}
			}
			if !hasSrc {
				// Extract JS content
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						jsContent.WriteString(c.Data)
						jsContent.WriteString("\n")
					}
				}
			}
		}
	}

	// Recursively process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractStylesAndScripts(c, cssContent, jsContent)
	}
}

// removeStyleAndScriptTags removes style and script tags from the document
func removeStyleAndScriptTags(n *html.Node) {
	if n.Type == html.ElementNode && (n.Data == "style" || n.Data == "script") {
		// Check if it's an inline script (no src attribute)
		if n.Data == "script" {
			hasSrc := false
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					hasSrc = true
					break
				}
			}
			if hasSrc {
				// Keep external scripts, don't remove
				return
			}
		}

		// Remove the node
		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	// Recursively process child nodes
	var toRemove []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "style" || c.Data == "script") {
			// Check if it's an inline script (no src attribute)
			if c.Data == "script" {
				hasSrc := false
				for _, attr := range c.Attr {
					if attr.Key == "src" {
						hasSrc = true
						break
					}
				}
				if hasSrc {
					// Keep external scripts, don't remove
					continue
				}
			}
			toRemove = append(toRemove, c)
		} else {
			removeStyleAndScriptTags(c)
		}
	}

	// Remove collected nodes
	for _, node := range toRemove {
		n.RemoveChild(node)
	}
}

// addLinksToDocument adds link and script tags to the document
func addLinksToDocument(doc *html.Node) {
	// Find or create head element
	head := findOrCreateHead(doc)

	// Find or create body element
	body := findOrCreateBody(doc)

	// Add CSS link to head
	addCSSToHead(head)

	// Add JS script to body
	addJSToBody(body)
}

// findOrCreateHead finds the head element or creates one
func findOrCreateHead(doc *html.Node) *html.Node {
	// Look for existing head
	head := findElement(doc, "head")
	if head != nil {
		return head
	}

	// Find html element
	htmlNode := findElement(doc, "html")
	if htmlNode == nil {
		// If no html element, create one
		htmlNode = &html.Node{
			Type: html.ElementNode,
			Data: "html",
		}
		doc.AppendChild(htmlNode)
	}

	// Create head element
	head = &html.Node{
		Type: html.ElementNode,
		Data: "head",
	}
	htmlNode.AppendChild(head)

	return head
}

// findOrCreateBody finds the body element or creates one
func findOrCreateBody(doc *html.Node) *html.Node {
	// Look for existing body
	body := findElement(doc, "body")
	if body != nil {
		return body
	}

	// Find html element
	htmlNode := findElement(doc, "html")
	if htmlNode == nil {
		// If no html element, create one
		htmlNode = &html.Node{
			Type: html.ElementNode,
			Data: "html",
		}
		doc.AppendChild(htmlNode)
	}

	// Create body element
	body = &html.Node{
		Type: html.ElementNode,
		Data: "body",
	}
	htmlNode.AppendChild(body)

	return body
}

// findElement recursively finds an element by tag name
func findElement(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findElement(c, tagName); result != nil {
			return result
		}
	}

	return nil
}

// addCSSToHead adds a link tag for CSS to the head
func addCSSToHead(head *html.Node) {
	link := &html.Node{
		Type: html.ElementNode,
		Data: "link",
		Attr: []html.Attribute{
			{Key: "rel", Val: "stylesheet"},
			{Key: "href", Val: "style.css"},
		},
	}
	head.AppendChild(link)
}

// addJSToBody adds a script tag for JS to the body
func addJSToBody(body *html.Node) {
	script := &html.Node{
		Type: html.ElementNode,
		Data: "script",
		Attr: []html.Attribute{
			{Key: "src", Val: "script.js"},
		},
	}
	body.AppendChild(script)
}
