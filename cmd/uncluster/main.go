package main

import (
	"encoding/json"
	"fmt"
	"github.com/omariomari2/uncluster/internal/analyzer"
	"github.com/omariomari2/uncluster/internal/bundle"
	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/formatter"
	"github.com/omariomari2/uncluster/internal/nodejs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var validFormats = []string{"split", "nodejs", "nodejs-ejs", "format", "analyze", "bundle"}

func usage() {
	fmt.Fprintf(os.Stderr, `uncluster — process HTML files from the command line

Usage:
  uncluster <file.html> -to <format> [-out <dir>]

Formats:
  split        Extract inline/external CSS and JS into separate files
  nodejs       Scaffold an Express + Vite + TypeScript project
  nodejs-ejs   Scaffold an Express + EJS server-rendered project
  format       Re-indent and normalize HTML (writes to stdout or output dir)
  analyze      Detect repeated UI patterns and suggest components (JSON)
  bundle       Find source index.html from a ZIP/HTML input and write index.html, unzip/, and ejs/

Examples:
  uncluster index.html -to split -out ./output
  uncluster example-site.zip -to bundle -out ./sites
  uncluster page.html -to nodejs -out ./my-project
  uncluster template.html -to format
  uncluster dashboard.html -to analyze

Flags:
  -to string    output format (required)
  -out string   output directory (default: ./<format>-output)
  -dest string  exact final output directory for bundle mode
`)
}

// parseArgs handles flag parsing regardless of argument order.
// Go's flag package stops at the first non-flag arg, so we separate
// flags and positional args ourselves.
func parseArgs() (inputFile, format, outDir, destDir string) {
	args := os.Args[1:]

	var positional []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-to":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "-out":
			if i+1 < len(args) {
				outDir = args[i+1]
				i++
			}
		case "-dest":
			if i+1 < len(args) {
				destDir = args[i+1]
				i++
			}
		case "-h", "-help", "--help":
			usage()
			os.Exit(0)
		default:
			if strings.HasPrefix(args[i], "-") {
				fmt.Fprintf(os.Stderr, "error: unknown flag %q\n", args[i])
				usage()
				os.Exit(2)
			}
			positional = append(positional, args[i])
		}
	}

	if len(positional) < 1 {
		return "", format, outDir, destDir
	}
	return positional[0], format, outDir, destDir
}

func main() {
	inputFile, format, outDir, destDir := parseArgs()

	if inputFile == "" {
		usage()
		os.Exit(2)
	}

	format = strings.ToLower(strings.TrimSpace(format))

	if format == "" {
		fmt.Fprintln(os.Stderr, "error: -to flag is required")
		usage()
		os.Exit(2)
	}

	if !isValidFormat(format) {
		fmt.Fprintf(os.Stderr, "error: unknown format %q — valid formats: %s\n", format, strings.Join(validFormats, ", "))
		os.Exit(2)
	}
	if destDir != "" && format != "bundle" {
		fmt.Fprintln(os.Stderr, "error: -dest is only supported with -to bundle")
		os.Exit(2)
	}

	inputAbs, err := filepath.Abs(inputFile)
	if err != nil {
		fail("resolve input path", err)
	}

	var htmlContent string
	if format != "bundle" {
		raw, err := os.ReadFile(inputAbs)
		if err != nil {
			fail("read input file", err)
		}
		htmlContent = string(raw)
	}

	switch format {
	case "format":
		runFormat(htmlContent, outDir)
	case "analyze":
		runAnalyze(htmlContent, outDir)
	case "split":
		runSplit(htmlContent, inputAbs, resolveOutDir(outDir, "split-output"))
	case "nodejs":
		runNodeJS(htmlContent, filepath.Dir(inputAbs), resolveOutDir(outDir, "nodejs-project"))
	case "nodejs-ejs":
		runNodeJSEJS(htmlContent, filepath.Dir(inputAbs), resolveOutDir(outDir, "nodejs-ejs-project"))
	case "bundle":
		runBundle(inputAbs, resolveOutDir(outDir, "bundle-output"), destDir)
	}
}

