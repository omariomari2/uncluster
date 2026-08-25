package nodejs

import (
	"strings"
	"testing"
)

// The generated Express server used to render index.ejs for every unmatched
// GET, so a request for a missing stylesheet returned 200 with HTML in it. The
// browser then discarded it silently and the page rendered under-styled with no
// error anywhere. Missing files must 404.
func TestEJSServerDoesNotServeHTMLForMissingAssets(t *testing.T) {
	project, err := GenerateEJSProject(&EJSProjectConfig{
		ProjectName: "server-fixture",
		HTML:        `<html><body><main>Fixture</main></body></html>`,
	})
	if err != nil {
		t.Fatalf("GenerateEJSProject() error = %v", err)
	}

	server, ok := project.Files["server.js"]
	if !ok {
		t.Fatal("generated project has no server.js")
	}

	for _, want := range []string{
		"path.extname(req.path)",
		"res.status(404)",
	} {
		if !strings.Contains(server, want) {
			t.Errorf("server.js missing %q; got:\n%s", want, server)
		}
	}

	// The unguarded catch-all is the specific shape that caused the bug.
	if strings.Contains(server, "app.get('*', (req, res) => {") {
		t.Errorf("server.js still renders the page for every unmatched request; got:\n%s", server)
	}
}
