package nodejs

import (
	"encoding/json"
	"fmt"
	"htmlfmt/internal/converter"
	"htmlfmt/internal/fetcher"
	"log"
	"os"
	"strings"
	"text/template"
	"time"
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

// generateIndexHTML creates the index.html file with template execution
func generateIndexHTML(config *ProjectConfig) (string, error) {
	// #region agent log
	logDebug("generateIndexHTML", "entry", map[string]interface{}{"projectName": config.ProjectName}, "A")
	// #endregion
	tmpl, err := template.New("index.html").Parse(indexHtmlTemplate)
	if err != nil {
		// #region agent log
		logDebug("generateIndexHTML", "parse_error", map[string]interface{}{"error": err.Error()}, "A")
		// #endregion
		return "", err
	}
	var buf strings.Builder
	err = tmpl.Execute(&buf, config)
	result := buf.String()
	// #region agent log
	hasPlaceholder := strings.Contains(result, "{{.ProjectName}}")
	logDebug("generateIndexHTML", "result", map[string]interface{}{
		"resultLength":   len(result),
		"hasPlaceholder": hasPlaceholder,
		"first100Chars":  safeSubstring(result, 0, 100),
		"error": func() string {
			if err != nil {
				return err.Error()
			} else {
				return ""
			}
		}(),
	}, "A")
	// #endregion
	return result, err
}

// organizeSourceFiles organizes the HTML, CSS, and JS files into the proper React/TypeScript structure
func organizeSourceFiles(config *ProjectConfig, files map[string]string) {
	// Add the main HTML file (for Vite) with template execution
	indexHTML, err := generateIndexHTML(config)
	if err != nil {
		log.Printf("⚠️ Failed to generate index.html: %v", err)
		indexHTML = indexHtmlTemplate
	}
	files["src/index.html"] = indexHTML

	// Convert HTML to JSX and create main component
	// #region agent log
	logDebug("organizeSourceFiles", "before_jsx_conversion", map[string]interface{}{
		"htmlLength":       len(config.HTML),
		"hasCSS":           config.CSS != "",
		"hasJS":            config.JS != "",
		"externalCSSCount": len(config.ExternalCSS),
		"externalJSCount":  len(config.ExternalJS),
	}, "B")
	// #endregion
	mainComponent, err := converter.ConvertToJSX(config.HTML, config.CSS, config.JS, config.ExternalCSS, config.ExternalJS)
	// #region agent log
	hasInvalidJSX := strings.Contains(mainComponent, "<!--") || strings.Contains(mainComponent, "class=") || strings.Contains(mainComponent, "/ />")
	logDebug("organizeSourceFiles", "after_jsx_conversion", map[string]interface{}{
		"componentLength": len(mainComponent),
		"hasError":        err != nil,
		"error": func() string {
			if err != nil {
				return err.Error()
			} else {
				return ""
			}
		}(),
		"hasInvalidJSX": hasInvalidJSX,
		"first200Chars": safeSubstring(mainComponent, 0, 200),
	}, "B")
	// #endregion
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
	// #region agent log
	logDebug("organizeSourceFiles", "before_component_generation", map[string]interface{}{}, "C")
	// #endregion
	components, err := converter.AnalyzeAndConvert(config.HTML)
	// #region agent log
	componentErrors := 0
	for _, comp := range components {
		if strings.Contains(comp, "class=") || strings.Contains(comp, "=\"{string}\"") {
			componentErrors++
		}
	}
	logDebug("organizeSourceFiles", "after_component_generation", map[string]interface{}{
		"componentCount": len(components),
		"hasError":       err != nil,
		"error": func() string {
			if err != nil {
				return err.Error()
			} else {
				return ""
			}
		}(),
		"componentErrors": componentErrors,
	}, "C")
	// #endregion
	if err == nil {
		for i, component := range components {
			filename := fmt.Sprintf("src/components/Component%d.tsx", i+1)
			files[filename] = component
		}
	}
}

func logDebug(location, message string, data map[string]interface{}, hypothesisId string) {
	logEntry := map[string]interface{}{
		"sessionId":    "debug-session",
		"runId":        "run1",
		"hypothesisId": hypothesisId,
		"location":     "builder.go:" + location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	jsonData, _ := json.Marshal(logEntry)
	logPath := ".cursor/debug.log"
	if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		fmt.Fprintln(f, string(jsonData))
		f.Close()
	}
}

func safeSubstring(s string, start, end int) string {
	if start >= len(s) {
		return ""
	}
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}
