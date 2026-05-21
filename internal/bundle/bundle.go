package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/nodejs"

	"golang.org/x/net/html"
)

type Result struct {
	SiteName  string
	OutputDir string
	IndexPath string
	SplitDir  string
	EJSDir    string
}

type Options struct {
	OutputBase  string
	Destination string
}

type sourceBundle struct {
	inputPath string
	rootDir   string
	htmlPath  string
	html      string
	siteName  string
	cleanup   func()
}

type localAsset struct {
	Path    string
	Content []byte
}

type indexCandidate struct {
	path       string
	relPath    string
	content    string
	size       int
	matchScore int
}

var cssURLPattern = regexp.MustCompile(`url\(\s*['"]?([^'")\s]+)['"]?\s*\)`)

func Process(inputPath, outputBase string) (*Result, error) {
	return ProcessWithOptions(inputPath, Options{OutputBase: outputBase})
}

func ProcessWithOptions(inputPath string, options Options) (*Result, error) {
	source, err := loadSource(inputPath)
	if err != nil {
		return nil, err
	}
	defer source.cleanup()

	siteName := source.siteName
	if siteName == "" {
		siteName = deriveSiteName(inputPath, source.html)
	}

	outputDir := options.Destination
	if outputDir == "" {
		outputDir = filepath.Join(options.OutputBase, siteName)
	}
	splitDir := filepath.Join(outputDir, "unzip")
	ejsDir := filepath.Join(outputDir, "ejs")

	rewrittenHTML, assets, err := rewriteLocalAssets(source.html, filepath.Dir(source.htmlPath), source.rootDir)
	if err != nil {
		return nil, err
	}

	if err := resetGeneratedOutput(outputDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create site output directory: %w", err)
	}
	indexPath := filepath.Join(outputDir, "index.html")
	if err := os.WriteFile(indexPath, []byte(source.html), 0o644); err != nil {
		return nil, fmt.Errorf("write source index.html: %w", err)
	}

	extracted, err := extractor.Extract(rewrittenHTML)
	if err != nil {
		return nil, fmt.Errorf("extract split resources: %w", err)
	}

	if err := writeSplitOutput(extracted, assets, inputPath, splitDir); err != nil {
		return nil, err
	}

	if err := writeEJSOutput(extracted, assets, filepath.Base(ejsDir), ejsDir); err != nil {
		return nil, err
	}

	return &Result{
		SiteName:  siteName,
		OutputDir: outputDir,
		IndexPath: indexPath,
		SplitDir:  splitDir,
		EJSDir:    ejsDir,
	}, nil
}

func loadSource(inputPath string) (*sourceBundle, error) {
	abs, err := filepath.Abs(inputPath)
	if err != nil {
		return nil, fmt.Errorf("resolve input path: %w", err)
	}

	if strings.EqualFold(filepath.Ext(abs), ".zip") {
		return loadZipSource(abs)
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read input HTML: %w", err)
	}

	return &sourceBundle{
		inputPath: abs,
		rootDir:   filepath.Dir(abs),
		htmlPath:  abs,
		html:      string(raw),
		siteName:  deriveSiteName(abs, string(raw)),
		cleanup:   func() {},
	}, nil
}

func loadZipSource(zipPath string) (*sourceBundle, error) {
	tempDir, err := os.MkdirTemp("", "uncluster-bundle-*")
	if err != nil {
		return nil, fmt.Errorf("create temp extraction directory: %w", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	if err := extractZip(zipPath, tempDir); err != nil {
		cleanup()
		return nil, err
	}

	candidate, err := selectIndexHTML(tempDir, strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath)))
	if err != nil {
		cleanup()
		return nil, err
	}

	return &sourceBundle{
		inputPath: zipPath,
		rootDir:   tempDir,
		htmlPath:  candidate.path,
		html:      candidate.content,
		siteName:  sanitizeSiteName(strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))),
		cleanup:   cleanup,
	}, nil
}

func extractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open ZIP input: %w", err)
	}
	defer reader.Close()

	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("resolve temp extraction path: %w", err)
	}

	for _, file := range reader.File {
		target := filepath.Join(destAbs, filepath.Clean(file.Name))
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("resolve ZIP entry path %q: %w", file.Name, err)
		}
		if targetAbs != destAbs && !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) {
			return fmt.Errorf("unsafe ZIP entry path %q", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetAbs, 0o755); err != nil {
				return fmt.Errorf("create ZIP directory %q: %w", file.Name, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
			return fmt.Errorf("create ZIP entry directory %q: %w", file.Name, err)
		}

		src, err := file.Open()
		if err != nil {
			return fmt.Errorf("open ZIP entry %q: %w", file.Name, err)
		}
		data, readErr := io.ReadAll(src)
		closeErr := src.Close()
		if readErr != nil {
			return fmt.Errorf("read ZIP entry %q: %w", file.Name, readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close ZIP entry %q: %w", file.Name, closeErr)
		}
		if err := os.WriteFile(targetAbs, data, file.FileInfo().Mode()); err != nil {
			return fmt.Errorf("write ZIP entry %q: %w", file.Name, err)
		}
	}

	return nil
}

func selectIndexHTML(rootDir, sourceName string) (*indexCandidate, error) {
	var candidates []indexCandidate
	sourceKey := normalizeForMatch(sourceName)

	err := filepath.WalkDir(rootDir, func(filePath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.EqualFold(entry.Name(), "index.html") {
			return nil
		}

		raw, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		content := string(raw)
		if !looksLikeHTML(content) {
			return nil
		}

		rel, err := filepath.Rel(rootDir, filePath)
		if err != nil {
			return err
		}
		score := candidateScore(rel, sourceKey)

		candidates = append(candidates, indexCandidate{
			path:       filePath,
			relPath:    filepath.ToSlash(rel),
			content:    content,
			size:       len(raw),
			matchScore: score,
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan ZIP for index.html: %w", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no usable index.html found in ZIP")
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].matchScore != candidates[j].matchScore {
			return candidates[i].matchScore > candidates[j].matchScore
		}
		if candidates[i].size != candidates[j].size {
			return candidates[i].size > candidates[j].size
		}
		return candidates[i].relPath < candidates[j].relPath
	})

	return &candidates[0], nil
}

func candidateScore(relPath, sourceKey string) int {
	dir := filepath.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	parent := normalizeForMatch(filepath.Base(dir))
	pathKey := normalizeForMatch(dir)

	score := 0
	if sourceKey != "" {
		switch {
		case parent == sourceKey:
			score += 100
		case parent != "" && (strings.Contains(parent, sourceKey) || strings.Contains(sourceKey, parent)):
			score += 75
		case pathKey != "" && strings.Contains(pathKey, sourceKey):
			score += 50
		}
	}

	depth := len(strings.Split(filepath.ToSlash(dir), "/"))
	if dir == "" {
		depth = 0
	}
	score -= depth
	return score
}

func looksLikeHTML(content string) bool {
	lower := strings.ToLower(content)
	return strings.Contains(lower, "<html") ||
		strings.Contains(lower, "<body") ||
		strings.Contains(lower, "<!doctype")
}

func rewriteLocalAssets(htmlContent, htmlDir, rootDir string) (string, []localAsset, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, fmt.Errorf("parse HTML for local asset rewrite: %w", err)
	}

	refs := collectHTMLRefs(doc, htmlDir, rootDir)
	assetMap, err := buildAssetMap(refs, htmlDir, rootDir)
	if err != nil {
		return "", nil, err
	}

	rewriteHTMLRefs(doc, htmlDir, rootDir, assetMap)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", nil, fmt.Errorf("render HTML after local asset rewrite: %w", err)
	}

	assets := make([]localAsset, 0, len(assetMap))
	for _, asset := range assetMap {
		assets = append(assets, asset)
	}
	sort.Slice(assets, func(i, j int) bool {
		return assets[i].Path < assets[j].Path
	})

	return buf.String(), assets, nil
}

