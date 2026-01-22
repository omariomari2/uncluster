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

// EJSProjectConfig represents the configuration for generating an EJS project.
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

// GenerateEJSProject creates a complete Express + EJS project from the given configuration.
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

	replacements := buildIncludeReplacements(components)
	rendered = applyIncludeReplacements(rendered, replacements)

	partials := make(map[string]string, len(components))
	for _, component := range components {
		if _, exists := partials[component.Name]; exists {
			continue
		}
		partials[component.Name] = applyIncludeReplacements(component.HTML, replacements)
	}

	return rendered, partials, nil
}

func collectBodyComponents(root *html.Node) []ejsComponent {
	primaryNodes := selectComponentNodes(root)
	if len(primaryNodes) == 0 {
		return nil
	}

	nestedNodes := collectNestedComponents(primaryNodes, 6)
	nodes := append([]*html.Node{}, nestedNodes...)
	nodes = append(nodes, primaryNodes...)
	nodes = uniqueNodes(nodes)

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
	switch n.Data {
	case "nav", "header", "footer", "section", "main", "aside":
		return true
	}

	classAttr := strings.ToLower(getAttributeValue(n, "class"))
	idAttr := strings.ToLower(getAttributeValue(n, "id"))
	combined := classAttr + " " + idAttr
	if combined == " " {
		return false
	}

	keywords := []string{
		"navbar", "nav", "menu", "header", "footer", "hero", "section", "cta",
		"pricing", "gallery", "grid", "slider", "carousel", "tabs", "accordion", "form",
	}
	for _, keyword := range keywords {
		if strings.Contains(combined, keyword) {
			return true
		}
	}

	return false
}

func collectNestedComponents(roots []*html.Node, maxDepth int) []*html.Node {
	var nested []*html.Node
	for _, root := range roots {
		nested = append(nested, collectNestedComponentsInRoot(root, maxDepth)...)
	}
	return nested
}

func collectNestedComponentsInRoot(root *html.Node, maxDepth int) []*html.Node {
	patternCounts := make(map[string]int)

	var walkCount func(n *html.Node, depth int)
	walkCount = func(n *html.Node, depth int) {
		if depth > maxDepth {
			return
		}
		if depth > 0 && isComponentCandidate(n) {
			if key := componentPatternKey(n); key != "" {
				patternCounts[key]++
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				walkCount(child, depth+1)
			}
		}
	}

	var walkSelect func(n *html.Node, depth int)
	var selected []*html.Node
	walkSelect = func(n *html.Node, depth int) {
		if depth > maxDepth {
			return
		}
		if depth > 0 && isComponentCandidate(n) {
			key := componentPatternKey(n)
			if shouldSelectNestedComponent(n, patternCounts[key]) {
				selected = append(selected, n)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				walkSelect(child, depth+1)
			}
		}
	}

	walkCount(root, 0)
	walkSelect(root, 0)

	return selected
}

func shouldSelectNestedComponent(n *html.Node, patternCount int) bool {
	if isNonContentElement(n) || isEmbedOnlyNode(n) {
		return false
	}

	keywordMatch := hasComponentKeyword(n)
	if isLayoutContainer(n) && !keywordMatch {
		return false
	}

	if isButtonElement(n) {
		return true
	}

	if keywordMatch && (elementChildCount(n) > 0 || nodeTextLength(n) > 10) {
		return true
	}

	if patternCount >= 2 && hasIdentifyingClassOrID(n) && hasMeaningfulContent(n) {
		return true
	}

	return false
}

func hasMeaningfulContent(n *html.Node) bool {
	return elementChildCount(n) >= 2 || nodeTextLength(n) >= 20
}

func hasIdentifyingClassOrID(n *html.Node) bool {
	return getAttributeValue(n, "class") != "" || getAttributeValue(n, "id") != ""
}

func componentPatternKey(n *html.Node) string {
	if n.Type != html.ElementNode {
		return ""
	}
	classes := normalizeClassList(strings.Fields(getAttributeValue(n, "class")))
	if len(classes) == 0 {
		return n.Data
	}
	return n.Data + "." + strings.Join(classes, ".")
}

func normalizeClassList(classes []string) []string {
	if len(classes) == 0 {
		return nil
	}

	seen := make(map[string]bool, len(classes))
	filtered := make([]string, 0, len(classes))
	for _, className := range classes {
		className = strings.ToLower(strings.TrimSpace(className))
		if className == "" || isStateClass(className) {
			continue
		}
		if seen[className] {
			continue
		}
		seen[className] = true
		filtered = append(filtered, className)
	}
	sort.Strings(filtered)
	return filtered
}

func isStateClass(className string) bool {
	switch className {
	case "active", "current", "open", "closed", "selected":
		return true
	default:
	}
	return strings.HasPrefix(className, "is-") ||
		strings.HasPrefix(className, "has-") ||
		strings.HasPrefix(className, "js-") ||
		strings.HasPrefix(className, "w--")
}

func hasComponentKeyword(n *html.Node) bool {
	classAttr := strings.ToLower(getAttributeValue(n, "class"))
	idAttr := strings.ToLower(getAttributeValue(n, "id"))
	combined := classAttr + " " + idAttr
	if strings.TrimSpace(combined) == "" {
		return false
	}

	keywords := []string{
		"navbar", "nav", "menu", "header", "footer", "hero", "section", "cta",
		"button", "btn", "card", "tile", "banner", "pricing", "gallery", "feature",
		"testimonial", "form", "input", "field", "dropdown", "modal", "popup",
		"slider", "carousel", "tabs", "accordion",
	}
	for _, keyword := range keywords {
		if strings.Contains(combined, keyword) {
			return true
		}
	}

	return false
}

func isButtonElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if n.Data == "button" {
		return true
	}
	role := strings.ToLower(getAttributeValue(n, "role"))
	if role == "button" {
		return true
	}
	if n.Data == "a" {
		classAttr := strings.ToLower(getAttributeValue(n, "class"))
		if strings.Contains(classAttr, "button") || strings.Contains(classAttr, "btn") {
			return true
		}
	}
	return false
}

func isLayoutContainer(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	classAttr := strings.ToLower(getAttributeValue(n, "class"))
	idAttr := strings.ToLower(getAttributeValue(n, "id"))
	wrapperHints := []string{
		"wrapper", "container", "page", "layout", "root", "app", "site", "content",
	}
	for _, hint := range wrapperHints {
		if strings.Contains(classAttr, hint) || strings.Contains(idAttr, hint) {
			return true
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

func elementChildren(n *html.Node) []*html.Node {
	var children []*html.Node
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			children = append(children, child)
		}
	}
	return children
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

func elementChildCount(n *html.Node) int {
	return len(elementChildren(n))
}

func nodeTextLength(n *html.Node) int {
	length := 0
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			trimmed := strings.TrimSpace(node.Data)
			if trimmed != "" {
				length += len(trimmed)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(n)
	return length
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

func buildIncludeReplacements(components []ejsComponent) map[string]string {
	replacements := make(map[string]string, len(components))
	for _, component := range components {
		placeholder := "<!--EJS_INCLUDE:" + component.Name + "-->"
		include := "<%- include('partials/" + component.Name + "') %>"
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
