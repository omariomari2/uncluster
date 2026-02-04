package nodejs

import (
	"archive/zip"
	"bytes"
	"io"
)

func CreateProjectZip(files map[string]string, projectName string) ([]byte, error) {

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for filepath, content := range files {
		fullPath := projectName + "/" + filepath

		file, err := writer.Create(fullPath)
		if err != nil {
			continue
		}

		_, err = io.WriteString(file, content)
		if err != nil {
			continue
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	zipData := buf.Bytes()

	return zipData, nil
}
