// Package htmlutil holds the HTML node and attribute primitives shared by the
// upload extractor and the URL scraper.
//
// Both packages previously kept private copies of these helpers, and the copies
// drifted: attribute lookup was case-sensitive in one and not the other, and one
// copy of the script-type check did not tolerate MIME parameters. Keeping a
// single implementation here is what stops that recurring.
package htmlutil

import (
	"strings"

	"golang.org/x/net/html"
)

// GetAttr returns the value of an attribute, matching the name
// case-insensitively as HTML requires. It returns "" when absent.
func GetAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

// SetAttr updates an attribute in place, or appends it when absent.
func SetAttr(n *html.Node, key, val string) {
	for i, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

// HasAttr reports whether an attribute is present, regardless of its value.
func HasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return true
		}
	}
	return false
}

// TextContent concatenates the direct text children of a node.
func TextContent(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}
	return sb.String()
}

// ReplaceNode swaps oldNode for newNode in oldNode's parent. It is a no-op for
// a detached node.
func ReplaceNode(oldNode, newNode *html.Node) {
	if oldNode.Parent == nil {
		return
	}
	oldNode.Parent.InsertBefore(newNode, oldNode)
	oldNode.Parent.RemoveChild(oldNode)
}

// FindElement returns the first element with the given tag name, depth first.
func FindElement(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := FindElement(c, tagName); found != nil {
			return found
		}
	}
	return nil
}

// IsJavaScriptType reports whether a <script type> denotes executable
// JavaScript. An empty type means JavaScript, and MIME parameters such as
// "text/javascript; charset=utf-8" are ignored. Anything else — notably
// application/ld+json — is data and must not be extracted as a script.
func IsJavaScriptType(scriptType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(scriptType))
	if normalized == "" {
		return true
	}
	if idx := strings.Index(normalized, ";"); idx != -1 {
		normalized = strings.TrimSpace(normalized[:idx])
	}
	switch normalized {
	case "text/javascript", "application/javascript", "text/ecmascript",
		"application/ecmascript", "application/x-javascript", "module":
		return true
	default:
		return false
	}
}

// IsGoogleFonts reports whether a URL points at the Google Fonts CSS endpoint,
// which is served per-user-agent and so is not worth localising.
func IsGoogleFonts(rawURL string) bool {
	return strings.Contains(rawURL, "fonts.googleapis.com")
}

// CopyAttributesExcluding returns attrs minus any whose lowercased name is in
// skip. It is used to carry attributes across when an element is replaced.
func CopyAttributesExcluding(attrs []html.Attribute, skip map[string]bool) []html.Attribute {
	var copied []html.Attribute
	for _, attr := range attrs {
		if skip[strings.ToLower(attr.Key)] {
			continue
		}
		copied = append(copied, attr)
	}
	return copied
}
