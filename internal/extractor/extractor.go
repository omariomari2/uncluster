package extractor

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/omariomari2/uncluster/internal/fetcher"
	"github.com/omariomari2/uncluster/internal/formatter"
	"github.com/omariomari2/uncluster/internal/htmlutil"

	"golang.org/x/net/html"
)

type ExtractedContent struct {
	HTML        string
	CSS         string
	JS          string
	InlineCSS   []InlineResource
	InlineJS    []InlineResource
	ExternalCSS []fetcher.FetchedResource
	ExternalJS  []fetcher.FetchedResource
	LocalAssets []LocalAsset
	Capture     *CaptureManifest
}

// InlineResource is an alias so callers keep using extractor.InlineResource
// while the extraction itself lives in htmlutil, shared with the scraper.
type InlineResource = htmlutil.InlineResource

// LocalAsset holds a binary file (image, font, SVG, etc.) that was either
// bundled in an uploaded ZIP or downloaded by the URL scraper.
type LocalAsset struct {
	Path    string // relative path as it should appear in the export, e.g. "assets/logo.png"
	Content []byte // raw binary content
	MIME    string // e.g. "image/png", "font/woff2"
}

func Extract(htmlContent string) (*ExtractedContent, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var inline htmlutil.InlineCollector
	inline.Walk(doc)

	cssURLs, jsURLs := findExternalResourceURLs(doc)

	var externalCSS []fetcher.FetchedResource
	var externalJS []fetcher.FetchedResource

	if len(cssURLs) > 0 {
		externalCSS = fetcher.FetchExternalResources(cssURLs, "css")
	}
	if len(jsURLs) > 0 {
		externalJS = fetcher.FetchExternalResources(jsURLs, "js")
	}

	capture := NewCaptureManifest("")
	addFetchedCaptureAssets(capture, externalCSS, "external/css/", "css")
	addFetchedCaptureAssets(capture, externalJS, "external/js/", "js")
	addRetainedCaptureAssets(doc, capture)

	rewriteLinks(doc, externalCSS, externalJS)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, fmt.Errorf("failed to render HTML: %w", err)
	}

	formattedHTML, err := formatter.Format(buf.String())
	if err != nil {
		return nil, fmt.Errorf("failed to format HTML: %w", err)
	}

	return &ExtractedContent{
		HTML:        formattedHTML,
		CSS:         inline.CSS.String(),
		JS:          inline.JS.String(),
		InlineCSS:   inline.InlineCSS,
		InlineJS:    inline.InlineJS,
		ExternalCSS: externalCSS,
		ExternalJS:  externalJS,
		Capture:     capture,
	}, nil
}

func addFetchedCaptureAssets(capture *CaptureManifest, resources []fetcher.FetchedResource, dir, assetType string) {
	for _, resource := range resources {
		if resource.Error == nil && resource.Content != "" {
			capture.AddAsset(CaptureAsset{
				URL: resource.URL, Path: dir + resource.Filename,
				Type: assetType, Status: CaptureLocalized,
			})
			continue
		}
		message := "empty response body"
		if resource.Error != nil {
			message = resource.Error.Error()
		}
		capture.AddAsset(CaptureAsset{
			URL: resource.URL, Type: assetType, Status: CaptureFailed, Error: message,
		})
	}
}