func collectHTMLRefs(doc *html.Node, htmlDir, rootDir string) []string {
	var refs []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i := range n.Attr {
				attr := &n.Attr[i]
				switch strings.ToLower(attr.Key) {
				case "src", "href", "poster":
					refs = append(refs, attr.Val)
				case "srcset":
					for _, item := range parseSrcset(attr.Val) {
						refs = append(refs, item.url)
					}
				case "style":
					refs = append(refs, extractCSSURLs(attr.Val)...)
				}
			}
			if n.Data == "style" {
				refs = append(refs, extractCSSURLs(collectNodeText(n))...)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return filterLocalRefs(refs, htmlDir, rootDir)
}

func buildAssetMap(refs []string, htmlDir, rootDir string) (map[string]localAsset, error) {
	assets := make(map[string]localAsset)
	queue := append([]string(nil), refs...)
	seen := make(map[string]bool)

	for len(queue) > 0 {
		absPath := queue[0]
		queue = queue[1:]
		if seen[absPath] {
			continue
		}
		seen[absPath] = true

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		outPath := assetOutputPath(absPath, htmlDir)
		assets[absPath] = localAsset{Path: outPath, Content: data}

		if strings.EqualFold(filepath.Ext(absPath), ".css") {
			cssRefs := extractCSSURLs(string(data))
			for _, ref := range filterLocalRefs(cssRefs, filepath.Dir(absPath), rootDir) {
				if !seen[ref] {
					queue = append(queue, ref)
				}
			}
		}
	}

	return assets, nil
}

func rewriteHTMLRefs(doc *html.Node, htmlDir, rootDir string, assetMap map[string]localAsset) {
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i := range n.Attr {
				attr := &n.Attr[i]
				switch strings.ToLower(attr.Key) {
				case "src", "href", "poster":
					if replacement, ok := localReplacement(attr.Val, htmlDir, rootDir, assetMap); ok {
						attr.Val = replacement
					}
				case "srcset":
					attr.Val = rewriteSrcset(attr.Val, htmlDir, rootDir, assetMap)
				case "style":
					attr.Val = rewriteCSSURLs(attr.Val, htmlDir, rootDir, assetMap)
				}
			}
			if n.Data == "style" {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						c.Data = rewriteCSSURLs(c.Data, htmlDir, rootDir, assetMap)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
}

func filterLocalRefs(refs []string, baseDir, rootDir string) []string {
	var out []string
	for _, ref := range refs {
		absPath, ok := resolveLocalRef(ref, baseDir, rootDir)
		if ok {
			out = append(out, absPath)
		}
	}
	return out
}

func resolveLocalRef(rawRef, baseDir, rootDir string) (string, bool) {
	ref := strings.TrimSpace(rawRef)
	if ref == "" || strings.HasPrefix(ref, "#") {
		return "", false
	}
	if parsed, err := url.Parse(ref); err == nil {
		if parsed.Scheme != "" || parsed.Host != "" {
			return "", false
		}
		ref = parsed.Path
	}
	ref = strings.TrimPrefix(ref, "/")
	if ref == "" {
		return "", false
	}

	candidate := filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(ref)))
	rootAbs, err := filepath.Abs(rootDir)
	if err != nil {
		return "", false
	}
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", false
	}
	if candidateAbs != rootAbs && !strings.HasPrefix(candidateAbs, rootAbs+string(os.PathSeparator)) {
		return "", false
	}
	if info, err := os.Stat(candidateAbs); err != nil || info.IsDir() {
		return "", false
	}
	return candidateAbs, true
}

func localReplacement(rawRef, baseDir, rootDir string, assetMap map[string]localAsset) (string, bool) {
	absPath, ok := resolveLocalRef(rawRef, baseDir, rootDir)
	if !ok {
		return "", false
	}
	asset, ok := assetMap[absPath]
	if !ok {
		return "", false
	}
	return asset.Path, true
}

