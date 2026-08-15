package skin

import (
	"errors"
	"regexp"
)

var splitterLayoutDeclaration = regexp.MustCompile(
	`(--splitter-layout\s*:\s*)(horizontal|vertical)(\s*;)`,
)

func (s *Service) SetDefaultLayout(
	layout string,
) error {
	if layout != "horizontal" &&
		layout != "vertical" {
		return errors.New("invalid skin layout")
	}

	files, _, _, _, _, _ := s.editor.Snapshot()

	for _, file := range files {
		if !file.Text {
			continue
		}

		if !splitterLayoutDeclaration.MatchString(
			file.Contents,
		) {
			continue
		}

		contents := splitterLayoutDeclaration.ReplaceAllString(
			file.Contents,
			`${1}`+layout+`${3}`,
		)

		return s.UpdateFileContents(
			file.Path,
			contents,
		)
	}

	return errors.New(
		"--splitter-layout declaration not found",
	)
}
