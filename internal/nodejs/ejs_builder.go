package nodejs

import (
	"bytes"
	"fmt"
	"htmlfmt/internal/extractor"
	"htmlfmt/internal/fetcher"
	"htmlfmt/internal/formatter"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/net/html"
)

type EJSProjectConfig struct {
	ProjectName string
	HTML        string
	InlineCSS   []extractor.InlineResource
	InlineJS    []extractor.InlineResource
	ExternalCSS []fetcher.FetchedResource
	ExternalJS  []fetcher.FetchedResource
}

type ejsComponent struct {
	Name string
	HTML string
	Node *html.Node
}

func GenerateEJSProject(config *EJSProjectConfig) (*ProjectFiles, error) {
	files := make(map[string]string)

	packageJSON, err := generateEJSPackageJSON(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate package.json: %w", err)
	}
	files["package.json"] = packageJSON
	files["server.js"] = ejsServerJSTemplate
	files[".gitignore"] = gitignoreTemplate

	readme, err := generateEJSReadme(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate README: %w", err)
	}
	files["README.md"] = readme

	indexHTML, partials, err := generateEJSViews(config.HTML)
	if err != nil {
		return nil, fmt.Errorf("failed to generate views: %w", err)
	}
	files["views/index.ejs"] = indexHTML

	for name, content := range partials {
		files["views/partials/"+name+".ejs"] = content
	}

	for _, css := range config.InlineCSS {
		if strings.TrimSpace(css.Content) != "" {
			files["public/"+css.Path] = css.Content
		}
	}

	for _, js := range config.InlineJS {
		if strings.TrimSpace(js.Content) != "" {
			files["public/"+js.Path] = js.Content
		}
	}

	for _, css := range config.ExternalCSS {
		if css.Error == nil && strings.TrimSpace(css.Content) != "" {
			files["public/external/css/"+css.Filename] = css.Content
		}
	}

	for _, js := range config.ExternalJS {
		if js.Error == nil && strings.TrimSpace(js.Content) != "" {
			files["public/external/js/"+js.Filename] = js.Content
		}
	}

	return &ProjectFiles{Files: files}, nil
}

func generateEJSPackageJSON(config *EJSProjectConfig) (string, error) {
	tmpl, err := template.New("package.json").Parse(ejsPackageJSONTemplate)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = tmpl.Execute(&buf, config)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateEJSReadme(config *EJSProjectConfig) (string, error) {
	tmpl, err := template.New("README.md").Parse(ejsReadmeTemplate)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = tmpl.Execute(&buf, config)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func generateEJSViews(htmlContent string) (string, map[string]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, err
	}

	body := findElement(doc, "body")
	if body == nil {
		return htmlContent, map[string]string{}, nil
	}

	root := selectComponentRoot(body)
	components := collectBodyComponents(root)

	if len(components) == 0 {
		return htmlContent, map[string]string{}, nil
	}

	usedNames := make(map[string]int)
	nameByContent := make(map[string]string)
	var resolved []ejsComponent

	for idx, component := range components {
		content, err := renderNodeHTML(component.Node)
		if err != nil {
			continue
		}
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			continue
		}

		name, ok := nameByContent[trimmed]
		if !ok {
			name = buildComponentName(component.Node, idx, usedNames)
			nameByContent[trimmed] = name
		}

		resolved = append(resolved, ejsComponent{
			Name: name,
			HTML: content,
			Node: component.Node,
		})

		replaceNodeWithIncludeMarker(component.Node, name)
	}

	components = resolved

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", nil, err
	}

	rendered := buf.String()
	if formatted, err := formatter.Format(rendered); err == nil {
		rendered = formatted
	}

	indexReplacements := buildIncludeReplacements(components, "partials/")
	partialReplacements := buildIncludeReplacements(components, "")
	rendered = applyIncludeReplacements(rendered, indexReplacements)

	partials := make(map[string]string, len(components))
	for _, component := range components {
		if _, exists := partials[component.Name]; exists {
			continue
		}
		partials[component.Name] = applyIncludeReplacements(component.HTML, partialReplacements)
	}

	return rendered, partials, nil
}

func collectBodyComponents(root *html.Node) []ejsComponent {
	nodes := selectComponentNodes(root)
	if len(nodes) == 0 {
		return nil
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodeDepth(nodes[i]) > nodeDepth(nodes[j])
	})

	var components []ejsComponent
	for _, child := range nodes {
		if !isComponentCandidate(child) {
			continue
		}
		components = append(components, ejsComponent{
			Name: "",
			HTML: "",
			Node: child,
		})
	}

	return components
}

func selectComponentRoot(body *html.Node) *html.Node {
	root := body
	for depth := 0; depth < 4; depth++ {
		children := contentChildren(root)
		if len(children) != 1 {
			break
		}
		child := children[0]
		if isWrapperElement(child) {
			root = child
			continue
		}
		break
	}
	return root
}

func isComponentCandidate(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if isNonContentElement(n) || isEmbedOnlyNode(n) {
		return false
	}
	if getAttributeValue(n, "data-component") != "" {
		return true
	}
	switch n.Data {
	case "html", "head", "body":
		return false
	default:
		return true
	}
}

func isWrapperElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch n.Data {
	case "div", "main", "section":
		classAttr := strings.ToLower(getAttributeValue(n, "class"))
		idAttr := strings.ToLower(getAttributeValue(n, "id"))
		wrapperHints := []string{
			"wrapper", "container", "page", "main", "layout", "root", "app", "site", "content",
		}
		for _, hint := range wrapperHints {
			if strings.Contains(classAttr, hint) || strings.Contains(idAttr, hint) {
				return true
			}
		}
		return true
	default:
		return false
	}
}

