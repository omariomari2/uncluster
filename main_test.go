package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newAPITestApp() *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 50 * 1024 * 1024})
	setupRoutes(app)
	return app
}

func performAPIRequest(t *testing.T, app *fiber.App, method, path, contentType string, body io.Reader) (*http.Response, []byte) {
	t.Helper()

	req := httptest.NewRequest(method, path, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read %s %s response: %v", method, path, err)
	}
	return resp, data
}

func TestSetupRoutesExposesCurrentAPIInventory(t *testing.T) {
	app := newAPITestApp()
	expected := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/format"},
		{http.MethodPost, "/api/analyze"},
		{http.MethodPost, "/api/export"},
		{http.MethodPost, "/api/export-nodejs"},
		{http.MethodPost, "/api/export-nodejs-ejs"},
		{http.MethodPost, "/api/bundle-zip"},
		{http.MethodPost, "/api/scrape"},
		{http.MethodPost, "/api/scrape-nodejs"},
		{http.MethodPost, "/api/scrape-nodejs-ejs"},
		{http.MethodGet, "/api/health"},
	}

	for _, endpoint := range expected {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			resp, _ := performAPIRequest(t, app, endpoint.method, endpoint.path, "application/json", strings.NewReader("{}"))
			if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
				t.Fatalf("registered endpoint returned %d", resp.StatusCode)
			}
		})
	}

	resp, _ := performAPIRequest(t, app, http.MethodPost, "/api/convert", "application/json", strings.NewReader(`{"html":"<p>x</p>"}`))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("removed /api/convert returned %d, want 404", resp.StatusCode)
	}
}

