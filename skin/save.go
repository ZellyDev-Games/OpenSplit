package skin

import (
	"github.com/zellydev-games/opensplit/logger"
	skinfiles "github.com/zellydev-games/opensplit/skin/files"
)

// SaveWorkingCopy persists the current editor working copy to disk.
func (s *Service) SaveWorkingCopy() error {
	_,
		rules,
		dirty,
		target,
		active,
		revision := s.editor.Snapshot()

	logger.Infof(
		logModule,
		"SaveWorkingCopy: rules=%d dirty=%t revision=%d target=%+v activeRule=%v",
		len(rules),
		dirty,
		revision,
		target,
		active != nil,
	)

	if !dirty {
		logger.Infof(
			logModule,
			"SaveWorkingCopy: working copy is clean; nothing to write",
		)

		return nil
	}

	s.suppressSaveWatcherEvents()

	if err := s.saveEditor(); err != nil {
		return err
	}

	if err := s.EmitSkinModel(); err != nil {
		return err
	}

	s.notifySkinUpdated()

	return nil
}

// saveEditor writes the current editor working copy to the selected skin.
func (s *Service) saveEditor() error {
	files,
		_,
		_,
		_,
		_,
		_ := s.editor.Snapshot()

	root := s.GetSkinPath()

	textFiles := make(
		[]skinfiles.TextFile,
		0,
		len(files),
	)

	for _, file := range files {
		if !file.Text {
			continue
		}

		logger.Infof(
			logModule,
			"saveEditor: writing %q length=%d",
			file.Path,
			len(file.Contents),
		)

		textFiles = append(
			textFiles,
			skinfiles.TextFile{
				Path:     file.Path,
				Contents: file.Contents,
			},
		)
	}

	if err := skinfiles.SaveTextFiles(
		root,
		textFiles,
	); err != nil {
		return err
	}

	s.editor.ClearDirty()

	return nil
}
