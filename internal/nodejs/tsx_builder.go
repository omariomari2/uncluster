package nodejs

import (
	"fmt"
	"log"
	"strings"

	"github.com/omariomari2/uncluster/internal/converter"
	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"
	"github.com/omariomari2/uncluster/internal/htmlutil"

	"golang.org/x/net/html"
)

type tsxComponent struct {
	Name string
	HTML string
	Node *html.Node
}

// generateTSXViews finds semantic sections in htmlContent, converts each to a
// TSX component, and returns:
//   - sectionFiles: map "src/components/<Name>.tsx" → file content
//   - mainComponent: content of MainComponent.tsx (imports + renders all sections)
//   - mainTsx: content of src/main.tsx (dynamic CSS imports)
func generateTSXViews(
	htmlContent string,
	inlineCSS string,
	inlineJS string,
	externalCSS []fetcher.FetchedResource,
	externalJS []fetcher.FetchedResource,
) (sectionFiles map[string]string, mainComponent string, mainTsx string, err error) {

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, "", "", err
	}

	body := htmlutil.FindElement(doc, "body")
	if body == nil {
		mc, convErr := converter.ConvertSectionToTSX(htmlContent, "MainComponent")
		if convErr != nil {
			return nil, "", "", convErr
		}
		return map[string]string{}, mc, generateMainTsx(htmlContent, inlineCSS, inlineJS, externalCSS, externalJS), nil
	}

	root := selectComponentRoot(body)
	sections := filterComponentCandidates(contentChildren(root))

	if len(sections) == 0 {
		mc, convErr := converter.ConvertSectionToTSX(htmlContent, "MainComponent")
		if convErr != nil {
			return nil, "", "", convErr
		}
		return map[string]string{}, mc, generateMainTsx(htmlContent, inlineCSS, inlineJS, externalCSS, externalJS), nil
	}

	usedNames := make(map[string]int)
	var resolved []tsxComponent

	for idx, node := range sections {
		rawHTML, renderErr := renderNodeHTML(node)
		if renderErr != nil {
			log.Printf("tsx_builder: failed to render section node %d: %v", idx, renderErr)
			continue
		}
		trimmed := strings.TrimSpace(rawHTML)
		if trimmed == "" {
			continue
		}

		kebab := buildComponentName(node, idx, usedNames)
		name := toPascalCase(kebab)

		resolved = append(resolved, tsxComponent{Name: name, HTML: rawHTML, Node: node})
	}

	if len(resolved) == 0 {
		mc, convErr := converter.ConvertSectionToTSX(htmlContent, "MainComponent")
		if convErr != nil {
			return nil, "", "", convErr
		}
		return map[string]string{}, mc, generateMainTsx(htmlContent, inlineCSS, inlineJS, externalCSS, externalJS), nil
	}

	sectionFiles = make(map[string]string, len(resolved))
	seen := make(map[string]bool)
	for _, comp := range resolved {
		if seen[comp.Name] {
			continue
		}
		seen[comp.Name] = true

		tsxContent, convErr := converter.ConvertSectionToTSX(comp.HTML, comp.Name)
		if convErr != nil {
			log.Printf("tsx_builder: failed to convert section %q: %v", comp.Name, convErr)
			continue
		}
		sectionFiles["src/components/"+comp.Name+".tsx"] = tsxContent
	}

	main, mainErr := buildMainComponent(body, resolved)
	if mainErr != nil {
		log.Printf("tsx_builder: falling back to a flat component list: %v", mainErr)
		main = generateMainComponentTSX(resolved)
	}

	return sectionFiles, main, generateMainTsx(htmlContent, inlineCSS, inlineJS, externalCSS, externalJS), nil
}

// componentMarkerTag is the placeholder element left where a component was
// lifted out. It has to be an element rather than a comment: the converter
// drops an element whose only children are comments, which would erase the
// markers along with the structure they sit in.
const componentMarkerTag = "uncluster-component"

// buildMainComponent converts the whole body, with each extracted section
// replaced in place by its component. Composing this way keeps the markup that
// surrounds and nests the sections — page wrappers, <main>, layout containers —
// which a flat list of components would silently drop.
func buildMainComponent(body *html.Node, components []tsxComponent) (string, error) {
	for _, comp := range components {
		replaceNodeWithComponentMarker(comp.Node, comp.Name)
	}

	bodyHTML, err := renderNodeHTML(body)
	if err != nil {
		return "", fmt.Errorf("render body for main component: %w", err)
	}

	converted, err := converter.ConvertSectionToTSX(bodyHTML, "MainComponent")
	if err != nil {
		return "", fmt.Errorf("convert body for main component: %w", err)
	}

	var imports strings.Builder
	seen := make(map[string]bool)
	for _, comp := range components {
		if seen[comp.Name] {
			continue
		}
		seen[comp.Name] = true
		imports.WriteString(fmt.Sprintf("import %s from './%s'\n", comp.Name, comp.Name))
		converted = strings.ReplaceAll(converted, componentMarkerJSX(comp.Name), "<"+comp.Name+" />")
	}

	if imports.Len() > 0 {
		converted = imports.String() + "\n" + converted
	}
	return converted, nil
}

