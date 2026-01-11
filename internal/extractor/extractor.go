package extractor

import (
	"bytes"
	"fmt"
	"htmlfmt/internal/fetcher"
	"htmlfmt/internal/formatter"
	"log"
	"strings"

	"golang.org/x/net/html"
)

type ExtractedContent struct {
	HTML        string                    // cleaned HTML with rewritten links
	CSS         string                    // inline CSS from <style> tags
	JS          string                    // inline JS from <script> tags
	ExternalCSS []fetcher.FetchedResource // downloaded external CSS files
	ExternalJS  []fetcher.FetchedResource // downloaded external JS files
}

func Extract(htmlContent string) (*ExtractedContent, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var cssContent strings.Builder
	var jsContent strings.Builder

	extractStylesAndScripts(doc, &cssContent, &jsContent)

	cssURLs, jsURLs := findExternalResourceURLs(doc)

	log.Printf("🔍 Found %d external CSS URLs and %d external JS URLs", len(cssURLs), len(jsURLs))

	var externalCSS []fetcher.FetchedResource
	var externalJS []fetcher.FetchedResource

	if len(cssURLs) > 0 {
		externalCSS = fetcher.FetchExternalResources(cssURLs, "css")
	}
	if len(jsURLs) > 0 {
		externalJS = fetcher.FetchExternalResources(jsURLs, "js")
	}

	rewriteExternalLinks(doc, externalCSS, externalJS)

	removeStyleAndScriptTags(doc)

	addLinksToDocument(doc)

	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to render HTML: %w", err)
	}

	formattedHTML, err := formatter.Format(buf.String())
	if err != nil {
		return nil, fmt.Errorf("failed to format HTML: %w", err)
	}

	return &ExtractedContent{
		HTML:        formattedHTML,
		CSS:         cssContent.String(),
		JS:          jsContent.String(),
		ExternalCSS: externalCSS,
		ExternalJS:  externalJS,
	}, nil
}

func extractStylesAndScripts(n *html.Node, cssContent, jsContent *strings.Builder) {
	if n.Type == html.ElementNode {
		if n.Data == "style" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					cssContent.WriteString(c.Data)
					cssContent.WriteString("\n")
				}
			}
		} else if n.Data == "script" {
			hasSrc := false
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					hasSrc = true
					break
				}
			}
			if !hasSrc {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						jsContent.WriteString(c.Data)
						jsContent.WriteString("\n")
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractStylesAndScripts(c, cssContent, jsContent)
	}
}

func removeStyleAndScriptTags(n *html.Node) {
	if n.Type == html.ElementNode && (n.Data == "style" || n.Data == "script") {
		if n.Data == "script" {
			hasSrc := false
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					hasSrc = true
					break
				}
			}
			if hasSrc {
				return
			}
		}

		if n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
		return
	}

	var toRemove []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "style" || c.Data == "script") {
			if c.Data == "script" {
				hasSrc := false
				for _, attr := range c.Attr {
					if attr.Key == "src" {
						hasSrc = true
						break
					}
				}
				if hasSrc {
					continue
				}
			}
			toRemove = append(toRemove, c)
		} else {
			removeStyleAndScriptTags(c)
		}
	}

	for _, node := range toRemove {
		n.RemoveChild(node)
	}
}

func addLinksToDocument(doc *html.Node) {
	head := findOrCreateHead(doc)

	body := findOrCreateBody(doc)

	addCSSToHead(head)

	addJSToBody(body)
}

func findOrCreateHead(doc *html.Node) *html.Node {
	head := findElement(doc, "head")
	if head != nil {
		return head
	}

	htmlNode := findElement(doc, "html")
	if htmlNode == nil {
		htmlNode = &html.Node{
			Type: html.ElementNode,
			Data: "html",
		}
		doc.AppendChild(htmlNode)
	}

	head = &html.Node{
		Type: html.ElementNode,
		Data: "head",
	}
	htmlNode.AppendChild(head)

	return head
}