func isValidFormat(f string) bool {
	for _, v := range validFormats {
		if f == v {
			return true
		}
	}
	return false
}

func resolveOutDir(explicit, fallback string) string {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			fail("resolve output path", err)
		}
		return abs
	}
	abs, err := filepath.Abs(fallback)
	if err != nil {
		fail("resolve output path", err)
	}
	return abs
}

// --- format ---

func runFormat(htmlContent, outDir string) {
	formatted, err := formatter.Format(htmlContent)
	if err != nil {
		fail("format HTML", err)
	}

	if outDir == "" {
		fmt.Print(formatted)
		return
	}

	dir := resolveOutDir(outDir, "")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fail("create output directory", err)
	}
	outPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(outPath, []byte(formatted), 0o644); err != nil {
		fail("write formatted HTML", err)
	}
	fmt.Printf("Formatted HTML written to %s\n", outPath)
}

// --- analyze ---

func runAnalyze(htmlContent, outDir string) {
	suggestions, err := analyzer.AnalyzeComponents(htmlContent)
	if err != nil {
		fail("analyze components", err)
	}

	data, err := json.MarshalIndent(suggestions, "", "  ")
	if err != nil {
		fail("marshal analysis", err)
	}

	if outDir == "" {
		fmt.Println(string(data))
		return
	}

	dir := resolveOutDir(outDir, "")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fail("create output directory", err)
	}
	outPath := filepath.Join(dir, "components.json")
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		fail("write analysis", err)
	}
	fmt.Printf("Component analysis written to %s\n", outPath)
}

// --- split ---

