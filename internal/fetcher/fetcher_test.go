package fetcher

import (
	"net/url"
	"testing"
)

func TestGenerateDescriptiveFilenameKeepsTheRealAssetName(t *testing.T) {
	tests := []struct {
		name         string
		rawURL       string
		resourceType string
		want         string
	}{
		{
			name:         "jsdelivr versioned package",
			rawURL:       "https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css",
			resourceType: "css",
			want:         "style-bootstrap.css",
		},
		{
			name:         "cdnjs nested library path",
			rawURL:       "https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css",
			resourceType: "css",
			want:         "style-font-awesome-all.css",
		},
		{
			name:         "jquery script",
			rawURL:       "https://code.jquery.com/jquery-3.6.0.min.js",
			resourceType: "js",
			want:         "script-jquery-3.6.0.js",
		},
		{
			name:         "segment starting with v is not a version",
			rawURL:       "https://example.com/assets/vendor/widget.js",
			resourceType: "js",
			want:         "script-vendor-widget.js",
		},
		{
			name:         "no usable path falls back to hostname",
			rawURL:       "https://example.com/",
			resourceType: "css",
			want:         "style-example-com.css",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("parse fixture URL: %v", err)
			}
			got := generateDescriptiveFilename(parsed, tt.resourceType)
			if got != tt.want {
				t.Errorf("generateDescriptiveFilename(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}

func TestIsVersionNumber(t *testing.T) {
	versions := []string{"1", "5.3.0", "v5", "V2.1", "0.0.1"}
	for _, s := range versions {
		if !isVersionNumber(s) {
			t.Errorf("expected %q to be treated as a version", s)
		}
	}

	notVersions := []string{"", "npm", "vue", "vendor", "video", "bootstrap", "all.min.css", "font-awesome"}
	for _, s := range notVersions {
		if isVersionNumber(s) {
			t.Errorf("expected %q not to be treated as a version", s)
		}
	}
}

func TestGenerateSafeFilenameResolvesCollisions(t *testing.T) {
	used := make(map[string]int)

	first := generateSafeFilename("https://a.example.com/site.css", "css", used)
	used[first]++
	second := generateSafeFilename("https://a.example.com/site.css", "css", used)

	if first == second {
		t.Fatalf("expected a distinct filename on collision, got %q twice", first)
	}
	if second != "style-site-1.css" {
		t.Errorf("second filename = %q, want %q", second, "style-site-1.css")
	}
}

func TestFetchExternalResourcesRejectsNonPublicTargets(t *testing.T) {
	results := FetchExternalResources([]string{
		"http://127.0.0.1/secret.css",
		"http://169.254.169.254/latest/meta-data/",
		"file:///etc/passwd",
	}, "css")

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error == nil {
			t.Errorf("expected %q to fail, got content %q", r.URL, r.Content)
		}
	}
}
