package files

import (
	"os"
	"path/filepath"
)

// TextFile represents a text file that should be persisted to disk.
type TextFile struct {
	Path     string
	Contents string
}

// SaveTextFiles writes the supplied text files below root.
func SaveTextFiles(
	root string,
	files []TextFile,
) error {
	for _, file := range files {
		filename, err := ResolvePath(
			root,
			file.Path,
		)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(
			filepath.Dir(filename),
			0o755,
		); err != nil {
			return err
		}

		if err := os.WriteFile(
			filename,
			[]byte(file.Contents),
			0o644,
		); err != nil {
			return err
		}
	}

	return nil
}
