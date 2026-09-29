package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/omariomari2/uncluster/internal/extractor"
)

func TestExportArchivesIncludeCaptureManifestWithoutChangingResponseContract(t *testing.T) {
	app := fiber.New(fiber.Config{BodyLimit: 50 * 1024 * 1024})
	setupRoutes(app)
	body := `{"html":"<!doctype html><html><body><h1>Captured</h1></body></html>"}`

	for _, route := range []string{"/api/export", "/api/export-nodejs", "/api/export-nodejs-ejs"} {
		t.Run(route, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, route, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, body = %s", resp.StatusCode, data)
			}
			if got := resp.Header.Get("Content-Type"); got != "application/zip" {
				t.Fatalf("Content-Type = %q", got)
			}
			if got := resp.Header.Get("Content-Disposition"); !strings.HasPrefix(got, "attachment; filename=") {
				t.Fatalf("Content-Disposition = %q", got)
			}

			manifestData := captureManifestFromArchive(t, data)
			var manifest extractor.CaptureManifest
			if err := json.Unmarshal(manifestData, &manifest); err != nil {
				t.Fatalf("decode uncluster-capture.json: %v", err)
			}
			if manifest.Version != extractor.CaptureManifestVersion || !manifest.Complete || len(manifest.Assets) != 0 {
				t.Fatalf("unexpected capture manifest: %+v", manifest)
			}
		})
	}
}

func TestExportNodeJSPassesOrderedInlineScriptsToProject(t *testing.T) {
	app := fiber.New(fiber.Config{BodyLimit: 50 * 1024 * 1024})
	setupRoutes(app)
	body := `{"html":"<!doctype html><html><body><script>window.first = 1</script><script>window.second = 2</script></body></html>"}`
	req := httptest.NewRequest(http.MethodPost, "/api/export-nodejs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, data)
	}

	entries := textEntriesFromArchive(t, data)
	var first, second string
	for name, content := range entries {
		switch {
		case strings.HasSuffix(name, "/public/scripts/inline/script-1.js"):
			first = content
		case strings.HasSuffix(name, "/public/scripts/inline/script-2.js"):
			second = content
		}
	}
	if first != "window.first = 1" || second != "window.second = 2" {
		t.Fatalf("inline scripts = %q, %q", first, second)
	}
}

func captureManifestFromArchive(t *testing.T, data []byte) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open ZIP: %v", err)
	}
	for _, file := range zr.File {
		if file.Name != "uncluster-capture.json" && !strings.HasSuffix(file.Name, "/uncluster-capture.json") {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		return contents
	}
	t.Fatal("ZIP is missing uncluster-capture.json")
	return nil
}

func textEntriesFromArchive(t *testing.T, data []byte) map[string]string {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("open ZIP: %v", err)
	}
	entries := make(map[string]string, len(zr.File))
	for _, file := range zr.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		contents, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries[file.Name] = string(contents)
	}
	return entries
}