func replaceNodeWithComponentMarker(n *html.Node, name string) {
	if n.Parent == nil {
		return
	}
	marker := &html.Node{
		Type: html.ElementNode,
		Data: componentMarkerTag,
		Attr: []html.Attribute{{Key: "data-name", Val: name}},
	}
	n.Parent.InsertBefore(marker, n)
	n.Parent.RemoveChild(n)
}

// componentMarkerJSX is how the placeholder element comes back out of the
// converter: an element with no children renders as an open/close pair.
func componentMarkerJSX(name string) string {
	return fmt.Sprintf(`<%s data-name="%s"></%s>`, componentMarkerTag, name, componentMarkerTag)
}

func toPascalCase(s string) string {
	if s == "" {
		return "Section"
	}
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '-' || r == '_' })
	var b strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	result := b.String()
	if result == "" {
		return "Section"
	}
	if result[0] >= '0' && result[0] <= '9' {
		result = "Section" + result
	}
	return result
}

func generateMainComponentTSX(sections []tsxComponent) string {
	var imports strings.Builder
	var jsxLines strings.Builder

	seen := make(map[string]bool)
	for _, comp := range sections {
		if seen[comp.Name] {
			continue
		}
		seen[comp.Name] = true
		imports.WriteString(fmt.Sprintf("import %s from './%s'\n", comp.Name, comp.Name))
		jsxLines.WriteString(fmt.Sprintf("      <%s />\n", comp.Name))
	}

	return fmt.Sprintf(`%s
function MainComponent() {
  return (
    <>
%s    </>
  )
}

export default MainComponent
`, imports.String(), jsxLines.String())
}

func generateMainTsx(
	htmlContent string,
	inlineCSS string,
	inlineJS string,
	externalCSS []fetcher.FetchedResource,
	externalJS []fetcher.FetchedResource,
	inlineResources ...[]extractor.InlineResource,
) string {
	var cssImports strings.Builder
	if strings.TrimSpace(inlineCSS) != "" {
		cssImports.WriteString("import './styles/main.css'\n")
	}
	for _, res := range externalCSS {
		if res.Error == nil && res.Type != "css-import" && strings.TrimSpace(res.Content) != "" {
			cssImports.WriteString(fmt.Sprintf("import './styles/external/%s'\n", res.Filename))
		}
	}

	var inlineScripts []extractor.InlineResource
	if len(inlineResources) > 0 {
		inlineScripts = inlineResources[0]
	}
	scripts := originalScripts(htmlContent, inlineJS, inlineScripts, externalJS)

	var scriptLoader strings.Builder
	if len(scripts) > 0 {
		scriptLoader.WriteString(`
type OriginalScript = {
  src: string
  parent: 'head' | 'body'
  attributes: Record<string, string>
}

const originalScripts: OriginalScript[] = [
`)
		for _, script := range scripts {
			scriptLoader.WriteString("  { src: " + quoteTypeScriptString(script.Src) + ", parent: " + quoteTypeScriptString(script.Parent) + ", attributes: {")
			for i, attr := range script.Attributes {
				if i > 0 {
					scriptLoader.WriteString(", ")
				}
				scriptLoader.WriteString(quoteTypeScriptString(attr.Name) + ": " + quoteTypeScriptString(attr.Value))
			}
			scriptLoader.WriteString("} },\n")
		}
		scriptLoader.WriteString(`]

function loadOriginalScript(original: OriginalScript, index: number): Promise<void> {
	const marker = ` + "`" + `uncluster-original-${index}` + "`" + `
	if (document.querySelector(` + "`" + `script[data-uncluster-script="${marker}"]` + "`" + `)) {
		return Promise.resolve()
	}

  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
		script.async = false
		for (const [name, value] of Object.entries(original.attributes)) {
			script.setAttribute(name, value)
		}
		script.dataset.unclusterScript = marker
		script.src = original.src
    script.onload = () => resolve()
		script.onerror = () => reject(new Error(` + "`" + `Failed to load ${original.src}` + "`" + `))
		const parent = original.parent === 'head' ? document.head : document.body
		parent.appendChild(script)
  })
}

async function loadOriginalScripts() {
	for (const [index, original] of originalScripts.entries()) {
		await loadOriginalScript(original, index)
  }
}

function HydrationComplete(): null {
	useEffect(() => {
		void loadOriginalScripts()
	}, [])
	return null
}
`)
	}

	reactImport := ""
	hydration := "hydrateRoot(document.body, <App />)\n"
	if len(scripts) > 0 {
		reactImport = "import { useEffect } from 'react'\n"
		hydration = `hydrateRoot(
  document.body,
  <>
    <App />
    <HydrationComplete />
  </>,
)
`
	}

	return fmt.Sprintf(`%simport { hydrateRoot } from 'react-dom/client'
import App from './App'
%s
%s
%s`, reactImport, cssImports.String(), scriptLoader.String(), hydration)
}

