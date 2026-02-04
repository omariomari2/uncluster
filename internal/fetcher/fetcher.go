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
		resp, err := client.Get(resourceURL)
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

	// Handle duplicates by adding a counter
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

	// Remove common CDN prefixes and make hostname more readable
	hostname = strings.ReplaceAll(hostname, "cdn.jsdelivr.net", "jsdelivr")
	hostname = strings.ReplaceAll(hostname, "cdnjs.cloudflare.com", "cloudflare")
	hostname = strings.ReplaceAll(hostname, "code.jquery.com", "jquery")
	hostname = strings.ReplaceAll(hostname, "fonts.googleapis.com", "google-fonts")
	hostname = strings.ReplaceAll(hostname, "unpkg.com", "unpkg")
	hostname = strings.ReplaceAll(hostname, "stackpath.bootstrapcdn.com", "bootstrap")
	hostname = strings.ReplaceAll(hostname, "maxcdn.bootstrapcdn.com", "bootstrap")

	// Clean up the hostname
	hostname = strings.ReplaceAll(hostname, ".", "-")
	hostname = strings.ReplaceAll(hostname, "www-", "")

	// Extract meaningful parts from the path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	var meaningfulParts []string

	for _, part := range pathParts {
		// Skip version numbers and common meaningless parts
		if part == "" || part == "dist" || part == "min" ||
			strings.HasPrefix(part, "v") ||
			strings.Contains(part, "@") ||
			part == "css" || part == "js" {
			continue
		}

		// Keep meaningful parts
		if len(part) > 0 && !isVersionNumber(part) {
			meaningfulParts = append(meaningfulParts, part)
		}
	}

	// Build the filename
	var filename string
	if len(meaningfulParts) > 0 {
		// Use meaningful parts from path
		filename = strings.Join(meaningfulParts, "-")
	} else {
		// Fallback to hostname
		filename = hostname
	}

	// Ensure we have a meaningful name
	if filename == "" || len(filename) < 2 {
		filename = "external"
	}

	// Add resource type prefix for clarity
	switch resourceType {
	case "css":
		filename = "style-" + filename
	case "js":
		filename = "script-" + filename
	}

	// Ensure we have an extension
	if !strings.Contains(filename, ".") {
		filename += getExtension(resourceType)
	}

	return filename
}

func isVersionNumber(s string) bool {
	// Check for common version patterns
	if strings.HasPrefix(s, "v") && len(s) > 1 {
		// v1.2.3, v2.0, etc.
		return true
	}

	// Check for semantic versioning patterns
	if strings.Count(s, ".") >= 1 {
		// 1.2.3, 2.0.0, etc.
		return true
	}

	// Check for single digit versions
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
	// Replace unsafe characters with underscores
	unsafeChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}

	for _, char := range unsafeChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Remove multiple consecutive underscores
	for strings.Contains(filename, "__") {
		filename = strings.ReplaceAll(filename, "__", "_")
	}

	// Trim underscores from start and end
	filename = strings.Trim(filename, "_")

	// Ensure filename is not empty
	if filename == "" {
		filename = "resource"
	}

	// Limit filename length
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