func assetOutputPath(absPath, htmlDir string) string {
	rel, err := filepath.Rel(htmlDir, absPath)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		rel = filepath.Base(absPath)
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	rel = strings.TrimPrefix(rel, "../")
	rel = strings.TrimPrefix(rel, "/")
	return path.Join("assets", rel)
}

type srcsetItem struct {
	url        string
	descriptor string
}

func parseSrcset(value string) []srcsetItem {
	var items []srcsetItem
	for _, part := range strings.Split(value, ",") {
		fields := strings.Fields(strings.TrimSpace(part))
		if len(fields) == 0 {
			continue
		}
		item := srcsetItem{url: fields[0]}
		if len(fields) > 1 {
			item.descriptor = strings.Join(fields[1:], " ")
		}
		items = append(items, item)
	}
	return items
}

func rewriteSrcset(value, baseDir, rootDir string, assetMap map[string]localAsset) string {
	items := parseSrcset(value)
	parts := make([]string, 0, len(items))
	for _, item := range items {
		urlValue := item.url
		if replacement, ok := localReplacement(item.url, baseDir, rootDir, assetMap); ok {
			urlValue = replacement
		}
		if item.descriptor != "" {
			parts = append(parts, urlValue+" "+item.descriptor)
		} else {
			parts = append(parts, urlValue)
		}
	}
	return strings.Join(parts, ", ")
}

func extractCSSURLs(css string) []string {
	matches := cssURLPattern.FindAllStringSubmatch(css, -1)
	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 {
			refs = append(refs, match[1])
		}
	}
	return refs
}

func rewriteCSSURLs(css, baseDir, rootDir string, assetMap map[string]localAsset) string {
	return cssURLPattern.ReplaceAllStringFunc(css, func(match string) string {
		parts := cssURLPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		if replacement, ok := localReplacement(parts[1], baseDir, rootDir, assetMap); ok {
			return "url(" + replacement + ")"
		}
		return match
	})
}

func collectNodeText(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

func writeSplitOutput(extracted *extractor.ExtractedContent, assets []localAsset, inputPath, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create split output directory: %w", err)
	}
	if err := writeText(filepath.Join(outDir, "index.html"), extracted.HTML); err != nil {
		return fmt.Errorf("write split index.html: %w", err)
	}

	for _, r := range extracted.InlineCSS {
		if err := writeText(filepath.Join(outDir, filepath.FromSlash(r.Path)), r.Content); err != nil {
			return fmt.Errorf("write split inline CSS: %w", err)
		}
	}
	for _, r := range extracted.InlineJS {
		if err := writeText(filepath.Join(outDir, filepath.FromSlash(r.Path)), r.Content); err != nil {
			return fmt.Errorf("write split inline JS: %w", err)
		}
	}

	var externalCSS []string
	for _, r := range extracted.ExternalCSS {
		if r.Error != nil || r.Content == "" {
			continue
		}
		rel := path.Join("external", "css", r.Filename)
		if err := writeText(filepath.Join(outDir, filepath.FromSlash(rel)), r.Content); err != nil {
			return fmt.Errorf("write split external CSS: %w", err)
		}
		externalCSS = append(externalCSS, rel)
	}

	var externalJS []string
	for _, r := range extracted.ExternalJS {
		if r.Error != nil || r.Content == "" {
			continue
		}
		rel := path.Join("external", "js", r.Filename)
		if err := writeText(filepath.Join(outDir, filepath.FromSlash(rel)), r.Content); err != nil {
			return fmt.Errorf("write split external JS: %w", err)
		}
		externalJS = append(externalJS, rel)
	}

	for _, asset := range assets {
		if err := writeBytes(filepath.Join(outDir, filepath.FromSlash(asset.Path)), asset.Content); err != nil {
			return fmt.Errorf("write split local asset %s: %w", asset.Path, err)
		}
	}

	manifest := map[string]interface{}{
		"created_at":       time.Now().Format(time.RFC3339),
		"input_path":       inputPath,
		"output_path":      outDir,
		"html_file":        "index.html",
		"inline_css_count": len(extracted.InlineCSS),
		"inline_js_count":  len(extracted.InlineJS),
		"external_css":     externalCSS,
		"external_js":      externalJS,
		"local_assets":     localAssetPaths(assets),
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal split manifest: %w", err)
	}
	if err := writeText(filepath.Join(outDir, "split-manifest.json"), string(data)); err != nil {
		return fmt.Errorf("write split manifest: %w", err)
	}

	return nil
}