func findOrCreateBody(doc *html.Node) *html.Node {
	body := findElement(doc, "body")
	if body != nil {
		return body
	}

	htmlNode := findElement(doc, "html")
	if htmlNode == nil {
		htmlNode = &html.Node{
			Type: html.ElementNode,
			Data: "html",
		}
		doc.AppendChild(htmlNode)
	}

	body = &html.Node{
		Type: html.ElementNode,
		Data: "body",
	}
	htmlNode.AppendChild(body)

	return body
}

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

func findExternalResourceURLs(doc *html.Node) ([]string, []string) {
	var cssURLs []string
	var jsURLs []string

	findExternalURLs(doc, &cssURLs, &jsURLs)
	return cssURLs, jsURLs
}

func findExternalURLs(n *html.Node, cssURLs, jsURLs *[]string) {
	if n.Type == html.ElementNode {
		if n.Data == "link" {
			href := getAttribute(n, "href")
			rel := getAttribute(n, "rel")
			if href != "" && rel == "stylesheet" && isExternalURL(href) {
				*cssURLs = append(*cssURLs, href)
			}
		} else if n.Data == "script" {
			src := getAttribute(n, "src")
			if src != "" && isExternalURL(src) {
				*jsURLs = append(*jsURLs, src)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findExternalURLs(c, cssURLs, jsURLs)
	}
}

func getAttribute(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func isExternalURL(urlStr string) bool {
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}

func rewriteExternalLinks(doc *html.Node, externalCSS, externalJS []fetcher.FetchedResource) {
	rewriteLinks(doc, externalCSS, externalJS)
}

func rewriteLinks(n *html.Node, externalCSS, externalJS []fetcher.FetchedResource) {
	if n.Type == html.ElementNode {
		if n.Data == "link" {
			href := getAttribute(n, "href")
			if href != "" && isExternalURL(href) {
				for _, resource := range externalCSS {
					if resource.URL == href && resource.Error == nil {
						updateAttribute(n, "href", "external/css/"+resource.Filename)
						break
					}
				}
			}
		} else if n.Data == "script" {
			src := getAttribute(n, "src")
			if src != "" && isExternalURL(src) {
				for _, resource := range externalJS {
					if resource.URL == src && resource.Error == nil {
						updateAttribute(n, "src", "external/js/"+resource.Filename)
						break
					}
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		rewriteLinks(c, externalCSS, externalJS)
	}
}

func updateAttribute(n *html.Node, key, value string) {
	for i, attr := range n.Attr {
		if attr.Key == key {
			n.Attr[i].Val = value
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: value})
}

func (e *ExtractedContent) RewriteForNodeJS() string {
	doc, err := html.Parse(strings.NewReader(e.HTML))
	if err != nil {
		return e.HTML
	}

	rewriteLinksForNodeJS(doc)

	var buf bytes.Buffer
	err = html.Render(&buf, doc)
	if err != nil {
		return e.HTML
	}

	return buf.String()
}

func rewriteLinksForNodeJS(n *html.Node) {
	if n.Type == html.ElementNode {
		if n.Data == "link" {
			href := getAttribute(n, "href")
			if href != "" {
				if href == "style.css" {
					updateAttribute(n, "href", "/styles/main.css")
				} else if strings.HasPrefix(href, "external/css/") {
					filename := strings.TrimPrefix(href, "external/css/")
					updateAttribute(n, "href", "/styles/external/"+filename)
				}
			}
		} else if n.Data == "script" {
			src := getAttribute(n, "src")
			if src != "" {
				if src == "script.js" {
					updateAttribute(n, "src", "/scripts/main.js")
				} else if strings.HasPrefix(src, "external/js/") {
					filename := strings.TrimPrefix(src, "external/js/")
					updateAttribute(n, "src", "/scripts/external/"+filename)
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		rewriteLinksForNodeJS(c)
	}
}
