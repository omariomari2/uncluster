package bundle

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSelectIndexHTMLPrefersFolderMatchingZipName(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, filepath.Join("noise", "index.html"), "<html><body>noise</body></html>")
	writeTestFile(t, root, filepath.Join("example-site", "index.html"), "<html><body>match</body></html>")

	candidate, err := selectIndexHTML(root, "example-site")
	if err != nil {
		t.Fatalf("selectIndexHTML returned error: %v", err)
	}

	if candidate.relPath != "example-site/index.html" {
		t.Fatalf("expected matching folder index, got %q", candidate.relPath)
	}
}

func TestSelectIndexHTMLFallsBackToLargestValidHTML(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, root, filepath.Join("a", "index.html"), "<html><body>small</body></html>")
	writeTestFile(t, root, filepath.Join("b", "index.html"), "<html><body>"+strings.Repeat("large", 100)+"</body></html>")

	candidate, err := selectIndexHTML(root, "different-site")
	if err != nil {
		t.Fatalf("selectIndexHTML returned error: %v", err)
	}

	if candidate.relPath != "b/index.html" {
		t.Fatalf("expected largest valid index, got %q", candidate.relPath)
	}
}

func TestProcessZipWritesExpectedBundleShapeAndReferencedAssetsOnly(t *testing.T) {
	workDir := t.TempDir()
	zipPath := filepath.Join(workDir, "example.com.zip")
	createTestZip(t, zipPath, map[string]string{
		"example.com/index.html":        `<!doctype html><html><head><link rel="stylesheet" href="css/site.css"></head><body><img src="images/logo.png"></body></html>`,
		"example.com/css/site.css":      `body{background:url("../images/bg.png")}`,
		"example.com/images/logo.png":   "logo",
		"example.com/images/bg.png":     "bg",
		"example.com/images/unused.png": "unused",
	})

	outDir := filepath.Join(workDir, "out")
	if err := os.MkdirAll(filepath.Join(outDir, "example.com", "zips"), 0o755); err != nil {
		t.Fatalf("create stale zips dir: %v", err)
	}

	result, err := Process(zipPath, outDir)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}

	expectFile(t, filepath.Join(result.OutputDir, "index.html"))
	expectFile(t, filepath.Join(result.OutputDir, "unzip", "index.html"))
	expectFile(t, filepath.Join(result.OutputDir, "unzip", "assets", "images", "logo.png"))
	expectFile(t, filepath.Join(result.OutputDir, "unzip", "assets", "images", "bg.png"))
	expectFile(t, filepath.Join(result.OutputDir, "ejs", "views", "index.ejs"))
	expectFile(t, filepath.Join(result.OutputDir, "ejs", "public", "assets", "images", "logo.png"))
	expectFile(t, filepath.Join(result.OutputDir, "ejs", "public", "assets", "images", "bg.png"))

	if _, err := os.Stat(filepath.Join(result.OutputDir, "zips")); !os.IsNotExist(err) {
		t.Fatalf("did not expect zips directory")
	}
	if _, err := os.Stat(filepath.Join(result.OutputDir, "example.com")); !os.IsNotExist(err) {
		t.Fatalf("original ZIP folder should not be copied to output")
	}
	if _, err := os.Stat(filepath.Join(result.OutputDir, "unzip", "assets", "images", "unused.png")); !os.IsNotExist(err) {
		t.Fatalf("unused asset should not be copied to split output")
	}
	if _, err := os.Stat(filepath.Join(result.OutputDir, "ejs", "public", "assets", "images", "unused.png")); !os.IsNotExist(err) {
		t.Fatalf("unused asset should not be copied to EJS output")
	}
}

func TestProcessWithOptionsWritesExactDestination(t *testing.T) {
	workDir := t.TempDir()
	htmlPath := filepath.Join(workDir, "index.html")
	writeTestFile(t, workDir, "index.html", `<!doctype html><html><body><img src="logo.png"></body></html>`)
	writeTestFile(t, workDir, "logo.png", "logo")

	destDir := filepath.Join(workDir, "chosen-output")
	result, err := ProcessWithOptions(htmlPath, Options{Destination: destDir})
	if err != nil {
		t.Fatalf("ProcessWithOptions returned error: %v", err)
	}

	if result.OutputDir != destDir {
		t.Fatalf("expected exact destination %q, got %q", destDir, result.OutputDir)
	}
	expectFile(t, filepath.Join(destDir, "index.html"))
	expectFile(t, filepath.Join(destDir, "unzip", "index.html"))
	expectFile(t, filepath.Join(destDir, "ejs", "views", "index.ejs"))
}

func writeTestFile(t *testing.T, root, relPath, content string) {
	t.Helper()
	path := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func createTestZip(t *testing.T, zipPath string, files map[string]string) {
	t.Helper()
	out, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	writer := zip.NewWriter(out)
	for relPath, content := range files {
		entry, err := writer.Create(relPath)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", relPath, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("write zip entry %s: %v", relPath, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}
}

func expectFile(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("expected file %s, got directory", path)
	}
}
