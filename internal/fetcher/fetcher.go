package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
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
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read body: %w", err)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	// Strip parameters like charset
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}

	return data, ct, nil
}

func FetchExternalResources(urls []string, resourceType string) []FetchedResource {
	if len(urls) == 0 {
		return []FetchedResource{}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	var results []FetchedResource
	usedFilenames := make(map[string]int)

	for _, resourceURL := range urls {
		req, reqErr := http.NewRequest("GET", resourceURL, nil)
		if reqErr != nil {
			results = append(results, FetchedResource{
				URL:   resourceURL,
				Type:  resourceType,
				Error: reqErr,
			})
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		resp, err := client.Do(req)
		if err != nil {
			results = append(results, FetchedResource{
				URL:   resourceURL,
				Type:  resourceType,
				Error: err,
			})
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err := fmt.Errorf("HTTP %d", resp.StatusCode)
			results = append(results, FetchedResource{
				URL:   resourceURL,
				Type:  resourceType,
				Error: err,
			})
			continue
		}

		content, err := io.ReadAll(resp.Body)
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

func generateSafeFilename(resourceURL, resourceType string, usedFilenames map[string]int) string {
	parsedURL, err := url.Parse(resourceURL)
	if err != nil {
		return fmt.Sprintf("external-%d.%s", len(usedFilenames), getExtension(resourceType))
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

	for _, part := range pathParts {
		if part == "" || part == "dist" || part == "min" ||
			strings.HasPrefix(part, "v") ||
			strings.Contains(part, "@") ||
			part == "css" || part == "js" {
			continue
		}

		if len(part) > 0 && !isVersionNumber(part) {
			meaningfulParts = append(meaningfulParts, part)
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

	if !strings.Contains(filename, ".") {
		filename += getExtension(resourceType)
	}

	return filename
}

func isVersionNumber(s string) bool {
	if strings.HasPrefix(s, "v") && len(s) > 1 {
		return true
	}

	if strings.Count(s, ".") >= 1 {
		return true
	}

	if len(s) == 1 && s >= "0" && s <= "9" {
		return true
	}

	return false
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