type originalScript struct {
	Src        string
	Parent     string
	Attributes []scriptAttribute
}

type scriptAttribute struct {
	Name  string
	Value string
}

func originalScripts(htmlContent, inlineJS string, inlineScripts []extractor.InlineResource, externalJS []fetcher.FetchedResource) []originalScript {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return fallbackOriginalScripts(inlineJS, inlineScripts, externalJS)
	}

	externalPaths := make(map[string]string)
	for _, resource := range externalJS {
		if resource.Error != nil || strings.TrimSpace(resource.Content) == "" {
			continue
		}
		path := "/scripts/external/" + resource.Filename
		externalPaths[resource.URL] = path
		externalPaths[resource.Filename] = path
	}
	inlinePaths := make(map[string]string)
	for _, resource := range inlineScripts {
		if strings.HasPrefix(resource.Path, "inline/") && strings.TrimSpace(resource.Content) != "" {
			inlinePaths[resource.Path] = "/scripts/" + resource.Path
			inlinePaths["/"+resource.Path] = "/scripts/" + resource.Path
		}
	}

	seen := make(map[string]bool)
	var scripts []originalScript
	var walk func(*html.Node, string)
	walk = func(n *html.Node, parent string) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "head":
				parent = "head"
			case "body":
				parent = "body"
			case "script":
				if htmlutil.IsJavaScriptType(htmlutil.GetAttr(n, "type")) {
					src := canonicalScriptPath(htmlutil.GetAttr(n, "src"), inlineJS, inlinePaths, externalPaths)
					if src != "" && !seen[src] {
						seen[src] = true
						scripts = append(scripts, originalScript{
							Src:        src,
							Parent:     scriptParent(parent),
							Attributes: originalScriptAttributes(n),
						})
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child, parent)
		}
	}
	walk(doc, "body")

	for _, fallback := range fallbackOriginalScripts(inlineJS, inlineScripts, externalJS) {
		if !seen[fallback.Src] {
			seen[fallback.Src] = true
			scripts = append(scripts, fallback)
		}
	}
	return scripts
}

func fallbackOriginalScripts(inlineJS string, inlineScripts []extractor.InlineResource, externalJS []fetcher.FetchedResource) []originalScript {
	var scripts []originalScript
	for _, resource := range inlineScripts {
		if strings.HasPrefix(resource.Path, "inline/") && strings.TrimSpace(resource.Content) != "" {
			scripts = append(scripts, originalScript{Src: "/scripts/" + resource.Path, Parent: "body"})
		}
	}
	if len(inlineScripts) == 0 && strings.TrimSpace(inlineJS) != "" {
		scripts = append(scripts, originalScript{Src: "/scripts/main.js", Parent: "body"})
	}
	for _, resource := range externalJS {
		if resource.Error == nil && strings.TrimSpace(resource.Content) != "" {
			scripts = append(scripts, originalScript{Src: "/scripts/external/" + resource.Filename, Parent: "body"})
		}
	}
	return scripts
}

func canonicalScriptPath(src, inlineJS string, inlinePaths, externalPaths map[string]string) string {
	src = strings.TrimSpace(src)
	if src == "" {
		if strings.TrimSpace(inlineJS) != "" {
			return "/scripts/main.js"
		}
		return ""
	}
	if path, ok := externalPaths[src]; ok {
		return path
	}
	if path, ok := inlinePaths[src]; ok {
		return path
	}

	trimmed := strings.TrimPrefix(src, "/")
	switch {
	case trimmed == "script.js", trimmed == "scripts/main.js", strings.HasPrefix(trimmed, "inline/"):
		if strings.TrimSpace(inlineJS) != "" {
			return "/scripts/main.js"
		}
	case strings.HasPrefix(trimmed, "external/js/"):
		return "/scripts/external/" + strings.TrimPrefix(trimmed, "external/js/")
	case strings.HasPrefix(trimmed, "scripts/external/"):
		return "/" + trimmed
	default:
		return src
	}
	return ""
}

func originalScriptAttributes(n *html.Node) []scriptAttribute {
	attrs := make([]scriptAttribute, 0, len(n.Attr))
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, "src") {
			continue
		}
		name := attr.Key
		if attr.Namespace != "" {
			name = attr.Namespace + ":" + name
		}
		attrs = append(attrs, scriptAttribute{Name: name, Value: attr.Val})
	}
	return attrs
}

func scriptParent(parent string) string {
	if parent == "head" {
		return "head"
	}
	return "body"
}

func quoteTypeScriptString(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`'`, `\'`,
		"\r", `\r`,
		"\n", `\n`,
		"\u2028", `\u2028`,
		"\u2029", `\u2029`,
	)
	return "'" + replacer.Replace(value) + "'"
}
