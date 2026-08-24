package scraper

import (
	"bytes"
	"fmt"
	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"
	"github.com/omariomari2/uncluster/internal/formatter"
	"github.com/omariomari2/uncluster/internal/htmlutil"
	"github.com/omariomari2/uncluster/internal/safehttp"
	"log"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var cssURLRegex = regexp.MustCompile(`url\(\s*['"]?([^'")\s]+)['"]?\s*\)`)

// ScrapeURL fetches a webpage and all its referenced assets (CSS, JS, images,
// fonts, SVGs) and returns an ExtractedContent ready for the export pipeline.
func ScrapeURL(rawURL string) (*extractor.ExtractedContent, error) {
	if err := safehttp.ValidateURL(rawURL); err != nil {
		return nil, err
	}

	base, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
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

	for i := range externalCSS {
		if externalCSS[i].Error != nil {
			continue
		}
		localCSSPath := "external/css/" + externalCSS[i].Filename
		externalCSS[i].Content = rewriteCSSURLs(
			externalCSS[i].Content,
			externalCSS[i].URL,
			localCSSPath,
			urlToLocal,
		)
	}

	// Rewrite src/href in the document to local relative paths
	rewriteHTMLPaths(doc, urlToLocal, base)

	// Extract inline <style> and <script> tags, sharing the upload path's logic.
	var inline htmlutil.InlineCollector
	inline.Walk(doc)

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
		CSS:         inline.CSS.String(),
		JS:          inline.JS.String(),
		InlineCSS:   inline.InlineCSS,
		InlineJS:    inline.InlineJS,
		ExternalCSS: externalCSS,
		ExternalJS:  externalJS,
		LocalAssets: localAssets,
	}, nil
}

