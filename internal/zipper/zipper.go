package zipper

import (
	"archive/zip"
	"bytes"
	"htmlfmt/internal/extractor"
	"htmlfmt/internal/fetcher"
	"io"
	"log"
)

// CreateZipWithMetadata creates a zip archive containing HTML, inline resources, and external resources.
func CreateZipWithMetadata(html string, inlineCSS, inlineJS []extractor.InlineResource, externalCSS, externalJS []fetcher.FetchedResource) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	// Add index.html file
	if html != "" {
		htmlFile, err := writer.Create("index.html")
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(htmlFile, html)
		if err != nil {
			return nil, err
		}
	}

	// Add inline CSS files
	if len(inlineCSS) > 0 {
		log.Printf("Adding %d inline CSS files to zip", len(inlineCSS))
		for _, resource := range inlineCSS {
			if resource.Content == "" {
				continue
			}
			cssFile, err := writer.Create(resource.Path)
			if err != nil {
				log.Printf("Failed to create %s in zip: %v", resource.Path, err)
				continue
			}
			_, err = io.WriteString(cssFile, resource.Content)
			if err != nil {
				log.Printf("Failed to write %s to zip: %v", resource.Path, err)
				continue
			}
			log.Printf("Added inline CSS: %s", resource.Path)
		}
	}

	// Add inline JS files
	if len(inlineJS) > 0 {
		log.Printf("Adding %d inline JS files to zip", len(inlineJS))
		for _, resource := range inlineJS {
			if resource.Content == "" {
				continue
			}
			jsFile, err := writer.Create(resource.Path)
			if err != nil {
				log.Printf("Failed to create %s in zip: %v", resource.Path, err)
				continue
			}
			_, err = io.WriteString(jsFile, resource.Content)
			if err != nil {
				log.Printf("Failed to write %s to zip: %v", resource.Path, err)
				continue
			}
			log.Printf("Added inline JS: %s", resource.Path)
		}
	}

	// Add external CSS files
	if len(externalCSS) > 0 {
		log.Printf("📁 Adding %d external CSS files to zip", len(externalCSS))
		for _, resource := range externalCSS {
			if resource.Error == nil && resource.Content != "" {
				path := "external/css/" + resource.Filename
				cssFile, err := writer.Create(path)
				if err != nil {
					log.Printf("⚠️ Failed to create %s in zip: %v", path, err)
					continue
				}
				_, err = io.WriteString(cssFile, resource.Content)
				if err != nil {
					log.Printf("⚠️ Failed to write %s to zip: %v", path, err)
					continue
				}
				log.Printf("✅ Added external CSS: %s", path)
			} else if resource.Error != nil {
				log.Printf("⚠️ Skipping failed CSS resource %s: %v", resource.URL, resource.Error)
			}
		}
	}

	// Add external JS files
	if len(externalJS) > 0 {
		log.Printf("📁 Adding %d external JS files to zip", len(externalJS))
		for _, resource := range externalJS {
			if resource.Error == nil && resource.Content != "" {
				path := "external/js/" + resource.Filename
				jsFile, err := writer.Create(path)
				if err != nil {
					log.Printf("⚠️ Failed to create %s in zip: %v", path, err)
					continue
				}
				_, err = io.WriteString(jsFile, resource.Content)
				if err != nil {
					log.Printf("⚠️ Failed to write %s to zip: %v", path, err)
					continue
				}
				log.Printf("✅ Added external JS: %s", path)
			} else if resource.Error != nil {
				log.Printf("⚠️ Skipping failed JS resource %s: %v", resource.URL, resource.Error)
			}
		}
	}

	// Close the zip writer
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
