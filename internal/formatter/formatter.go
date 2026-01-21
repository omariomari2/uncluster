package formatter

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"golang.org/x/net/html"
	"strings"
)

func Format(htmlInput string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var buf bytes.Buffer
	err = formatNode(&buf, doc, 0, false)
	if err != nil {
		return "", fmt.Errorf("failed to format HTML: %w", err)
	}

	return buf.String(), nil
}

func formatNode(buf *bytes.Buffer, n *html.Node, depth int, inline bool) error {
	switch n.Type {
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err := formatNode(buf, c, depth, inline); err != nil {
				return err
			}
		}
	case html.ElementNode:
		// Handle self-closing/void elements
		if isVoidElement(n.Data) {
			writeIndent(buf, depth, inline)
			writeOpenTag(buf, n)
			buf.WriteString(" />")
			if !inline {
				buf.WriteString("\n")
			}
		} else {
			writeIndent(buf, depth, inline)
			writeOpenTag(buf, n)
			buf.WriteString(">")

			if isRawTextElement(n.Data) {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if err := formatNode(buf, c, 0, true); err != nil {
						return err
					}
				}
			} else if shouldInlineChildren(n) {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if err := formatNode(buf, c, 0, true); err != nil {
						return err
					}
				}
			} else if hasChildren(n) {
				buf.WriteString("\n")
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if err := formatNode(buf, c, depth+1, false); err != nil {
						return err
					}
				}
				buf.WriteString(strings.Repeat("\t", depth))
			}

			buf.WriteString("</")
			buf.WriteString(n.Data)
			buf.WriteString(">")
			if !inline {
				buf.WriteString("\n")
			}
		}

	case html.TextNode:
		if n.Parent != nil && isRawTextElement(n.Parent.Data) {
			buf.WriteString(n.Data)
		} else {
			buf.WriteString(stdhtml.EscapeString(n.Data))
		}

	case html.CommentNode:
		if !inline {
			buf.WriteString(strings.Repeat("\t", depth))
		}
		buf.WriteString("<!--")
		buf.WriteString(n.Data)
		buf.WriteString("-->")
		if !inline {
			buf.WriteString("\n")
		}

	case html.DoctypeNode:
		buf.WriteString("<!DOCTYPE ")
		buf.WriteString(n.Data)
		buf.WriteString(">")
		if !inline {
			buf.WriteString("\n")
		}
	}

	return nil
}

func writeIndent(buf *bytes.Buffer, depth int, inline bool) {
	if inline {
		return
	}
	buf.WriteString(strings.Repeat("\t", depth))
}

func writeOpenTag(buf *bytes.Buffer, n *html.Node) {
	buf.WriteString("<")
	buf.WriteString(n.Data)

	for _, attr := range n.Attr {
		buf.WriteString(" ")
		buf.WriteString(attr.Key)
		buf.WriteString(`="`)
		buf.WriteString(escapeAttributeValue(attr.Val))
		buf.WriteString(`"`)
	}
}

func escapeAttributeValue(value string) string {
	return stdhtml.EscapeString(value)
}

func shouldInlineChildren(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			return true
		case html.CommentNode:
			return true
		case html.ElementNode:
			if !isBlockElement(c.Data) {
				return true
			}
		}
	}
	return false
}

func isRawTextElement(tagName string) bool {
	rawTextElements := map[string]bool{
		"script": true,
		"style":  true,
		"pre":    true,
		"textarea": true,
	}
	return rawTextElements[strings.ToLower(tagName)]
}

func isVoidElement(tagName string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	return voidElements[strings.ToLower(tagName)]
}

func isBlockElement(tagName string) bool {
	blockElements := map[string]bool{
		"address": true,
		"article": true,
		"aside": true,
		"blockquote": true,
		"body": true,
		"canvas": true,
		"dd": true,
		"div": true,
		"dl": true,
		"dt": true,
		"fieldset": true,
		"figcaption": true,
		"figure": true,
		"footer": true,
		"form": true,
		"h1": true,
		"h2": true,
		"h3": true,
		"h4": true,
		"h5": true,
		"h6": true,
		"head": true,
		"header": true,
		"hr": true,
		"html": true,
		"li": true,
		"main": true,
		"nav": true,
		"noscript": true,
		"ol": true,
		"p": true,
		"section": true,
		"table": true,
		"tbody": true,
		"td": true,
		"tfoot": true,
		"th": true,
		"thead": true,
		"tr": true,
		"ul": true,
	}
	return blockElements[strings.ToLower(tagName)]
}

// hasChildren checks if a node has child nodes
func hasChildren(n *html.Node) bool {
	return n.FirstChild != nil
}