// fetchPage downloads the HTML content of a URL with a browser User-Agent.
func fetchPage(rawURL string) (string, error) {
	client := safehttp.Client(30 * time.Second)

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", safehttp.BrowserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := safehttp.ReadBody(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// assetKind classifies where a discovered URL belongs in the export.
type assetKind int

const (
	assetBinary assetKind = iota
	assetCSS
	assetJS
	assetNone
)

// assetAttrs is the single description of which element attributes hold asset
// URLs. Both findAllAssetURLs and rewriteHTMLPaths iterate it, so discovery and
// rewriting cannot drift apart — a mismatch between the two is what caused
// srcset assets to be downloaded but never referenced.
//
// <link> is deliberately absent: its classification depends on rel and as
// rather than on the tag, so it is handled by linkKind instead.
var assetAttrs = []struct {
	tag    string
	attr   string
	kind   assetKind
	srcset bool
}{
	{tag: "script", attr: "src", kind: assetJS},
	{tag: "img", attr: "src", kind: assetBinary},
	{tag: "img", attr: "srcset", kind: assetBinary, srcset: true},
	{tag: "source", attr: "src", kind: assetBinary},
	{tag: "source", attr: "srcset", kind: assetBinary, srcset: true},
	{tag: "video", attr: "src", kind: assetBinary},
	{tag: "video", attr: "poster", kind: assetBinary},
	{tag: "audio", attr: "src", kind: assetBinary},
}

// walkElements visits every element node in the tree.
func walkElements(n *html.Node, visit func(*html.Node)) {
	if n.Type == html.ElementNode {
		visit(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkElements(c, visit)
	}
}

// linkKind classifies a <link> by its rel and as attributes.
func linkKind(n *html.Node, abs string) assetKind {
	rel := strings.ToLower(htmlutil.GetAttr(n, "rel"))

	switch {
	case strings.Contains(rel, "stylesheet"):
		// Catches "stylesheet", "preload stylesheet", etc.
		if htmlutil.IsGoogleFonts(abs) {
			return assetNone
		}
		return assetCSS
	case rel == "modulepreload":
		// JS module chunks
		return assetJS
	case strings.Contains(rel, "icon"):
		// Catches "icon", "shortcut icon", "apple-touch-icon"
		return assetBinary
	case rel == "preload":
		switch strings.ToLower(htmlutil.GetAttr(n, "as")) {
		case "image", "font":
			return assetBinary
		case "script":
			return assetJS
		case "style":
			if htmlutil.IsGoogleFonts(abs) {
				return assetNone
			}
			return assetCSS
		}
	}
	return assetNone
}

// findAllAssetURLs walks the HTML tree and collects absolute URLs for
// CSS, JS, and binary assets (images, fonts, SVGs).
func findAllAssetURLs(doc *html.Node, base *url.URL) (cssURLs, jsURLs, binaryURLs []string) {
	sets := map[assetKind]map[string]bool{
		assetCSS:    {},
		assetJS:     {},
		assetBinary: {},
	}

	add := func(kind assetKind, abs string) {
		if abs == "" || kind == assetNone {
			return
		}
		sets[kind][abs] = true
	}

	walkElements(doc, func(n *html.Node) {
		if n.Data == "link" {
			href := htmlutil.GetAttr(n, "href")
			if href == "" {
				return
			}
			if abs := resolveURL(base, href); abs != "" {
				add(linkKind(n, abs), abs)
			}
			return
		}

		for _, ref := range assetAttrs {
			if ref.tag != n.Data {
				continue
			}
			val := htmlutil.GetAttr(n, ref.attr)
			if val == "" {
				continue
			}
			if ref.srcset {
				// "img.png 1x, img2x.png 2x"
				for _, u := range parseSrcset(val, base) {
					add(ref.kind, u)
				}
				continue
			}
			add(ref.kind, resolveURL(base, val))
		}
	})

	for u := range sets[assetCSS] {
		cssURLs = append(cssURLs, u)
	}
	for u := range sets[assetJS] {
		jsURLs = append(jsURLs, u)
	}
	for u := range sets[assetBinary] {
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

// rewriteCSSURLs replaces downloaded url(...) dependencies with paths that are
// relative to the downloaded stylesheet's location in the exported archive.
func rewriteCSSURLs(cssContent, cssBaseURL, cssLocalPath string, urlToLocal map[string]string) string {
	cssBase, err := url.Parse(cssBaseURL)
	if err != nil {
		return cssContent
	}

	cssDir := filepath.Dir(filepath.FromSlash(cssLocalPath))
	return cssURLRegex.ReplaceAllStringFunc(cssContent, func(match string) string {
		parts := cssURLRegex.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		absolute := resolveURL(cssBase, parts[1])
		localPath, ok := urlToLocal[absolute]
		if absolute == "" || !ok {
			return match
		}

		relative, err := filepath.Rel(cssDir, filepath.FromSlash(localPath))
		if err != nil {
			return match
		}
		relative = filepath.ToSlash(relative)
		return strings.Replace(match, parts[1], relative, 1)
	})
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

// rewriteHTMLPaths updates asset attributes in the document to use local paths.
// It resolves relative attribute values against base before looking up in urlToLocal,
// so both absolute and relative references are correctly rewritten.
//
// It walks the same assetAttrs table as findAllAssetURLs, so an attribute can
// never be discovered without also being rewritten.
func rewriteHTMLPaths(doc *html.Node, urlToLocal map[string]string, base *url.URL) {
	walkElements(doc, func(n *html.Node) {
		if n.Data == "link" {
			rewriteAttr(n, "href", urlToLocal, base)
			return
		}
		for _, ref := range assetAttrs {
			if ref.tag != n.Data {
				continue
			}
			if ref.srcset {
				rewriteSrcset(n, ref.attr, urlToLocal, base)
			} else {
				rewriteAttr(n, ref.attr, urlToLocal, base)
			}
		}
	})
}

func rewriteSrcset(n *html.Node, attr string, urlToLocal map[string]string, base *url.URL) {
	srcset := htmlutil.GetAttr(n, attr)
	if srcset == "" {
		return
	}

	candidates := strings.Split(srcset, ",")
	for i, candidate := range candidates {
		fields := strings.Fields(strings.TrimSpace(candidate))
		if len(fields) == 0 {
			continue
		}

		if local, ok := urlToLocal[fields[0]]; ok {
			fields[0] = local
		} else if absolute := resolveURL(base, fields[0]); absolute != "" {
			if local, ok := urlToLocal[absolute]; ok {
				fields[0] = local
			}
		}
		candidates[i] = strings.Join(fields, " ")
	}

	htmlutil.SetAttr(n, attr, strings.Join(candidates, ", "))
}

// rewriteAttr rewrites a single attribute on a node. It first tries a direct
// lookup, then resolves to an absolute URL and retries, handling relative paths.
func rewriteAttr(n *html.Node, attr string, urlToLocal map[string]string, base *url.URL) {
	val := htmlutil.GetAttr(n, attr)
	if val == "" {
		return
	}
	// Direct match (attribute already contains absolute URL)
	if local, ok := urlToLocal[val]; ok {
		htmlutil.SetAttr(n, attr, local)
		return
	}
	// Resolve relative to absolute and retry
	abs := resolveURL(base, val)
	if abs != "" {
		if local, ok := urlToLocal[abs]; ok {
			htmlutil.SetAttr(n, attr, local)
		}
	}
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
