package nodejs

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"
	"github.com/omariomari2/uncluster/internal/htmlutil"

	"golang.org/x/net/html"
)

type ProjectConfig struct {
	ProjectName string
	HTML        string
	CSS         string
	JS          string
	ExternalCSS []fetcher.FetchedResource
	ExternalJS  []fetcher.FetchedResource
	LocalAssets []extractor.LocalAsset
}

type ProjectFiles struct {
	Files map[string]string

	// Binary holds files that are not text, keyed the same way as Files.
	Binary map[string][]byte
}

func GenerateProject(config *ProjectConfig) (*ProjectFiles, error) {
	log.Printf("🏗️ Generating Node.js project: %s", config.ProjectName)

	files := make(map[string]string)

	packageJSON, err := generatePackageJSON(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate package.json: %w", err)
	}
	files["package.json"] = packageJSON

	files["vite.config.js"] = viteConfigTemplate
	files["server.js"] = serverJSTemplate
	files[".eslintrc.json"] = eslintConfigTemplate
	files[".prettierrc"] = prettierConfigTemplate
	files["tsconfig.json"] = tsconfigTemplate
	files[".gitignore"] = gitignoreTemplate

	readme, err := generateREADME(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate README: %w", err)
	}
	files["README.md"] = readme

	organizeSourceFiles(config, files)

	binary := localAssetFiles(config.LocalAssets, "public")

	log.Printf("✅ Generated %d files for Node.js project", len(files)+len(binary))

	return &ProjectFiles{Files: files, Binary: binary}, nil
}

func generatePackageJSON(config *ProjectConfig) (string, error) {
	tmpl, err := template.New("package.json").Parse(packageJSONTemplate)
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

func generateREADME(config *ProjectConfig) (string, error) {
	tmpl, err := template.New("README.md").Parse(readmeTemplate)
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

func generateIndexHTML(config *ProjectConfig) (string, error) {
	tmpl, err := template.New("index.html").Parse(indexHtmlTemplate)
	if err != nil {
		return "", err
	}

	title, headMeta := documentHead(config.HTML)
	if strings.TrimSpace(title) == "" {
		title = config.ProjectName
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, struct {
		ProjectName string
		Title       string
		HeadMeta    string
	}{
		ProjectName: config.ProjectName,
		Title:       title,
		HeadMeta:    headMeta,
	})
	return buf.String(), err
}

// documentHead returns the source page's title and the <head> elements worth
// carrying into the generated shell. The body becomes components, so without
// this the page loses its identity: title, description, Open Graph tags,
// canonical URL and icons all live in <head>.
//
// Stylesheets and scripts are excluded because main.tsx already imports the
// former and loads the latter; re-emitting them here would double-load them.
func documentHead(htmlContent string) (title string, meta string) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", ""
	}
	head := htmlutil.FindElement(doc, "head")
	if head == nil {
		return "", ""
	}

	var b strings.Builder
	for child := head.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		switch child.Data {
		case "title":
			title = strings.TrimSpace(htmlutil.TextContent(child))
		case "meta":
			if htmlutil.HasAttr(child, "charset") {
				continue
			}
			if strings.EqualFold(htmlutil.GetAttr(child, "name"), "viewport") {
				continue
			}
			writeHeadElement(&b, child)
		case "link":
			if strings.Contains(strings.ToLower(htmlutil.GetAttr(child, "rel")), "stylesheet") {
				continue
			}
			writeHeadElement(&b, child)
		}
	}
	return title, b.String()
}

func writeHeadElement(b *strings.Builder, n *html.Node) {
	var rendered bytes.Buffer
	if err := html.Render(&rendered, n); err != nil {
		return
	}
	b.WriteString("    " + rendered.String() + "\n")
}

func organizeSourceFiles(config *ProjectConfig, files map[string]string) {
	indexHTML, err := generateIndexHTML(config)
	if err != nil {
		log.Printf("⚠️ Failed to generate index.html: %v", err)
		indexHTML = indexHtmlTemplate
	}
	files["src/index.html"] = indexHTML

	sectionFiles, mainComponent, mainTsx, err := generateTSXViews(
		config.HTML,
		config.CSS,
		config.JS,
		config.ExternalCSS,
		config.ExternalJS,
	)
	if err != nil {
		log.Printf("⚠️ Failed to generate TSX views: %v", err)
		mainComponent = fmt.Sprintf(`import React from 'react'

function MainComponent() {
  return (
    <div dangerouslySetInnerHTML={{__html: %q}} />
  )
}

export default MainComponent
`, config.HTML)
		mainTsx = mainTsxFallback
	}

	for filename, content := range sectionFiles {
		files[filename] = content
	}
	files["src/components/MainComponent.tsx"] = mainComponent
	files["src/App.tsx"] = appTsxTemplate
	files["src/main.tsx"] = mainTsx

	if config.CSS != "" {
		files["src/styles/main.css"] = config.CSS
	}
	if strings.TrimSpace(config.JS) != "" {
		files["public/scripts/main.js"] = config.JS
	}

	for _, css := range config.ExternalCSS {
		if css.Error == nil && css.Content != "" {
			files["src/styles/external/"+css.Filename] = cssForPublicAssets(css.Content)
		}
	}

	for _, js := range config.ExternalJS {
		if js.Error == nil && js.Content != "" {
			files["public/scripts/external/"+js.Filename] = js.Content
		}
	}
}

func cssForPublicAssets(content string) string {
	return strings.ReplaceAll(content, "../../assets/", "/assets/")
}

// localAssetFiles places a page's referenced images, fonts and other binaries
// under the project's public directory, keeping the relative layout the HTML
// already points at.
func localAssetFiles(assets []extractor.LocalAsset, publicDir string) map[string][]byte {
	if len(assets) == 0 {
		return nil
	}
	out := make(map[string][]byte, len(assets))
	for _, asset := range assets {
		if len(asset.Content) == 0 {
			continue
		}
		out[publicDir+"/"+asset.Path] = asset.Content
	}
	return out
}
