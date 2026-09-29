package zipper

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/omariomari2/uncluster/internal/extractor"
)

func TestCreateZipWithMetadataWritesCaptureManifest(t *testing.T) {
	manifest := &extractor.CaptureManifest{
		Version:   1,
		SourceURL: "https://example.com/page",
		Complete:  false,
		Assets: []extractor.CaptureAsset{
			{
				URL:    "https://example.com/site.css",
				Path:   "external/css/site.css",
				Type:   "css",
				Status: extractor.CaptureLocalized,
			},
			{
				URL:    "https://example.com/app.js",
				Type:   "js",
				Status: extractor.CaptureFailed,
				Error:  "HTTP 503",
			},
		},
	}

	data, err := CreateZipWithMetadata("<html></html>", nil, nil, nil, nil, nil, manifest)
	if err != nil {
		t.Fatalf("CreateZipWithMetadata() error = %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	var manifestData []byte
	for _, file := range zr.File {
		if file.Name != "uncluster-capture.json" {
			continue
		}
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		manifestData, err = io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(manifestData) == 0 {
		t.Fatal("ZIP is missing uncluster-capture.json")
	}

	var got extractor.CaptureManifest
	if err := json.Unmarshal(manifestData, &got); err != nil {
		t.Fatalf("decode capture manifest: %v", err)
	}
	if got.Version != 1 || got.SourceURL != manifest.SourceURL || got.Complete {
		t.Fatalf("capture manifest header = %+v", got)
	}
	if len(got.Assets) != 2 || got.Assets[0].Status != extractor.CaptureLocalized || got.Assets[1].Status != extractor.CaptureFailed {
		t.Fatalf("capture manifest assets = %+v", got.Assets)
	}
}
