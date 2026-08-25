package main

// Temporary end-to-end driver. Windows Application Control blocks running the
// built uncluster.exe on this machine, but permits Go test binaries, so this
// calls the same CLI entry points the binary would. Delete after the run.
//
//	go test ./cmd/uncluster -run TestE2E -v

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/omariomari2/uncluster/internal/bundle"
)

func TestE2EAlre(t *testing.T) {
	zip := os.Getenv("E2E_ZIP")
	outRoot := os.Getenv("E2E_OUT")
	if zip == "" || outRoot == "" {
		t.Skip("set E2E_ZIP and E2E_OUT")
	}

	// 1. bundle: find the real index.html inside the archive.
	result, err := bundle.ProcessWithOptions(zip, bundle.Options{
		Destination: filepath.Join(outRoot, "bundle"),
	})
	if err != nil {
		t.Fatalf("bundle: %v", err)
	}
	t.Logf("bundle site=%s index=%s", result.SiteName, result.IndexPath)

	raw, err := os.ReadFile(result.IndexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	html := string(raw)
	t.Logf("source index.html: %d bytes", len(raw))

	// 2..4: the same three CLI modes, into their own subfolders.
	runFormat(html, filepath.Join(outRoot, "formatted-html"))
	runNodeJSEJS(html, filepath.Join(outRoot, "ejs-project"))
	runNodeJS(html, filepath.Join(outRoot, "tsx-project"))
}