func TestHealthRoute(t *testing.T) {
	resp, data := performAPIRequest(t, newAPITestApp(), http.MethodGet, "/api/health", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, data)
	}

	var payload struct {
		Status   string `json:"status"`
		Service  string `json:"service"`
		Revision string `json:"revision"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Status != "healthy" || payload.Service != "uncluster-api" || payload.Revision == "" {
		t.Fatalf("unexpected health payload: %+v", payload)
	}
}

func TestJSONRoutesRejectMalformedAndBlankInput(t *testing.T) {
	app := newAPITestApp()
	routes := []string{
		"/api/format",
		"/api/analyze",
		"/api/export",
		"/api/export-nodejs",
		"/api/export-nodejs-ejs",
		"/api/scrape",
		"/api/scrape-nodejs",
		"/api/scrape-nodejs-ejs",
	}

	for _, path := range routes {
		t.Run(path+" malformed", func(t *testing.T) {
			resp, data := performAPIRequest(t, app, http.MethodPost, path, "application/json", strings.NewReader("{"))
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", resp.StatusCode, data)
			}
		})
		t.Run(path+" blank", func(t *testing.T) {
			resp, data := performAPIRequest(t, app, http.MethodPost, path, "application/json", strings.NewReader("{}"))
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", resp.StatusCode, data)
			}
		})
	}
}

func TestJSONRoutesRejectWrongFieldTypes(t *testing.T) {
	app := newAPITestApp()
	routes := []struct {
		path  string
		field string
	}{
		{"/api/format", "html"},
		{"/api/analyze", "html"},
		{"/api/export", "html"},
		{"/api/export-nodejs", "html"},
		{"/api/export-nodejs-ejs", "html"},
		{"/api/scrape", "url"},
		{"/api/scrape-nodejs", "url"},
		{"/api/scrape-nodejs-ejs", "url"},
	}

	for _, route := range routes {
		t.Run(route.path, func(t *testing.T) {
			body := `{"` + route.field + `":42}`
			resp, data := performAPIRequest(t, app, http.MethodPost, route.path, "application/json", strings.NewReader(body))
			assertAPIError(t, resp, data, http.StatusBadRequest)
		})
	}
}

func TestAnalyzeRouteHonorsSuggestionThreshold(t *testing.T) {
	app := newAPITestApp()
	tests := []struct {
		name      string
		html      string
		wantCount int
		wantTags  []string
	}{
		{
			name:      "two repeated controls stay below threshold",
			html:      `<button class="btn-primary">One</button><button class="btn-primary">Two</button>`,
			wantCount: 0,
		},
		{
			name:      "three repeated controls produce a suggestion",
			html:      `<button class="btn-primary">One</button><button class="btn-primary">Two</button><button class="btn-primary">Three</button>`,
			wantCount: 1,
			wantTags:  []string{"button"},
		},
		{
			name:      "structural elements are not suggested",
			html:      `<div class="card">One</div><div class="card">Two</div><div class="card">Three</div>`,
			wantCount: 0,
		},
		{
			name:      "pattern keywords require token boundaries",
			html:      `<input class="portable"><input class="portable"><input class="portable">`,
			wantCount: 0,
		},
		{
			name: "multiple patterns are returned without relying on map order",
			html: `<button class="btn-primary">One</button><button class="btn-primary">Two</button><button class="btn-primary">Three</button>` +
				`<input class="form-field"><input class="form-field"><input class="form-field">`,
			wantCount: 2,
			wantTags:  []string{"button", "input"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody, err := json.Marshal(map[string]string{"html": tt.html})
			if err != nil {
				t.Fatal(err)
			}
			resp, data := performAPIRequest(t, app, http.MethodPost, "/api/analyze", "application/json", bytes.NewReader(requestBody))
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status=%d body=%s", resp.StatusCode, data)
			}
			var payload struct {
				Success     bool `json:"success"`
				Suggestions []struct {
					Count   int    `json:"count"`
					TagName string `json:"tagName"`
				} `json:"suggestions"`
			}
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !payload.Success || len(payload.Suggestions) != tt.wantCount {
				t.Fatalf("unexpected payload: %+v", payload)
			}
			seenTags := make(map[string]bool, len(payload.Suggestions))
			for _, suggestion := range payload.Suggestions {
				if suggestion.Count != 3 {
					t.Fatalf("suggestion count=%d, want 3", suggestion.Count)
				}
				seenTags[suggestion.TagName] = true
			}
			for _, tag := range tt.wantTags {
				if !seenTags[tag] {
					t.Fatalf("suggestion tags=%v, want %q", seenTags, tag)
				}
			}
		})
	}
}

func TestScrapeRoutesBlockUnsafeTargets(t *testing.T) {
	app := newAPITestApp()
	routes := []string{"/api/scrape", "/api/scrape-nodejs", "/api/scrape-nodejs-ejs"}
	unsafeTargets := []string{"file:///etc/passwd", "http://127.0.0.1:1"}

	for _, path := range routes {
		for _, target := range unsafeTargets {
			t.Run(path+" "+target, func(t *testing.T) {
				body, err := json.Marshal(map[string]string{"url": target})
				if err != nil {
					t.Fatal(err)
				}
				resp, data := performAPIRequest(t, app, http.MethodPost, path, "application/json", bytes.NewReader(body))
				assertAPIError(t, resp, data, http.StatusInternalServerError)
			})
		}
	}
}

func TestLocalJSONRoutesReturnExpectedPayloadFamilies(t *testing.T) {
	app := newAPITestApp()
	body := `{"html":"<!doctype html><html><body><main><h1>Hello</h1></main></body></html>"}`

	for _, path := range []string{"/api/format", "/api/analyze"} {
		t.Run(path, func(t *testing.T) {
			resp, data := performAPIRequest(t, app, http.MethodPost, path, "application/json", strings.NewReader(body))
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status=%d body=%s", resp.StatusCode, data)
			}
			if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
				t.Fatalf("content-type=%q", resp.Header.Get("Content-Type"))
			}
			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if payload["success"] != true {
				t.Fatalf("unexpected payload: %#v", payload)
			}
		})
	}

	exports := map[string]struct {
		expectedEntries       []string
		dispositionPrefix     string
		dispositionNameSuffix string
	}{
		"/api/export": {
			expectedEntries:       []string{"index.html"},
			dispositionPrefix:     `attachment; filename="extracted.zip"`,
			dispositionNameSuffix: `attachment; filename="extracted.zip"`,
		},
		"/api/export-nodejs": {
			expectedEntries:       []string{"/package.json", "/src/main.tsx"},
			dispositionPrefix:     `attachment; filename="project-`,
			dispositionNameSuffix: `.zip"`,
		},
		"/api/export-nodejs-ejs": {
			expectedEntries:       []string{"/package.json", "/views/index.ejs"},
			dispositionPrefix:     `attachment; filename="project-`,
			dispositionNameSuffix: `-ejs.zip"`,
		},
	}
	for path, expected := range exports {
		t.Run(path, func(t *testing.T) {
			resp, data := performAPIRequest(t, app, http.MethodPost, path, "application/json", strings.NewReader(body))
			entries := assertZipResponse(t, resp, data)
			for _, suffix := range expected.expectedEntries {
				if !hasZipEntrySuffix(entries, suffix) {
					t.Errorf("ZIP entries %v do not contain suffix %q", entries, suffix)
				}
			}
			disposition := resp.Header.Get("Content-Disposition")
			if !strings.HasPrefix(disposition, expected.dispositionPrefix) || !strings.HasSuffix(disposition, expected.dispositionNameSuffix) {
				t.Errorf("content-disposition=%q", disposition)
			}
		})
	}
}

