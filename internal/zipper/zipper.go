package zipper

import (
	"archive/zip"
	"bytes"
	"htmlfmt/internal/extractor"
	"htmlfmt/internal/fetcher"
	"io"
)

func CreateZipWithMetadata(html string, inlineCSS, inlineJS []extractor.InlineResource, externalCSS, externalJS []fetcher.FetchedResource) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

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

	if len(inlineCSS) > 0 {
		for _, resource := range inlineCSS {
			if resource.Content == "" {
				continue
			}
			cssFile, err := writer.Create(resource.Path)
			if err != nil {
				continue
			}
			_, err = io.WriteString(cssFile, resource.Content)
			if err != nil {
				continue
			}
		}
	}

	if len(inlineJS) > 0 {
		for _, resource := range inlineJS {
			if resource.Content == "" {
				continue
			}
			jsFile, err := writer.Create(resource.Path)
			if err != nil {
				continue
			}
			_, err = io.WriteString(jsFile, resource.Content)
			if err != nil {
				continue
			}
		}
	}

	if len(externalCSS) > 0 {
		for _, resource := range externalCSS {
			if resource.Error == nil && resource.Content != "" {
				path := "external/css/" + resource.Filename
				cssFile, err := writer.Create(path)
				if err != nil {
					continue
				}
				_, err = io.WriteString(cssFile, resource.Content)
				if err != nil {
					continue
				}
			}
		}
	}

	if len(externalJS) > 0 {
		for _, resource := range externalJS {
			if resource.Error == nil && resource.Content != "" {
				path := "external/js/" + resource.Filename
				jsFile, err := writer.Create(path)
				if err != nil {
					continue
				}
				_, err = io.WriteString(jsFile, resource.Content)
				if err != nil {
					continue
				}
			}
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
