package editor

import (
	"path/filepath"
	"strings"

	"github.com/zellydev-games/opensplit/dto"
)

// FileInfo describes how a skin file should be represented in the editor.
func FileInfo(
	filePath string,
	contents string,
	text bool,
	url string,
) dto.CSSFile {
	relative := filepath.ToSlash(filePath)

	return dto.CSSFile{
		Name: filepath.Base(relative),

		Path: relative,

		Contents: contents,

		OriginalContents: contents,

		Text: text,

		URL: url,

		Type: FileType(relative),
	}
}

// FileType returns the editor-facing type for a skin file.
func FileType(
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

// IsTextFile reports whether a skin file should have its contents loaded into
// the editor working copy.
func IsTextFile(
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
