package nodejs

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
)

func CreateProjectZip(files map[string]string, projectName string) ([]byte, error) {
	log.Printf("Creating zip archive for project: %s", projectName)

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	for filepath, content := range files {
		fullPath := projectName + "/" + filepath

		file, err := writer.Create(fullPath)
		if err != nil {
			log.Printf("Failed to create %s in zip: %v", fullPath, err)
			continue
		}

		_, err = io.WriteString(file, content)
		if err != nil {
			log.Printf("Failed to write %s to zip: %v", fullPath, err)
			continue
		}

		log.Printf("Added to zip: %s", fullPath)
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	zipData := buf.Bytes()
	log.Printf("Zip archive created successfully (size: %d bytes)", len(zipData))

	return zipData, nil
}
