package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/omariomari2/uncluster/internal/extractor"
	"os"
	"path/filepath"
	"time"
)

type outputManifest struct {
	CreatedAt      string   `json:"created_at"`
	InputPath      string   `json:"input_path"`
	OutputPath     string   `json:"output_path"`
	HTMLFile       string   `json:"html_file"`
	InlineCSSCount int      `json:"inline_css_count"`
	InlineJSCount  int      `json:"inline_js_count"`
	ExternalCSS    []string `json:"external_css"`
	ExternalJS     []string `json:"external_js"`
}

func main() {
	inputPath := flag.String("input", "", "path to clustered HTML input file")
	outputPath := flag.String("output", "", "directory to write split output")
	writeManifest := flag.Bool("manifest", true, "write split-manifest.json in output directory")
	flag.Parse()

	if *inputPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/uncluster-split -input <file.html> -output <dir>")
		os.Exit(2)
	}

	inputAbs, err := filepath.Abs(*inputPath)
	if err != nil {
		fail("resolve input path", err)
	}
	outputAbs, err := filepath.Abs(*outputPath)
	if err != nil {
		fail("resolve output path", err)
	}

	raw, err := os.ReadFile(inputAbs)
	if err != nil {
		fail("read input file", err)
	}

	extracted, err := extractor.Extract(string(raw))
	if err != nil {
		fail("extract resources", err)
	}

	if err := os.MkdirAll(outputAbs, 0o755); err != nil {
		fail("create output directory", err)
	}

	indexPath := filepath.Join(outputAbs, "index.html")
	if err := writeFile(indexPath, extracted.HTML); err != nil {
		fail("write index.html", err)
	}

	for _, resource := range extracted.InlineCSS {
		path := filepath.Join(outputAbs, filepath.FromSlash(resource.Path))
		if err := writeFile(path, resource.Content); err != nil {
			fail("write inline css", err)
		}
	}

	for _, resource := range extracted.InlineJS {
		path := filepath.Join(outputAbs, filepath.FromSlash(resource.Path))
		if err := writeFile(path, resource.Content); err != nil {
			fail("write inline js", err)
		}
	}

	var externalCSS []string
	for _, resource := range extracted.ExternalCSS {
		if resource.Error != nil || resource.Content == "" {
			continue
		}
		path := filepath.Join(outputAbs, "external", "css", resource.Filename)
		if err := writeFile(path, resource.Content); err != nil {
			fail("write external css", err)
		}
		externalCSS = append(externalCSS, filepath.ToSlash(filepath.Join("external", "css", resource.Filename)))
	}

	var externalJS []string
	for _, resource := range extracted.ExternalJS {
		if resource.Error != nil || resource.Content == "" {
			continue
		}
		path := filepath.Join(outputAbs, "external", "js", resource.Filename)
		if err := writeFile(path, resource.Content); err != nil {
			fail("write external js", err)
		}
		externalJS = append(externalJS, filepath.ToSlash(filepath.Join("external", "js", resource.Filename)))
	}

	if *writeManifest {
		manifest := outputManifest{
			CreatedAt:      time.Now().Format(time.RFC3339),
			InputPath:      inputAbs,
			OutputPath:     outputAbs,
			HTMLFile:       "index.html",
			InlineCSSCount: len(extracted.InlineCSS),
			InlineJSCount:  len(extracted.InlineJS),
			ExternalCSS:    externalCSS,
			ExternalJS:     externalJS,
		}
		manifestData, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			fail("marshal manifest", err)
		}
		if err := writeFile(filepath.Join(outputAbs, "split-manifest.json"), string(manifestData)); err != nil {
			fail("write split-manifest.json", err)
		}
	}

	fmt.Printf("Split completed: %s\n", outputAbs)
}

func writeFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func fail(step string, err error) {
	fmt.Fprintf(os.Stderr, "uncluster-split: %s: %v\n", step, err)
	os.Exit(1)
}
