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
var cssImportRegex = regexp.MustCompile(`(?i)@import\s+(?:url\(\s*['"]?([^'")\s]+)['"]?\s*\)|["']([^"']+)["'])`)

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
	installWebflowLocalRuntimeCompat(doc, base)

	cssURLs, jsURLs, binaryURLs := findAllAssetURLs(doc, base)
	retainedAssets := findRetainedExternalAssets(doc, base)
	capture := extractor.NewCaptureManifest(rawURL)

	// Build a URL→localPath map for path rewriting
	urlToLocal := make(map[string]string)

	// Fetch CSS and JS as text resources
	var externalJS []fetcher.FetchedResource

	externalCSS, cssBinaryURLs := collectCSSResources(cssURLs, fetchCSSResource, capture, urlToLocal)
	binaryURLs = append(binaryURLs, cssBinaryURLs...)

	if len(jsURLs) > 0 {
		externalJS = fetcher.FetchExternalResources(jsURLs, "js")
		for _, r := range externalJS {
			if r.Error == nil && r.Content != "" {
				urlToLocal[r.URL] = "external/js/" + r.Filename
				capture.AddAsset(extractor.CaptureAsset{
					URL: r.URL, Path: "external/js/" + r.Filename,
					Type: "js", Status: extractor.CaptureLocalized,
				})
			} else {
				capture.AddAsset(failedCaptureAsset(r.URL, "js", r.Error))
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
			capture.AddAsset(failedCaptureAsset(bURL, "asset", err))
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
		capture.AddAsset(extractor.CaptureAsset{
			URL: bURL, Path: localPath, Type: "asset", Status: extractor.CaptureLocalized,
		})
	}

	for _, retained := range retainedAssets {
		capture.AddAsset(extractor.CaptureAsset{
			URL: retained.URL, Type: retained.Type, Status: extractor.CaptureRetainedExternal,
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
		Capture:     capture,
	}, nil
}

const webflowCompatPrepare = `(function(){var r=document.documentElement,d=r.getAttribute("data-wf-domain");if(!d||location.hostname===d)return;window.__unclusterWebflowDomain=d;r.setAttribute("data-wf-domain",location.hostname)})();`

const webflowCompatRestore = `(function(){var d=window.__unclusterWebflowDomain;if(!d)return;var f=function(){setTimeout(function(){document.documentElement.setAttribute("data-wf-domain",d);delete window.__unclusterWebflowDomain},250)};if(document.readyState==="loading")addEventListener("DOMContentLoaded",f,{once:true});else f()})();`

// installWebflowLocalRuntimeCompat preserves the behavior of a page served on
// its own *.webflow.io hostname when the export is opened from another host.
// Webflow's brand module otherwise treats the hostname change as a reason to
// inject a remote badge that is absent from the captured source page.
func installWebflowLocalRuntimeCompat(doc *html.Node, base *url.URL) bool {
	var documentElement, head, body *html.Node
	walkElements(doc, func(node *html.Node) {
		switch node.Data {
		case "html":
			documentElement = node
		case "head":
			head = node
		case "body":
			body = node
		}
	})
	if documentElement == nil || head == nil || body == nil || base == nil {
		return false
	}

	domain := strings.ToLower(strings.TrimSuffix(htmlutil.GetAttr(documentElement, "data-wf-domain"), "."))
	if !strings.HasSuffix(domain, ".webflow.io") || !strings.EqualFold(domain, base.Hostname()) {
		return false
	}

	prepare := inlineScriptNode(webflowCompatPrepare)
	if head.FirstChild == nil {
		head.AppendChild(prepare)
	} else {
		head.InsertBefore(prepare, head.FirstChild)
	}
	body.AppendChild(inlineScriptNode(webflowCompatRestore))
	return true
}

func inlineScriptNode(source string) *html.Node {
	script := &html.Node{Type: html.ElementNode, Data: "script"}
	script.AppendChild(&html.Node{Type: html.TextNode, Data: source})
	return script
}

func collectCSSResources(cssURLs []string, fetchOne cssFetcher, capture *extractor.CaptureManifest, urlToLocal map[string]string) ([]fetcher.FetchedResource, []string) {
	resources, binaryURLs := fetchCSSResources(cssURLs, fetchOne)
	for _, resource := range resources {
		if resource.Error == nil && resource.Content != "" {
			localPath := "external/css/" + resource.Filename
			urlToLocal[resource.URL] = localPath
			capture.AddAsset(extractor.CaptureAsset{
				URL: resource.URL, Path: localPath,
				Type: "css", Status: extractor.CaptureLocalized,
			})
		} else {
			capture.AddAsset(failedCaptureAsset(resource.URL, "css", resource.Error))
		}
	}
	return resources, binaryURLs
}

func failedCaptureAsset(rawURL, assetType string, err error) extractor.CaptureAsset {
	message := "empty response body"
	if err != nil {
		message = err.Error()
	}
	return extractor.CaptureAsset{
		URL: rawURL, Type: assetType, Status: extractor.CaptureFailed, Error: message,
	}
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
	seen := map[assetKind]map[string]bool{
		assetCSS:    {},
		assetJS:     {},
		assetBinary: {},
	}

	add := func(kind assetKind, abs string) {
		if abs == "" || kind == assetNone || seen[kind][abs] {
			return
		}
		seen[kind][abs] = true
		switch kind {
		case assetCSS:
			cssURLs = append(cssURLs, abs)
		case assetJS:
			jsURLs = append(jsURLs, abs)
		case assetBinary:
			binaryURLs = append(binaryURLs, abs)
		}
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

		if abs := resolveDataSrcURL(base, htmlutil.GetAttr(n, "data-src")); abs != "" {
			add(assetBinary, abs)
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

	return
}

type captureReference struct {
	URL  string
	Type string
}

// findRetainedExternalAssets records render/runtime dependencies that remain
// intentionally external. Navigation and connection-metadata URLs are not
// capture assets and are deliberately excluded.
func findRetainedExternalAssets(doc *html.Node, base *url.URL) []captureReference {
	seen := make(map[string]bool)
	var retained []captureReference
	add := func(rawURL, assetType string) {
		absolute := resolveURL(base, rawURL)
		key := assetType + "\x00" + absolute
		if absolute == "" || seen[key] {
			return
		}
		seen[key] = true
		retained = append(retained, captureReference{URL: absolute, Type: assetType})
	}

	walkElements(doc, func(n *html.Node) {
		switch n.Data {
		case "link":
			href := htmlutil.GetAttr(n, "href")
			if strings.Contains(strings.ToLower(htmlutil.GetAttr(n, "rel")), "stylesheet") && htmlutil.IsGoogleFonts(href) {
				add(href, "css")
			}
		case "iframe":
			add(htmlutil.GetAttr(n, "src"), "iframe")
		}
	})
	return retained
}

// resolveDataSrcURL recognizes URL-bearing data-src values used by Webflow
// lazy-loading and Lottie embeds without treating ordinary component metadata
// (for example data-src="carousel-panel") as a network asset.
func resolveDataSrcURL(base *url.URL, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}

	parsed, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	pathLike := strings.HasPrefix(ref, "//") || strings.HasPrefix(ref, "/") ||
		strings.HasPrefix(ref, "./") || strings.HasPrefix(ref, "../") ||
		parsed.IsAbs() || strings.Contains(parsed.Path, "/") || path.Ext(parsed.Path) != ""
	if !pathLike {
		return ""
	}
	return resolveURL(base, ref)
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
	_, assets := extractCSSReferences(cssContent, cssBaseURL)
	return assets
}

// extractCSSReferences returns imported stylesheets and non-import url()
// assets separately, preserving their order within each category.
func extractCSSReferences(cssContent, cssBaseURL string) (imports, assets []string) {
	cssBase, err := url.Parse(cssBaseURL)
	if err != nil {
		return nil, nil
	}

	importMatches := cssImportRegex.FindAllStringSubmatchIndex(cssContent, -1)
	for _, match := range importMatches {
		var ref string
		switch {
		case match[2] >= 0:
			ref = cssContent[match[2]:match[3]]
		case match[4] >= 0:
			ref = cssContent[match[4]:match[5]]
		}
		if abs := resolveURL(cssBase, ref); abs != "" {
			imports = append(imports, abs)
		}
	}

	matches := cssURLRegex.FindAllStringSubmatchIndex(cssContent, -1)
	for _, m := range matches {
		if len(m) < 4 || insideAnyRange(m[0], m[1], importMatches) {
			continue
		}
		ref := strings.TrimSpace(cssContent[m[2]:m[3]])
		abs := resolveURL(cssBase, ref)
		if abs != "" {
			assets = append(assets, abs)
		}
	}
	return imports, assets
}

func insideAnyRange(start, end int, ranges [][]int) bool {
	for _, candidate := range ranges {
		if start >= candidate[0] && end <= candidate[1] {
			return true
		}
	}
	return false
}

type cssFetcher func(string) fetcher.FetchedResource

func fetchCSSResource(rawURL string) fetcher.FetchedResource {
	resources := fetcher.FetchExternalResources([]string{rawURL}, "css")
	if len(resources) == 0 {
		return fetcher.FetchedResource{URL: rawURL, Type: "css", Error: fmt.Errorf("stylesheet fetch returned no result")}
	}
	return resources[0]
}

// fetchCSSResources follows @import references depth-first in stylesheet source
// order. The seen set makes cycles safe and keeps each URL emitted once.
func fetchCSSResources(rootURLs []string, fetchOne cssFetcher) ([]fetcher.FetchedResource, []string) {
	seen := make(map[string]bool)
	usedFilenames := make(map[string]int)
	var resources []fetcher.FetchedResource
	var binaryURLs []string

	var visit func(string, bool)
	visit = func(rawURL string, imported bool) {
		if seen[rawURL] {
			return
		}
		seen[rawURL] = true

		resource := fetchOne(rawURL)
		resource.URL = rawURL
		resource.Type = "css"
		if imported {
			resource.Type = "css-import"
		}
		if resource.Error == nil {
			resource.Filename = uniqueResourceFilename(resource.Filename, usedFilenames)
		}
		resources = append(resources, resource)
		if resource.Error != nil {
			return
		}

		imports, assets := extractCSSReferences(resource.Content, resource.URL)
		for _, importedURL := range imports {
			visit(importedURL, true)
		}
		binaryURLs = append(binaryURLs, assets...)
	}

	for _, rootURL := range rootURLs {
		visit(rootURL, false)
	}
	return resources, deduplicateStrings(binaryURLs)
}

func uniqueResourceFilename(filename string, used map[string]int) string {
	if used[filename] == 0 {
		used[filename] = 1
		return filename
	}

	ext := filepath.Ext(filename)
	stem := strings.TrimSuffix(filename, ext)
	for suffix := used[filename]; ; suffix++ {
		candidate := fmt.Sprintf("%s-%d%s", stem, suffix, ext)
		if used[candidate] == 0 {
			used[filename] = suffix + 1
			used[candidate] = 1
			return candidate
		}
	}
}

// rewriteCSSURLs replaces downloaded url(...) dependencies with paths that are
// relative to the downloaded stylesheet's location in the exported archive.
func rewriteCSSURLs(cssContent, cssBaseURL, cssLocalPath string, urlToLocal map[string]string) string {
	cssBase, err := url.Parse(cssBaseURL)
	if err != nil {
		return cssContent
	}

	cssDir := filepath.Dir(filepath.FromSlash(cssLocalPath))
	rewritten := cssURLRegex.ReplaceAllStringFunc(cssContent, func(match string) string {
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

	return cssImportRegex.ReplaceAllStringFunc(rewritten, func(match string) string {
		parts := cssImportRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		ref := parts[1]
		if ref == "" {
			ref = parts[2]
		}
		absolute := resolveURL(cssBase, ref)
		localPath, ok := urlToLocal[absolute]
		if absolute == "" || !ok {
			return match
		}
		relative, err := filepath.Rel(cssDir, filepath.FromSlash(localPath))
		if err != nil {
			return match
		}
		return strings.Replace(match, ref, filepath.ToSlash(relative), 1)
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
		if resolveDataSrcURL(base, htmlutil.GetAttr(n, "data-src")) != "" {
			rewriteAttr(n, "data-src", urlToLocal, base)
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