func resetGeneratedOutput(outputDir string) error {
	for _, rel := range []string{"unzip", "ejs", "zips"} {
		if err := os.RemoveAll(filepath.Join(outputDir, rel)); err != nil {
			return fmt.Errorf("remove stale %s output: %w", rel, err)
		}
	}
	return nil
}

func writeEJSOutput(extracted *extractor.ExtractedContent, assets []localAsset, projectName, outDir string) error {
	projectFiles, err := nodejs.GenerateEJSProject(&nodejs.EJSProjectConfig{
		ProjectName: projectName,
		HTML:        extracted.RewriteForEJS(),
		InlineCSS:   extracted.InlineCSS,
		InlineJS:    extracted.InlineJS,
		ExternalCSS: extracted.ExternalCSS,
		ExternalJS:  extracted.ExternalJS,
	})
	if err != nil {
		return fmt.Errorf("generate EJS project: %w", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create EJS output directory: %w", err)
	}

	for relPath, content := range projectFiles.Files {
		if err := writeText(filepath.Join(outDir, filepath.FromSlash(relPath)), content); err != nil {
			return fmt.Errorf("write EJS file %s: %w", relPath, err)
		}
	}

	for _, asset := range assets {
		rel := path.Join("public", asset.Path)
		if err := writeBytes(filepath.Join(outDir, filepath.FromSlash(rel)), asset.Content); err != nil {
			return fmt.Errorf("write EJS local asset %s: %w", rel, err)
		}
	}

	return nil
}

func writeText(path, content string) error {
	return writeBytes(path, []byte(content))
}

func writeBytes(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func localAssetPaths(assets []localAsset) []string {
	paths := make([]string, 0, len(assets))
	for _, asset := range assets {
		paths = append(paths, asset.Path)
	}
	sort.Strings(paths)
	return paths
}

func deriveSiteName(inputPath, htmlContent string) string {
	stem := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if name := sanitizeSiteName(stem); name != "" && name != "index" {
		return name
	}

	if host := htmlURLHost(htmlContent); host != "" {
		return sanitizeSiteName(host)
	}
	if title := htmlTitle(htmlContent); title != "" {
		return sanitizeSiteName(title)
	}
	if name := sanitizeSiteName(stem); name != "" {
		return name
	}
	return "site"
}

func htmlURLHost(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}
	var result string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if result != "" {
			return
		}
		if n.Type == html.ElementNode {
			if n.Data == "link" && strings.EqualFold(getAttr(n, "rel"), "canonical") {
				result = hostFromURL(getAttr(n, "href"))
				return
			}
			if n.Data == "meta" {
				key := strings.ToLower(getAttr(n, "property"))
				if key == "" {
					key = strings.ToLower(getAttr(n, "name"))
				}
				if key == "og:url" || key == "twitter:url" {
					result = hostFromURL(getAttr(n, "content"))
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return result
}

func htmlTitle(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}
	var result string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if result != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "title" {
			result = strings.TrimSpace(collectNodeText(n))
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return result
}

func hostFromURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}

func sanitizeSiteName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.', r == '_':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == ' ':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-._")
}

func normalizeForMatch(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(value) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}