func runSplit(htmlContent, inputAbs, outDir string) {
	extracted, err := extractor.Extract(htmlContent)
	if err != nil {
		fail("extract resources", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail("create output directory", err)
	}

	if err := writeFile(filepath.Join(outDir, "index.html"), extracted.HTML); err != nil {
		fail("write index.html", err)
	}

	for _, r := range extracted.InlineCSS {
		p := filepath.Join(outDir, filepath.FromSlash(r.Path))
		if err := writeFile(p, r.Content); err != nil {
			fail("write inline CSS", err)
		}
	}

	for _, r := range extracted.InlineJS {
		p := filepath.Join(outDir, filepath.FromSlash(r.Path))
		if err := writeFile(p, r.Content); err != nil {
			fail("write inline JS", err)
		}
	}

	var externalCSS []string
	for _, r := range extracted.ExternalCSS {
		if r.Error != nil || r.Content == "" {
			continue
		}
		p := filepath.Join(outDir, "external", "css", r.Filename)
		if err := writeFile(p, r.Content); err != nil {
			fail("write external CSS", err)
		}
		externalCSS = append(externalCSS, "external/css/"+r.Filename)
	}

	var externalJS []string
	for _, r := range extracted.ExternalJS {
		if r.Error != nil || r.Content == "" {
			continue
		}
		p := filepath.Join(outDir, "external", "js", r.Filename)
		if err := writeFile(p, r.Content); err != nil {
			fail("write external JS", err)
		}
		externalJS = append(externalJS, "external/js/"+r.Filename)
	}

	manifest := map[string]interface{}{
		"created_at":       time.Now().Format(time.RFC3339),
		"input_path":       inputAbs,
		"output_path":      outDir,
		"html_file":        "index.html",
		"inline_css_count": len(extracted.InlineCSS),
		"inline_js_count":  len(extracted.InlineJS),
		"external_css":     externalCSS,
		"external_js":      externalJS,
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fail("marshal manifest", err)
	}
	if err := writeFile(filepath.Join(outDir, "split-manifest.json"), string(manifestData)); err != nil {
		fail("write manifest", err)
	}

	fmt.Printf("Split completed: %s\n", outDir)
}

// --- nodejs ---

func runNodeJS(htmlContent, srcDir, outDir string) {
	htmlContent, localAssets := localizeAssets(htmlContent, srcDir)

	extracted, err := extractor.Extract(htmlContent)
	if err != nil {
		fail("extract resources", err)
	}

	rewrittenHTML := extracted.RewriteForNodeJS()
	projectName := filepath.Base(outDir)

	config := &nodejs.ProjectConfig{
		ProjectName: projectName,
		HTML:        rewrittenHTML,
		CSS:         extracted.CSS,
		JS:          extracted.JS,
		InlineJS:    extracted.InlineJS,
		ExternalCSS: extracted.ExternalCSS,
		ExternalJS:  extracted.ExternalJS,
		LocalAssets: localAssets,
	}

	projectFiles, err := nodejs.GenerateProject(config)
	if err != nil {
		fail("generate Node.js project", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail("create output directory", err)
	}

	for relPath, content := range projectFiles.Files {
		p := filepath.Join(outDir, filepath.FromSlash(relPath))
		if err := writeFile(p, content); err != nil {
			fail("write "+relPath, err)
		}
	}

	for relPath, content := range projectFiles.Binary {
		p := filepath.Join(outDir, filepath.FromSlash(relPath))
		if err := writeBinaryFile(p, content); err != nil {
			fail("write "+relPath, err)
		}
	}

	fmt.Printf("Node.js project generated: %s\n", outDir)
	fmt.Printf("  cd %s && npm install && npm run dev\n", outDir)
}

// --- nodejs-ejs ---

func runNodeJSEJS(htmlContent, srcDir, outDir string) {
	htmlContent, localAssets := localizeAssets(htmlContent, srcDir)

	extracted, err := extractor.Extract(htmlContent)
	if err != nil {
		fail("extract resources", err)
	}

	rewrittenHTML := extracted.RewriteForEJS()
	projectName := filepath.Base(outDir)

	config := &nodejs.EJSProjectConfig{
		ProjectName: projectName,
		HTML:        rewrittenHTML,
		InlineCSS:   extracted.InlineCSS,
		InlineJS:    extracted.InlineJS,
		ExternalCSS: extracted.ExternalCSS,
		ExternalJS:  extracted.ExternalJS,
		LocalAssets: localAssets,
	}

	projectFiles, err := nodejs.GenerateEJSProject(config)
	if err != nil {
		fail("generate EJS project", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fail("create output directory", err)
	}

	for relPath, content := range projectFiles.Files {
		p := filepath.Join(outDir, filepath.FromSlash(relPath))
		if err := writeFile(p, content); err != nil {
			fail("write "+relPath, err)
		}
	}

	for relPath, content := range projectFiles.Binary {
		p := filepath.Join(outDir, filepath.FromSlash(relPath))
		if err := writeBinaryFile(p, content); err != nil {
			fail("write "+relPath, err)
		}
	}

	fmt.Printf("EJS project generated: %s\n", outDir)
	fmt.Printf("  cd %s && npm install && npm start\n", outDir)
}

// --- bundle ---

func runBundle(inputPath, outDir, destDir string) {
	options := bundle.Options{OutputBase: outDir}
	if destDir != "" {
		destAbs, err := filepath.Abs(destDir)
		if err != nil {
			fail("resolve destination path", err)
		}
		options.Destination = destAbs
	}

	result, err := bundle.ProcessWithOptions(inputPath, options)
	if err != nil {
		fail("bundle source", err)
	}

	fmt.Printf("Bundle generated: %s\n", result.OutputDir)
	fmt.Printf("  source: %s\n", result.IndexPath)
	fmt.Printf("  split:  %s\n", result.SplitDir)
	fmt.Printf("  ejs:    %s\n", result.EJSDir)
}

// --- helpers ---

func writeFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "uncluster: %s: %v\n", step, err)
	os.Exit(1)
}

func writeBinaryFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

// localizeAssets rewrites references to files sitting beside the input HTML so
// the generated project carries them, instead of pointing back at the origin
// site. A failure here is not fatal: the project is still generated, just
// without local copies.
func localizeAssets(htmlContent, srcDir string) (string, []extractor.LocalAsset) {
	if srcDir == "" {
		return htmlContent, nil
	}
	rewritten, assets, err := bundle.LocalizeAssets(htmlContent, srcDir)
	if err != nil {
		log.Printf("uncluster: could not localize assets from %s: %v", srcDir, err)
		return htmlContent, nil
	}
	return rewritten, assets
}