func addRetainedCaptureAssets(doc *html.Node, capture *CaptureManifest) {
	seen := make(map[string]bool)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var rawURL, assetType string
			switch n.Data {
			case "link":
				href := htmlutil.GetAttr(n, "href")
				if strings.Contains(strings.ToLower(htmlutil.GetAttr(n, "rel")), "stylesheet") && htmlutil.IsGoogleFonts(href) {
					rawURL, assetType = href, "css"
				}
			case "iframe":
				rawURL, assetType = htmlutil.GetAttr(n, "src"), "iframe"
			}
			key := assetType + "\x00" + rawURL
			if assetType != "" && isExternalURL(rawURL) && !seen[key] {
				seen[key] = true
				capture.AddAsset(CaptureAsset{URL: rawURL, Type: assetType, Status: CaptureRetainedExternal})
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
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
			href := htmlutil.GetAttr(n, "href")
			rel := htmlutil.GetAttr(n, "rel")
			if href != "" && rel == "stylesheet" && isExternalURL(href) && !htmlutil.IsGoogleFonts(href) {
				*cssURLs = append(*cssURLs, href)
			}
		} else if n.Data == "script" {
			src := htmlutil.GetAttr(n, "src")
			if src != "" && isExternalURL(src) {
				*jsURLs = append(*jsURLs, src)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findExternalURLs(c, cssURLs, jsURLs)
	}
}

func isExternalURL(urlStr string) bool {
	return strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://")
}

func rewriteLinks(n *html.Node, externalCSS, externalJS []fetcher.FetchedResource) {
	if n.Type == html.ElementNode {
		if n.Data == "link" {
			href := htmlutil.GetAttr(n, "href")
			if href != "" && isExternalURL(href) {
				for _, resource := range externalCSS {
					if resource.URL == href && resource.Error == nil && resource.Content != "" {
						htmlutil.SetAttr(n, "href", "external/css/"+resource.Filename)
						break
					}
				}
			}
		} else if n.Data == "script" {
			src := htmlutil.GetAttr(n, "src")
			if src != "" && isExternalURL(src) {
				for _, resource := range externalJS {
					if resource.URL == src && resource.Error == nil && resource.Content != "" {
						htmlutil.SetAttr(n, "src", "external/js/"+resource.Filename)
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

func (e *ExtractedContent) RewriteForNodeJS() string {
	return e.rewritten(rewriteLinksForNodeJS)
}

func (e *ExtractedContent) RewriteForEJS() string {
	return e.rewritten(rewriteLinksForEJS)
}

// rewritten reparses the extracted HTML, applies a link rewriter, and renders it
// back. It returns the original HTML unchanged if either step fails.
func (e *ExtractedContent) rewritten(rewrite func(*html.Node)) string {
	doc, err := html.Parse(strings.NewReader(e.HTML))
	if err != nil {
		return e.HTML
	}

	rewrite(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return e.HTML
	}
	return buf.String()
}

func rewriteLinksForEJS(n *html.Node) {
	if n.Type == html.ElementNode {
		if n.Data == "link" {
			if href := htmlutil.GetAttr(n, "href"); strings.HasPrefix(href, "inline/") || strings.HasPrefix(href, "external/css/") {
				htmlutil.SetAttr(n, "href", "/"+href)
			}
		} else if n.Data == "script" {
			if src := htmlutil.GetAttr(n, "src"); strings.HasPrefix(src, "inline/") || strings.HasPrefix(src, "external/js/") {
				htmlutil.SetAttr(n, "src", "/"+src)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		rewriteLinksForEJS(c)
	}
}

func rewriteLinksForNodeJS(n *html.Node) {
	if n.Type == html.ElementNode {
		if n.Data == "link" {
			href := htmlutil.GetAttr(n, "href")
			if href == "style.css" {
				htmlutil.SetAttr(n, "href", "/styles/main.css")
			} else if strings.HasPrefix(href, "external/css/") {
				htmlutil.SetAttr(n, "href", "/styles/external/"+strings.TrimPrefix(href, "external/css/"))
			}
		} else if n.Data == "script" {
			src := htmlutil.GetAttr(n, "src")
			if src == "script.js" {
				htmlutil.SetAttr(n, "src", "/scripts/main.js")
			} else if strings.HasPrefix(src, "external/js/") {
				htmlutil.SetAttr(n, "src", "/scripts/external/"+strings.TrimPrefix(src, "external/js/"))
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		rewriteLinksForNodeJS(c)
	}
}
