package nodejs

import (
	"fmt"
	"github.com/omariomari2/uncluster/internal/fetcher"
	"log"
	"strings"
	"text/template"
)

type ProjectConfig struct {
	ProjectName    string
	PackageManager string
	HTML           string
	CSS            string
	JS             string
	ExternalCSS    []fetcher.FetchedResource
	ExternalJS     []fetcher.FetchedResource
}

type ProjectFiles struct {
	Files map[string]string
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

	log.Printf("✅ Generated %d files for Node.js project", len(files))

	return &ProjectFiles{Files: files}, nil
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
	var buf strings.Builder
	err = tmpl.Execute(&buf, config)
	result := buf.String()
	return result, err
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
		config.ExternalCSS,
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

	for _, css := range config.ExternalCSS {
		if css.Error == nil && css.Content != "" {
			files["src/styles/external/"+css.Filename] = css.Content
		}
	}

	for _, js := range config.ExternalJS {
		if js.Error == nil && js.Content != "" {
			files["src/scripts/external/"+js.Filename] = js.Content
		}
	}
}
