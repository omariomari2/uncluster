package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/omariomari2/uncluster/internal/safehttp"
)

type FetchedResource struct {
	URL      string
	Content  string
	Filename string
	Type     string
	Error    error
}

// FetchRaw downloads a URL and returns the raw bytes plus the detected MIME type.
// Used for binary assets such as images, fonts, and SVGs.
// A 30-second timeout is used to accommodate slower CDNs.
func FetchRaw(rawURL string) (content []byte, mimeType string, err error) {
	return fetchResource(safehttp.Client(30*time.Second), rawURL)
}

func FetchExternalResources(urls []string, resourceType string) []FetchedResource {
	if len(urls) == 0 {
		return []FetchedResource{}
	}

	client := safehttp.Client(10 * time.Second)

	var results []FetchedResource
	usedFilenames := make(map[string]int)

	for _, resourceURL := range urls {
		content, _, err := fetchResource(client, resourceURL)
		if err != nil {
			results = append(results, FetchedResource{
				URL:   resourceURL,
				Type:  resourceType,
				Error: err,
			})
			continue
		}

		filename := generateSafeFilename(resourceURL, resourceType, usedFilenames)
		usedFilenames[filename]++

		results = append(results, FetchedResource{
			URL:      resourceURL,
			Content:  string(content),
			Filename: filename,
			Type:     resourceType,
			Error:    nil,
		})
	}

	return results
}

func fetchResource(client *http.Client, rawURL string) (content []byte, mimeType string, err error) {
	if err := safehttp.ValidateURL(rawURL); err != nil {
		return nil, "", err
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", safehttp.BrowserUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := safehttp.ReadBody(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read body: %w", err)
	}

	return data, responseMIME(resp), nil
}

func responseMIME(resp *http.Response) string {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return "application/octet-stream"
	}
	if idx := strings.Index(contentType, ";"); idx != -1 {
		return strings.TrimSpace(contentType[:idx])
	}
	return contentType
}

func generateSafeFilename(resourceURL, resourceType string, usedFilenames map[string]int) string {
	parsedURL, err := url.Parse(resourceURL)
	if err != nil {
		return fmt.Sprintf("external-%d%s", len(usedFilenames), getExtension(resourceType))
	}

	filename := generateDescriptiveFilename(parsedURL, resourceType)

	filename = sanitizeFilename(filename)

	originalFilename := filename
	counter := 1
	for usedFilenames[filename] > 0 {
		ext := filepath.Ext(originalFilename)
		base := strings.TrimSuffix(originalFilename, ext)
		filename = fmt.Sprintf("%s-%d%s", base, counter, ext)
		counter++
	}

	return filename
}

var ignoredPathSegments = map[string]bool{
	"dist": true, "min": true, "css": true, "js": true,
	"npm": true, "ajax": true, "libs": true, "assets": true,
}

func generateDescriptiveFilename(parsedURL *url.URL, resourceType string) string {
	hostname := parsedURL.Host
	path := parsedURL.Path

	hostname = strings.ReplaceAll(hostname, "cdn.jsdelivr.net", "jsdelivr")
	hostname = strings.ReplaceAll(hostname, "cdnjs.cloudflare.com", "cloudflare")
	hostname = strings.ReplaceAll(hostname, "code.jquery.com", "jquery")
	hostname = strings.ReplaceAll(hostname, "fonts.googleapis.com", "google-fonts")
	hostname = strings.ReplaceAll(hostname, "unpkg.com", "unpkg")
	hostname = strings.ReplaceAll(hostname, "stackpath.bootstrapcdn.com", "bootstrap")
	hostname = strings.ReplaceAll(hostname, "maxcdn.bootstrapcdn.com", "bootstrap")

	hostname = strings.ReplaceAll(hostname, ".", "-")
	hostname = strings.ReplaceAll(hostname, "www-", "")

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	var meaningfulParts []string

	for i, part := range pathParts {
		if part == "" || ignoredPathSegments[part] || strings.Contains(part, "@") {
			continue
		}

		isFinalSegment := i == len(pathParts)-1
		if !isFinalSegment && isVersionNumber(part) {
			continue
		}

		if base := stripResourceExtension(part); base != "" {
			meaningfulParts = append(meaningfulParts, base)
		}
	}

	var filename string
	if len(meaningfulParts) > 0 {
		filename = strings.Join(meaningfulParts, "-")
	} else {
		filename = hostname
	}

	if filename == "" || len(filename) < 2 {
		filename = "external"
	}

	switch resourceType {
	case "css":
		filename = "style-" + filename
	case "js":
		filename = "script-" + filename
	}

	ext := getExtension(resourceType)
	if !strings.HasSuffix(filename, ext) {
		filename += ext
	}

	return filename
}

func stripResourceExtension(part string) string {
	part = strings.TrimSuffix(part, filepath.Ext(part))
	return strings.TrimSuffix(part, ".min")
}

func isVersionNumber(s string) bool {
	if s == "" {
		return false
	}

	if len(s) > 1 && (s[0] == 'v' || s[0] == 'V') && s[1] >= '0' && s[1] <= '9' {
		return true
	}

	for _, r := range s {
		if (r < '0' || r > '9') && r != '.' {
			return false
		}
	}

	return true
}

func getExtension(resourceType string) string {
	switch resourceType {
	case "css":
		return ".css"
	case "js":
		return ".js"
	default:
		return ".txt"
	}
}

func sanitizeFilename(filename string) string {
	unsafeChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}

	for _, char := range unsafeChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	filename = strings.Trim(filename, "_")

	if filename == "" {
		filename = "resource"
	}

	if len(filename) > 100 {
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		if len(base) > 100-len(ext) {
			base = base[:100-len(ext)]
		}
		filename = base + ext
	}

	return filename
}
