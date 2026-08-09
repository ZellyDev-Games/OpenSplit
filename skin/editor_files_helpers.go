package skin

import (
	"path/filepath"
	"strings"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
)

// buildEditorFiles reads the files belonging to the selected skin and converts
// them into the editor's working-copy representation.
//
// Files are loaded from disk only. This function does not mutate EditorState.
func (s *Service) buildEditorFiles() (
	[]dto.CSSFile,
	error,
) {
	files, err := s.SkinFiles()
	if err != nil {
		return nil, err
	}

	out := make(
		[]dto.CSSFile,
		0,
		len(files),
	)

	base := strings.TrimSuffix(
		s.GetSkinAddress(),
		"/"+EntryPoint,
	)

	for _, file := range files {
		contents := ""

		text := isTextFile(file)

		if text {
			contents, err = s.ReadFile(file)
			if err != nil {
				return nil, err
			}

			logger.Infof(
				logModule,
				"%s length=%d",
				file,
				len(contents),
			)
		}

		relative := filepath.ToSlash(file)

		out = append(
			out,
			dto.CSSFile{
				Name: filepath.Base(relative),

				Path: relative,

				Contents: contents,

				OriginalContents: contents,

				Text: text,

				URL: base + "/" + relative,

				Type: fileType(relative),
			},
		)
	}

	return out, nil
}

// fileType returns the editor-facing type for a skin file.
func fileType(
	filePath string,
) string {
	ext := strings.ToLower(
		filepath.Ext(filePath),
	)

	switch ext {
	case ".css":
		return "css"

	case ".png",
		".jpg",
		".jpeg",
		".gif",
		".webp":
		return "image"

	case ".svg":
		return "svg"

	case ".txt",
		".md",
		".html",
		".js",
		".json":
		return "text"

	default:
		return "binary"
	}
}

// isTextFile reports whether a skin file should have its contents loaded into
// the editor working copy.
//
// Binary assets are represented by metadata only.
func isTextFile(
	filePath string,
) bool {
	switch strings.ToLower(
		filepath.Ext(filePath),
	) {
	case ".css",
		".txt",
		".md",
		".html",
		".js",
		".json":
		return true

	default:
		return false
	}
}
