package editor

import (
	"github.com/zellydev-games/opensplit/dto"
)

// AddFile adds a new file to the working copy.
//
// The file is kept entirely in memory until the working copy is saved.
func (e *EditorState) AddFile(
	file dto.CSSFile,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.Files = append(
		e.Files,
		file,
	)

	e.Dirty = true
	e.Revision++
}

// UpdateFile replaces an existing file in the working copy.
func (e *EditorState) UpdateFile(
	file dto.CSSFile,
) bool {
	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Files {
		if e.Files[i].Path != file.Path {
			continue
		}

		e.Files[i] = file

		e.Dirty = true
		e.Revision++

		return true
	}

	return false
}

// UpdateFileContents updates the text contents of an existing working-copy
// file.
func (e *EditorState) UpdateFileContents(
	path string,
	contents string,
) bool {
	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Files {
		if e.Files[i].Path != path {
			continue
		}

		if e.Files[i].Contents == contents {
			return false
		}

		e.Files[i].Contents = contents

		e.Dirty = true
		e.Revision++

		return true
	}

	return false
}
