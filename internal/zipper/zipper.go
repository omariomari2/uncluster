package zipper

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"

	"github.com/omariomari2/uncluster/internal/extractor"
	"github.com/omariomari2/uncluster/internal/fetcher"
)

func CreateZipWithMetadata(html string, inlineCSS, inlineJS []extractor.InlineResource, externalCSS, externalJS []fetcher.FetchedResource, localAssets []extractor.LocalAsset) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	if html != "" {
		if err := writeEntry(writer, "index.html", []byte(html)); err != nil {
			return nil, err
		}
	}

	if err := writeInlineResources(writer, inlineCSS); err != nil {
		return nil, err
	}
	if err := writeInlineResources(writer, inlineJS); err != nil {
		return nil, err
	}
	if err := writeFetchedResources(writer, "external/css/", externalCSS); err != nil {
		return nil, err
	}
	if err := writeFetchedResources(writer, "external/js/", externalJS); err != nil {
		return nil, err
	}

	for _, asset := range localAssets {
		if len(asset.Content) == 0 {
			continue
		}
		if err := writeEntry(writer, asset.Path, asset.Content); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeInlineResources(writer *zip.Writer, resources []extractor.InlineResource) error {
	for _, resource := range resources {
		if resource.Content == "" {
			continue
		}
		if err := writeEntry(writer, resource.Path, []byte(resource.Content)); err != nil {
			return err
		}
	}
	return nil
}

func writeFetchedResources(writer *zip.Writer, dir string, resources []fetcher.FetchedResource) error {
	for _, resource := range resources {
		if resource.Error != nil || resource.Content == "" {
			continue
		}
		if err := writeEntry(writer, dir+resource.Filename, []byte(resource.Content)); err != nil {
			return err
		}
	}
	return nil
}

func writeEntry(writer *zip.Writer, path string, content []byte) error {
	entry, err := writer.Create(path)
	if err != nil {
		return fmt.Errorf("create ZIP entry %q: %w", path, err)
	}
	if _, err := io.Copy(entry, bytes.NewReader(content)); err != nil {
		return fmt.Errorf("write ZIP entry %q: %w", path, err)
	}
	return nil
}
