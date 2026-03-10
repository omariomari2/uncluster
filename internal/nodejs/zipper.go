package nodejs

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
)

func CreateProjectZip(files map[string]string, projectName string) ([]byte, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	written := 0
	for filepath, content := range files {
		fullPath := projectName + "/" + filepath

		file, err := writer.Create(fullPath)
		if err != nil {
			log.Printf("zip: failed to create entry %s: %v", fullPath, err)
			continue
		}

		if _, err = io.WriteString(file, content); err != nil {
			log.Printf("zip: failed to write entry %s: %v", fullPath, err)
			continue
		}
		written++
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	if written == 0 && len(files) > 0 {
		return nil, fmt.Errorf("failed to write any files to zip archive")
	}

	return buf.Bytes(), nil
}
