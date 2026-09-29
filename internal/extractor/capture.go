package extractor

import "encoding/json"

const CaptureManifestVersion = 1

type CaptureStatus string

const (
	CaptureLocalized        CaptureStatus = "localized"
	CaptureRetainedExternal CaptureStatus = "retained-external"
	CaptureFailed           CaptureStatus = "failed"
)

// CaptureAsset records what happened to one render/runtime asset during
// capture. Path is present only when the asset was written into the export.
type CaptureAsset struct {
	URL    string        `json:"url"`
	Path   string        `json:"path,omitempty"`
	Type   string        `json:"type"`
	Status CaptureStatus `json:"status"`
	Error  string        `json:"error,omitempty"`
}

// CaptureManifest describes whether a capture contains every asset Uncluster
// attempted to localize. Retained external assets are intentional and do not
// make the capture incomplete; failed render/runtime assets do.
type CaptureManifest struct {
	Version   int            `json:"version"`
	SourceURL string         `json:"source_url,omitempty"`
	Complete  bool           `json:"complete"`
	Assets    []CaptureAsset `json:"assets"`
}

func NewCaptureManifest(sourceURL string) *CaptureManifest {
	return &CaptureManifest{
		Version:   CaptureManifestVersion,
		SourceURL: sourceURL,
		Complete:  true,
		Assets:    []CaptureAsset{},
	}
}

func (m *CaptureManifest) AddAsset(asset CaptureAsset) {
	if m == nil {
		return
	}
	m.Assets = append(m.Assets, asset)
	if asset.Status == CaptureFailed {
		m.Complete = false
	}
}

func (m *CaptureManifest) JSON() ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