func selectComponentNodes(root *html.Node) []*html.Node {
	sections := collectSectionComponents(root, 5)
	if len(sections) > 1 {
		return sections
	}

	children := filterComponentCandidates(contentChildren(root))
	if len(children) > 1 {
		return children
	}

	if len(children) == 1 {
		deeper := filterComponentCandidates(contentChildren(children[0]))
		if len(deeper) > 1 {
			return deeper
		}
	}

	return children
}

func filterComponentCandidates(nodes []*html.Node) []*html.Node {
	var filtered []*html.Node
	for _, node := range nodes {
		if isComponentCandidate(node) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func collectSectionComponents(root *html.Node, maxDepth int) []*html.Node {
	var nodes []*html.Node

	var walk func(n *html.Node, depth int)
	walk = func(n *html.Node, depth int) {
		if depth >= maxDepth {
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type != html.ElementNode {
				continue
			}
			if isSectionBoundary(child) {
				nodes = append(nodes, child)
				continue
			}
			walk(child, depth+1)
		}
	}

	walk(root, 0)
	return nodes
}

func isSectionBoundary(n *html.Node) bool {
	if isNonContentElement(n) || isEmbedOnlyNode(n) {
		return false
	}
	// 'main' is treated as a transparent container — we recurse through it
	// to find its section children rather than extracting it as one giant partial.
	switch n.Data {
	case "nav", "header", "footer", "section", "aside":
		return true
	}

	// For non-semantic elements, only match if a class or the id is exactly a known keyword.
	classes := strings.Fields(strings.ToLower(getAttributeValue(n, "class")))
	id := strings.ToLower(getAttributeValue(n, "id"))
	keywords := []string{
		"navbar", "nav", "header", "footer", "hero", "section",
	}
	for _, keyword := range keywords {
		if id == keyword {
			return true
		}
		for _, class := range classes {
			if class == keyword {
				return true
			}
		}
	}

	return false
}

func buildComponentName(n *html.Node, index int, used map[string]int) string {
	base := n.Data
	if id := getAttributeValue(n, "id"); id != "" {
		base += "-" + id
	} else if classAttr := getAttributeValue(n, "class"); classAttr != "" {
		if firstClass := strings.Fields(classAttr); len(firstClass) > 0 {
			base += "-" + firstClass[0]
		}
	}

	base = sanitizeComponentName(base)
	if base == "" {
		base = fmt.Sprintf("component-%d", index+1)
	}

	if count, ok := used[base]; ok {
		count++
		used[base] = count
		base = fmt.Sprintf("%s-%d", base, count)
	} else {
		used[base] = 1
	}

	return base
}

func sanitizeComponentName(name string) string {
	var b strings.Builder
	b.Grow(len(name))

	lastDash := false
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			r = r - 'A' + 'a'
		}

		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}

		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}

	s := strings.Trim(b.String(), "-")
	return s
}

func contentChildren(n *html.Node) []*html.Node {
	var children []*html.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		if isNonContentElement(child) {
			continue
		}
		children = append(children, child)
	}
	return children
}

func isNonContentElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return true
	}
	switch n.Data {
	case "script", "style", "link", "meta", "title", "noscript",
		"svg", "path", "circle", "rect", "line", "polygon", "polyline", "defs", "g", "use":
		return true
	default:
		return false
	}
}

func isEmbedOnlyNode(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	classAttr := strings.ToLower(getAttributeValue(n, "class"))
	if !strings.Contains(classAttr, "w-embed") {
		return false
	}

	hasElement := false
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			hasElement = true
			switch child.Data {
			case "style", "script", "link":
				continue
			default:
				return false
			}
		case html.TextNode:
			if strings.TrimSpace(child.Data) != "" {
				return true
			}
		}
	}

	return hasElement
}

func nodeDepth(n *html.Node) int {
	depth := 0
	for current := n.Parent; current != nil; current = current.Parent {
		depth++
	}
	return depth
}

func uniqueNodes(nodes []*html.Node) []*html.Node {
	seen := make(map[*html.Node]bool, len(nodes))
	unique := make([]*html.Node, 0, len(nodes))
	for _, node := range nodes {
		if node == nil || seen[node] {
			continue
		}
		seen[node] = true
		unique = append(unique, node)
	}
	return unique
}

func buildIncludeReplacements(components []ejsComponent, prefix string) map[string]string {
	replacements := make(map[string]string, len(components))
	for _, component := range components {
		placeholder := "<!--EJS_INCLUDE:" + component.Name + "-->"
		include := "<%- include('" + prefix + component.Name + "') %>"
		replacements[placeholder] = include
	}
	return replacements
}

func applyIncludeReplacements(content string, replacements map[string]string) string {
	updated := content
	for placeholder, include := range replacements {
		updated = strings.ReplaceAll(updated, placeholder, include)
	}
	return updated
}

func replaceNodeWithIncludeMarker(n *html.Node, name string) {
	if n.Parent == nil {
		return
	}
	comment := &html.Node{
		Type: html.CommentNode,
		Data: "EJS_INCLUDE:" + name,
	}
	n.Parent.InsertBefore(comment, n)
	n.Parent.RemoveChild(n)
}

func renderNodeHTML(n *html.Node) (string, error) {
	var buf bytes.Buffer
	if err := html.Render(&buf, n); err != nil {
		return "", err
	}
	return buf.String(), nil
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

func getAttributeValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, key) {
			return attr.Val
		}
	}
	return ""
}
