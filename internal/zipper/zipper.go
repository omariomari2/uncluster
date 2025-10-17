package zipper

import (
	"archive/zip"
	"bytes"
	"io"
)

// CreateZipWithMetadata creates a zip archive containing HTML, CSS, and JS files
func CreateZipWithMetadata(html, css, js string) ([]byte, error) {
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

	// Add style.css file
	if css != "" {
		cssFile, err := writer.Create("style.css")
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(cssFile, css)
		if err != nil {
			return nil, err
		}
	}

	// Add script.js file
	if js != "" {
		jsFile, err := writer.Create("script.js")
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(jsFile, js)
		if err != nil {
			return nil, err
		}
	}

	// Close the zip writer
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
