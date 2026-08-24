package nodejs

import (
	"fmt"
	"github.com/omariomari2/uncluster/internal/converter"
	"github.com/omariomari2/uncluster/internal/fetcher"
	"github.com/omariomari2/uncluster/internal/htmlutil"
	"log"
	"strings"

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
		return map[string]string{}, mc, generateMainTsx(inlineCSS, inlineJS, externalCSS, externalJS), nil
	}

	root := selectComponentRoot(body)
	sections := filterComponentCandidates(contentChildren(root))

	if len(sections) == 0 {
		mc, convErr := converter.ConvertSectionToTSX(htmlContent, "MainComponent")
		if convErr != nil {
			return nil, "", "", convErr
		}
		return map[string]string{}, mc, generateMainTsx(inlineCSS, inlineJS, externalCSS, externalJS), nil
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
		return map[string]string{}, mc, generateMainTsx(inlineCSS, inlineJS, externalCSS, externalJS), nil
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

	return sectionFiles, generateMainComponentTSX(resolved), generateMainTsx(inlineCSS, inlineJS, externalCSS, externalJS), nil
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

	return fmt.Sprintf(`import React from 'react'
%s
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
	inlineCSS string,
	inlineJS string,
	externalCSS []fetcher.FetchedResource,
	externalJS []fetcher.FetchedResource,
) string {
	var cssImports strings.Builder
	if strings.TrimSpace(inlineCSS) != "" {
		cssImports.WriteString("import './styles/main.css'\n")
	}
	for _, res := range externalCSS {
		if res.Error == nil && strings.TrimSpace(res.Content) != "" {
			cssImports.WriteString(fmt.Sprintf("import './styles/external/%s'\n", res.Filename))
		}
	}

	var scriptPaths []string
	if strings.TrimSpace(inlineJS) != "" {
		scriptPaths = append(scriptPaths, "/scripts/main.js")
	}
	for _, res := range externalJS {
		if res.Error == nil && strings.TrimSpace(res.Content) != "" {
			scriptPaths = append(scriptPaths, "/scripts/external/"+res.Filename)
		}
	}

	var scriptLoader strings.Builder
	if len(scriptPaths) > 0 {
		scriptLoader.WriteString("\nconst originalScripts = [\n")
		for _, scriptPath := range scriptPaths {
			scriptLoader.WriteString(fmt.Sprintf("  '%s',\n", scriptPath))
		}
		scriptLoader.WriteString(`]

function loadOriginalScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = src
    script.onload = () => resolve()
    script.onerror = () => reject(new Error(` + "`" + `Failed to load ${src}` + "`" + `))
    document.body.appendChild(script)
  })
}

async function loadOriginalScripts() {
  for (const src of originalScripts) {
    await loadOriginalScript(src)
  }
}
`)
	}

	return fmt.Sprintf(`import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
%s
%s
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
%s`, cssImports.String(), scriptLoader.String(), func() string {
		if len(scriptPaths) == 0 {
			return ""
		}
		return "\nvoid loadOriginalScripts()\n"
	}())
}
