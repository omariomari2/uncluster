package htmlutil

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// InlineResource is a <style> or <script> block lifted out of a document into
// its own file.
type InlineResource struct {
	Path    string
	Content string
}

// InlineCollector walks a document, moves each inline <style> and <script> body
// into a numbered file, and replaces the original element with a reference to
// it. The replacement keeps every other attribute on the original element, so
// things like media, nonce and data-* survive the move.
//
// The zero value is ready to use.
type InlineCollector struct {
	// CSS and JS accumulate every extracted block, in document order.
	CSS strings.Builder
	JS  strings.Builder

	// InlineCSS and InlineJS record the individual files that were written.
	InlineCSS []InlineResource
	InlineJS  []InlineResource

	cssIndex int
	jsIndex  int
}

// Walk extracts inline resources from n and its descendants.
func (c *InlineCollector) Walk(n *html.Node) {
	if n.Type == html.ElementNode {
		switch {
		case n.Data == "style":
			if c.extractStyle(n) {
				return
			}
		case n.Data == "script" && !HasAttr(n, "src") && IsJavaScriptType(GetAttr(n, "type")):
			if c.extractScript(n) {
				return
			}
		}
	}

	// The child is replaced during extraction, so capture the sibling first.
	for child := n.FirstChild; child != nil; {
		next := child.NextSibling
		c.Walk(child)
		child = next
	}
}

// extractStyle moves a <style> body to a file, reporting whether it did.
func (c *InlineCollector) extractStyle(n *html.Node) bool {
	content := TextContent(n)
	if strings.TrimSpace(content) == "" {
		return false
	}

	c.cssIndex++
	path := fmt.Sprintf("inline/style-%d.css", c.cssIndex)
	c.InlineCSS = append(c.InlineCSS, InlineResource{Path: path, Content: content})
	writeBlock(&c.CSS, content)

	attrs := append([]html.Attribute{
		{Key: "rel", Val: "stylesheet"},
		{Key: "href", Val: path},
	}, CopyAttributesExcluding(n.Attr, map[string]bool{"rel": true, "href": true})...)

	ReplaceNode(n, &html.Node{Type: html.ElementNode, Data: "link", Attr: attrs})
	return true
}

// extractScript moves a <script> body to a file, reporting whether it did.
func (c *InlineCollector) extractScript(n *html.Node) bool {
	content := TextContent(n)
	if strings.TrimSpace(content) == "" {
		return false
	}

	c.jsIndex++
	path := fmt.Sprintf("inline/script-%d.js", c.jsIndex)
	c.InlineJS = append(c.InlineJS, InlineResource{Path: path, Content: content})
	writeBlock(&c.JS, content)

	attrs := append([]html.Attribute{
		{Key: "src", Val: path},
	}, CopyAttributesExcluding(n.Attr, map[string]bool{"src": true})...)

	ReplaceNode(n, &html.Node{Type: html.ElementNode, Data: "script", Attr: attrs})
	return true
}

func writeBlock(sb *strings.Builder, content string) {
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
}
