package scraper

import (
	"bytes"
	"fmt"
	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"
	"github.com/omariomari2/uncluster/internal/formatter"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var cssURLRegex = regexp.MustCompile(`url\(\s*['"]?([^'")\s]+)['"]?\s*\)`)

// ScrapeURL fetches a webpage and all its referenced assets (CSS, JS, images,
// fonts, SVGs) and returns an ExtractedContent ready for the export pipeline.
func ScrapeURL(rawURL string) (*extractor.ExtractedContent, error) {
	base, err := url.Parse(rawURL)
	if err != nil || (base.Scheme != "http" && base.Scheme != "https") {
		return nil, fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	pageHTML, err := fetchPage(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(pageHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	cssURLs, jsURLs, binaryURLs := findAllAssetURLs(doc, base)

	// Build a URL→localPath map for path rewriting
	urlToLocal := make(map[string]string)

	// Fetch CSS and JS as text resources
	var externalCSS []fetcher.FetchedResource
	var externalJS []fetcher.FetchedResource

	if len(cssURLs) > 0 {
		externalCSS = fetcher.FetchExternalResources(cssURLs, "css")
		for _, r := range externalCSS {
			if r.Error == nil {
				urlToLocal[r.URL] = "external/css/" + r.Filename
				// Also scan CSS content for url() references (fonts, bg images)
				extraBinary := extractCSSURLs(r.Content, r.URL)
				binaryURLs = append(binaryURLs, extraBinary...)
			}
		}
	}

	if len(jsURLs) > 0 {
		externalJS = fetcher.FetchExternalResources(jsURLs, "js")
		for _, r := range externalJS {
			if r.Error == nil {
				urlToLocal[r.URL] = "external/js/" + r.Filename
			}
		}
	}

	// Deduplicate binary URLs
	binaryURLs = deduplicateStrings(binaryURLs)

	// Fetch binary assets
	var localAssets []extractor.LocalAsset
	binaryUsedNames := make(map[string]int)
	for _, bURL := range binaryURLs {
		data, mime, err := fetcher.FetchRaw(bURL)
		if err != nil {
			log.Printf("scraper: skipping binary asset %s: %v", bURL, err)
			continue
		}
		filename := binaryFilename(bURL, mime, binaryUsedNames)
		localPath := "assets/" + filename
		urlToLocal[bURL] = localPath
		localAssets = append(localAssets, extractor.LocalAsset{
			Path:    localPath,
			Content: data,
			MIME:    mime,
		})
	}

	// Rewrite src/href in the document to local relative paths
	rewriteHTMLPaths(doc, urlToLocal, base)

	// Extract inline <style> and <script> tags (reuse extractor logic)
	var cssContent strings.Builder
	var jsContent strings.Builder
	var inlineCSS []extractor.InlineResource
	var inlineJS []extractor.InlineResource
	cssIndex := 0
	jsIndex := 0
	extractInlineResources(doc, &cssContent, &jsContent, &inlineCSS, &inlineJS, &cssIndex, &jsIndex)

	// Render the final HTML
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, fmt.Errorf("failed to render HTML: %w", err)
	}

	formattedHTML, err := formatter.Format(buf.String())
	if err != nil {
		formattedHTML = buf.String()
	}

	return &extractor.ExtractedContent{
		HTML:        formattedHTML,
		CSS:         cssContent.String(),
		JS:          jsContent.String(),
		InlineCSS:   inlineCSS,
		InlineJS:    inlineJS,
		ExternalCSS: externalCSS,
		ExternalJS:  externalJS,
		LocalAssets: localAssets,
	}, nil
}

// fetchPage downloads the HTML content of a URL with a browser User-Agent.
func fetchPage(rawURL string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// findAllAssetURLs walks the HTML tree and collects absolute URLs for
// CSS, JS, and binary assets (images, fonts, SVGs).
func findAllAssetURLs(doc *html.Node, base *url.URL) (cssURLs, jsURLs, binaryURLs []string) {
	cssSet := make(map[string]bool)
	jsSet := make(map[string]bool)
	binarySet := make(map[string]bool)

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "link":
				rel := strings.ToLower(getAttr(n, "rel"))
				href := getAttr(n, "href")
				if href == "" {
					break
				}
				abs := resolveURL(base, href)
				if abs == "" {
					break
				}
				switch {
				case strings.Contains(rel, "stylesheet"):
					// Catches "stylesheet", "preload stylesheet", etc.
					if !isGoogleFonts(abs) {
						cssSet[abs] = true
					}
				case rel == "modulepreload":
					// JS module chunks
					jsSet[abs] = true
				case strings.Contains(rel, "icon"):
					// Catches "icon", "shortcut icon", "apple-touch-icon"
					binarySet[abs] = true
				case rel == "preload":
					as := strings.ToLower(getAttr(n, "as"))
					switch as {
					case "image", "font":
						binarySet[abs] = true
					case "script":
						jsSet[abs] = true
					case "style":
						if !isGoogleFonts(abs) {
							cssSet[abs] = true
						}
					}
				}
			case "script":
				src := getAttr(n, "src")
				if src != "" {
					if abs := resolveURL(base, src); abs != "" {
						jsSet[abs] = true
					}
				}
			case "img":
				if src := getAttr(n, "src"); src != "" {
					if abs := resolveURL(base, src); abs != "" {
						binarySet[abs] = true
					}
				}
				// Handle srcset: "img.png 1x, img2x.png 2x"
				if srcset := getAttr(n, "srcset"); srcset != "" {
					for _, u := range parseSrcset(srcset, base) {
						binarySet[u] = true
					}
				}
			case "source":
				if src := getAttr(n, "src"); src != "" {
					if abs := resolveURL(base, src); abs != "" {
						binarySet[abs] = true
					}
				}
			case "video", "audio":
				if src := getAttr(n, "src"); src != "" {
					if abs := resolveURL(base, src); abs != "" {
						binarySet[abs] = true
					}
				}
				if poster := getAttr(n, "poster"); poster != "" {
					if abs := resolveURL(base, poster); abs != "" {
						binarySet[abs] = true
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	for u := range cssSet {
		cssURLs = append(cssURLs, u)
	}
	for u := range jsSet {
		jsURLs = append(jsURLs, u)
	}
	for u := range binarySet {
		binaryURLs = append(binaryURLs, u)
	}
	return
}

// resolveURL converts any href/src to an absolute URL relative to base.
// Returns empty string if the ref is a data URI or otherwise unresolvable.
func resolveURL(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || strings.HasPrefix(ref, "data:") || strings.HasPrefix(ref, "javascript:") || strings.HasPrefix(ref, "#") {
		return ""
	}
	// Protocol-relative
	if strings.HasPrefix(ref, "//") {
		ref = base.Scheme + ":" + ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(refURL)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	return abs.String()
}

// extractCSSURLs scans CSS content for url(...) references and returns
// absolute URLs, resolving relative refs against cssBaseURL.
func extractCSSURLs(cssContent, cssBaseURL string) []string {
	cssBase, err := url.Parse(cssBaseURL)
	if err != nil {
		return nil
	}

	matches := cssURLRegex.FindAllStringSubmatch(cssContent, -1)
	var result []string
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := strings.TrimSpace(m[1])
		abs := resolveURL(cssBase, ref)
		if abs != "" {
			result = append(result, abs)
		}
	}
	return result
}

// parseSrcset splits a srcset attribute and returns absolute URLs.
func parseSrcset(srcset string, base *url.URL) []string {
	var urls []string
	for _, part := range strings.Split(srcset, ",") {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) > 0 {
			if abs := resolveURL(base, fields[0]); abs != "" {
				urls = append(urls, abs)
			}
		}
	}
	return urls
}

// rewriteHTMLPaths updates src/href attributes in the document to use local paths.
// It resolves relative attribute values against base before looking up in urlToLocal,
// so both absolute and relative references are correctly rewritten.
func rewriteHTMLPaths(doc *html.Node, urlToLocal map[string]string, base *url.URL) {
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "link":
				rewriteAttr(n, "href", urlToLocal, base)
			case "script":
				rewriteAttr(n, "src", urlToLocal, base)
			case "img", "source", "video", "audio":
				rewriteAttr(n, "src", urlToLocal, base)
				rewriteAttr(n, "poster", urlToLocal, base)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
}

// rewriteAttr rewrites a single attribute on a node. It first tries a direct
// lookup, then resolves to an absolute URL and retries, handling relative paths.
func rewriteAttr(n *html.Node, attr string, urlToLocal map[string]string, base *url.URL) {
	val := getAttr(n, attr)
	if val == "" {
		return
	}
	// Direct match (attribute already contains absolute URL)
	if local, ok := urlToLocal[val]; ok {
		setAttr(n, attr, "/"+local)
		return
	}
	// Resolve relative to absolute and retry
	abs := resolveURL(base, val)
	if abs != "" {
		if local, ok := urlToLocal[abs]; ok {
			setAttr(n, attr, "/"+local)
		}
	}
}

// extractInlineResources extracts inline <style> and <script> blocks,
// replacing them with file references. Mirrors the extractor package logic.
func extractInlineResources(n *html.Node, cssContent, jsContent *strings.Builder, inlineCSS, inlineJS *[]extractor.InlineResource, cssIndex, jsIndex *int) {
	if n.Type == html.ElementNode {
		if n.Data == "style" {
			content := collectTextContent(n)
			if strings.TrimSpace(content) != "" {
				*cssIndex++
				filename := fmt.Sprintf("inline/style-%d.css", *cssIndex)
				*inlineCSS = append(*inlineCSS, extractor.InlineResource{Path: filename, Content: content})
				cssContent.WriteString(content)
				if !strings.HasSuffix(content, "\n") {
					cssContent.WriteString("\n")
				}
				link := &html.Node{
					Type: html.ElementNode,
					Data: "link",
					Attr: []html.Attribute{
						{Key: "rel", Val: "stylesheet"},
						{Key: "href", Val: "/" + filename},
					},
				}
				replaceNode(n, link)
				return
			}
		} else if n.Data == "script" && !hasAttrKey(n, "src") {
			content := collectTextContent(n)
			if strings.TrimSpace(content) != "" {
				*jsIndex++
				filename := fmt.Sprintf("inline/script-%d.js", *jsIndex)
				*inlineJS = append(*inlineJS, extractor.InlineResource{Path: filename, Content: content})
				jsContent.WriteString(content)
				if !strings.HasSuffix(content, "\n") {
					jsContent.WriteString("\n")
				}
				script := &html.Node{
					Type: html.ElementNode,
					Data: "script",
					Attr: []html.Attribute{
						{Key: "src", Val: "/" + filename},
					},
				}
				replaceNode(n, script)
				return
			}
		}
	}

	for c := n.FirstChild; c != nil; {
		next := c.NextSibling
		extractInlineResources(c, cssContent, jsContent, inlineCSS, inlineJS, cssIndex, jsIndex)
		c = next
	}
}

func collectTextContent(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}
	return sb.String()
}

func replaceNode(old, new *html.Node) {
	if old.Parent == nil {
		return
	}
	old.Parent.InsertBefore(new, old)
	old.Parent.RemoveChild(old)
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func setAttr(n *html.Node, key, val string) {
	for i, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			n.Attr[i].Val = val
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: val})
}

func hasAttrKey(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return true
		}
	}
	return false
}

func isGoogleFonts(u string) bool {
	return strings.Contains(u, "fonts.googleapis.com")
}

func deduplicateStrings(ss []string) []string {
	seen := make(map[string]bool, len(ss))
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// binaryFilename creates a safe, unique filename for a binary asset.
func binaryFilename(rawURL, mime string, used map[string]int) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Sprintf("asset-%d%s", len(used), mimeExt(mime))
	}

	base := path.Base(parsed.Path)
	if base == "" || base == "." || base == "/" {
		base = "asset"
	}

	// Sanitize
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, base)

	// Ensure extension
	if !strings.Contains(base, ".") {
		base += mimeExt(mime)
	}

	original := base
	counter := 1
	for used[base] > 0 {
		ext := path.Ext(original)
		stem := strings.TrimSuffix(original, ext)
		base = fmt.Sprintf("%s-%d%s", stem, counter, ext)
		counter++
	}
	used[base]++
	return base
}

func mimeExt(mime string) string {
	switch mime {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	case "image/x-icon", "image/vnd.microsoft.icon":
		return ".ico"
	case "font/woff":
		return ".woff"
	case "font/woff2":
		return ".woff2"
	case "font/ttf", "application/x-font-ttf":
		return ".ttf"
	case "font/otf", "application/x-font-otf":
		return ".otf"
	default:
		return ".bin"
	}
}
