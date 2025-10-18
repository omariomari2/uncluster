package nodejs

import (
	"fmt"
	"htmlfmt/internal/converter"
	"htmlfmt/internal/fetcher"
	"log"
	"strings"
	"text/template"
)

// ProjectConfig represents the configuration for generating a Node.js project
type ProjectConfig struct {
	ProjectName    string
	PackageManager string // "npm"
	HTML           string
	CSS            string
	JS             string
	ExternalCSS    []fetcher.FetchedResource
	ExternalJS     []fetcher.FetchedResource
}

// ProjectFiles represents the generated project files
type ProjectFiles struct {
	Files map[string]string // filename -> content
}

// GenerateProject creates a complete Node.js project from the given configuration
func GenerateProject(config *ProjectConfig) (*ProjectFiles, error) {
	log.Printf("🏗️ Generating Node.js project: %s", config.ProjectName)

	files := make(map[string]string)

	// Generate configuration files
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

	// Generate README
	readme, err := generateREADME(config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate README: %w", err)
	}
	files["README.md"] = readme

	// Organize source files
	organizeSourceFiles(config, files)

	log.Printf("✅ Generated %d files for Node.js project", len(files))

	return &ProjectFiles{Files: files}, nil
}

// generatePackageJSON creates the package.json file
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

// generateREADME creates the README.md file
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

// organizeSourceFiles organizes the HTML, CSS, and JS files into the proper React/TypeScript structure
func organizeSourceFiles(config *ProjectConfig, files map[string]string) {
	// Add the main HTML file (for Vite)
	files["src/index.html"] = indexHtmlTemplate

	// Convert HTML to JSX and create main component
	mainComponent, err := converter.ConvertToJSX(config.HTML, config.CSS, config.JS, config.ExternalCSS, config.ExternalJS)
	if err != nil {
		log.Printf("⚠️ Failed to convert HTML to JSX: %v", err)
		// Fallback to basic JSX
		mainComponent = fmt.Sprintf(`import React from 'react'

function MainComponent() {
  return (
    <div dangerouslySetInnerHTML={{__html: %q}} />
  )
}

export default MainComponent
`, config.HTML)
	}
	files["src/components/MainComponent.tsx"] = mainComponent

	// Add App.tsx
	files["src/App.tsx"] = appTsxTemplate

	// Add main.tsx
	files["src/main.tsx"] = mainTsxTemplate

	// Add inline CSS if present
	if config.CSS != "" {
		files["src/styles/main.css"] = config.CSS
	}

	// Add external CSS files
	for _, css := range config.ExternalCSS {
		if css.Error == nil && css.Content != "" {
			files["src/styles/external/"+css.Filename] = css.Content
		}
	}

	// Add external JS files (as modules)
	for _, js := range config.ExternalJS {
		if js.Error == nil && js.Content != "" {
			files["src/scripts/external/"+js.Filename] = js.Content
		}
	}

	// Try to create additional components from HTML analysis
	components, err := converter.AnalyzeAndConvert(config.HTML)
	if err == nil {
		for i, component := range components {
			filename := fmt.Sprintf("src/components/Component%d.tsx", i+1)
			files[filename] = component
		}
	}
}
