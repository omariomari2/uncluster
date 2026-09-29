package extractor

import "testing"

func TestExtractRecordsRetainedRuntimeAssetsButNotNavigation(t *testing.T) {
	extracted, err := Extract(`<!doctype html><html><head>
		<link rel="preconnect" href="https://static.example.com">
		<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter">
		</head><body>
		<a href="https://example.com/next">Next</a>
		<iframe src="https://player.example.com/embed/123"></iframe>
		</body></html>`)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	want := []CaptureAsset{
		{URL: "https://fonts.googleapis.com/css2?family=Inter", Type: "css", Status: CaptureRetainedExternal},
		{URL: "https://player.example.com/embed/123", Type: "iframe", Status: CaptureRetainedExternal},
	}
	if len(extracted.Capture.Assets) != len(want) {
		t.Fatalf("capture assets = %+v, want %+v", extracted.Capture.Assets, want)
	}
	for i := range want {
		if extracted.Capture.Assets[i] != want[i] {
			t.Fatalf("capture asset %d = %+v, want %+v", i, extracted.Capture.Assets[i], want[i])
		}
	}
	if !extracted.Capture.Complete {
		t.Fatal("retained external assets made capture incomplete")
	}
}