func TestBundleZipRouteValidatesAndProcessesUpload(t *testing.T) {
	app := newAPITestApp()

	missing, data := performAPIRequest(t, app, http.MethodPost, "/api/bundle-zip", "multipart/form-data", nil)
	if missing.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing file status=%d body=%s", missing.StatusCode, data)
	}

	var source bytes.Buffer
	zw := zip.NewWriter(&source)
	entry, err := zw.Create("site/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(entry, "<!doctype html><html><body><h1>Fixture</h1></body></html>"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	var requestBody bytes.Buffer
	mw := multipart.NewWriter(&requestBody)
	part, err := mw.CreateFormFile("file", "fixture.ZIP")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(source.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	resp, responseBody := performAPIRequest(t, app, http.MethodPost, "/api/bundle-zip", mw.FormDataContentType(), &requestBody)
	if entries := assertZipResponse(t, resp, responseBody); len(entries) == 0 {
		t.Fatal("bundle response ZIP is empty")
	}
	if got := resp.Header.Get("Content-Disposition"); got != `attachment; filename="bundle.zip"` {
		t.Fatalf("content-disposition=%q", got)
	}
}

func TestBundleZipRouteRejectsInvalidArchives(t *testing.T) {
	app := newAPITestApp()
	tests := []struct {
		name       string
		filename   string
		archive    []byte
		wantStatus int
	}{
		{
			name:       "non zip filename",
			filename:   "fixture.txt",
			archive:    []byte("not a zip"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "corrupt zip",
			filename:   "fixture.zip",
			archive:    []byte("not a zip"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "archive without html",
			filename:   "fixture.zip",
			archive:    createZipFixture(t, map[string]string{"site/style.css": "body{}"}),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "path traversal entry",
			filename:   "fixture.zip",
			archive:    createZipFixture(t, map[string]string{"../escape.txt": "blocked", "site/index.html": "<h1>Fixture</h1>"}),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contentType, body := createMultipartUpload(t, tt.filename, tt.archive)
			resp, data := performAPIRequest(t, app, http.MethodPost, "/api/bundle-zip", contentType, body)
			assertAPIError(t, resp, data, tt.wantStatus)
		})
	}
}

func assertAPIError(t *testing.T, resp *http.Response, data []byte, wantStatus int) {
	t.Helper()
	if resp.StatusCode != wantStatus {
		t.Fatalf("status=%d want=%d body=%s", resp.StatusCode, wantStatus, data)
	}
	var payload struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Success || strings.TrimSpace(payload.Error) == "" {
		t.Fatalf("unexpected error payload: %+v", payload)
	}
}

func createMultipartUpload(t *testing.T, filename string, data []byte) (string, *bytes.Buffer) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return writer.FormDataContentType(), &body
}

func createZipFixture(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var archive bytes.Buffer
	writer := zip.NewWriter(&archive)
	for name, contents := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, contents); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return archive.Bytes()
}

func assertZipResponse(t *testing.T, resp *http.Response, data []byte) []string {
	t.Helper()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, data)
	}
	if resp.Header.Get("Content-Type") != "application/zip" {
		t.Fatalf("content-type=%q", resp.Header.Get("Content-Type"))
	}
	if !strings.Contains(resp.Header.Get("Content-Disposition"), "attachment; filename=") {
		t.Fatalf("content-disposition=%q", resp.Header.Get("Content-Disposition"))
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("invalid zip response: %v", err)
	}
	entries := make([]string, 0, len(zr.File))
	for _, file := range zr.File {
		entries = append(entries, file.Name)
	}
	return entries
}

func hasZipEntrySuffix(entries []string, suffix string) bool {
	for _, entry := range entries {
		if strings.HasSuffix(entry, suffix) {
			return true
		}
	}
	return false
}
