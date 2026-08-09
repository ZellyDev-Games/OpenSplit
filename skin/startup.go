package skin

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zellydev-games/opensplit/logger"
)

//go:embed default-skin.zip
var DefaultSkinZip []byte

// Startup validates the configured skin directory, installs the embedded
// default skin, restores the configured skin, and starts filesystem watching.
func (s *Service) Startup() error {
	if err := s.validateSkinDirectory(); err != nil {
		return err
	}

	if err := s.installDefaultSkin(); err != nil {
		return err
	}

	if err := s.restoreSelectedSkin(); err != nil {
		return err
	}

	s.startSkinWatcher()

	return nil
}

// validateSkinDirectory verifies that the configured skin directory is suitable
// for destructive default-skin installation.
func (s *Service) validateSkinDirectory() error {
	if s.skinDir == "" {
		return errors.New("skinDir not set")
	}

	if !filepath.IsAbs(s.skinDir) {
		msg := fmt.Sprintf(
			"skinDir must be absolute: %s",
			s.skinDir,
		)

		logger.Error(
			logModule,
			msg,
		)

		return errors.New(msg)
	}

	target := filepath.Clean(s.skinDir)

	if !strings.Contains(target, "OpenSplit") {
		msg := fmt.Sprintf(
			"refusing to delete outside OpenSplit directory: %s",
			target,
		)

		logger.Error(
			logModule,
			msg,
		)

		return errors.New(msg)
	}

	return nil
}

// installDefaultSkin replaces the installed default skin with the embedded
// default skin shipped with the application.
func (s *Service) installDefaultSkin() error {
	target := filepath.Join(
		s.skinDir,
		"default",
	)

	if err := os.RemoveAll(target); err != nil {
		return err
	}

	if err := os.MkdirAll(
		target,
		0o755,
	); err != nil {
		return err
	}

	reader, err := zip.NewReader(
		bytes.NewReader(DefaultSkinZip),
		int64(len(DefaultSkinZip)),
	)
	if err != nil {
		return err
	}

	logger.Info(
		logModule,
		"extracting default skin",
	)

	for _, file := range reader.File {
		if err := s.extractDefaultSkinFile(
			target,
			file,
		); err != nil {
			return err
		}
	}

	return nil
}

// extractDefaultSkinFile extracts one entry from the embedded default skin
// while preventing ZipSlip paths.
func (s *Service) extractDefaultSkinFile(
	target string,
	file *zip.File,
) error {
	p := filepath.Join(
		target,
		file.Name,
	)

	logger.Infof(
		logModule,
		"default skin extracted to %s",
		p,
	)

	cleanTarget := filepath.Clean(target)
	cleanPath := filepath.Clean(p)

	if !strings.HasPrefix(
		cleanPath+string(os.PathSeparator),
		cleanTarget+string(os.PathSeparator),
	) {
		return fmt.Errorf(
			"illegal zip path: %s",
			file.Name,
		)
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(
			p,
			0o755,
		)
	}

	if err := os.MkdirAll(
		filepath.Dir(p),
		0o755,
	); err != nil {
		return err
	}

	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()

	output, err := os.OpenFile(
		p,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		0o644,
	)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(
		output,
		reader,
	)

	closeErr := output.Close()

	if copyErr != nil {
		return copyErr
	}

	return closeErr
}

// restoreSelectedSkin restores the configured skin and falls back to the
// default skin when the configured skin is unavailable or cannot be loaded.
func (s *Service) restoreSelectedSkin() error {
	selected := s.configService.SelectedSkin

	if selected == "" {
		return s.setDefaultSkin()
	}

	if err := s.SetSkin(
		selected,
		false,
	); err == nil {
		return nil
	} else {
		logger.Warnf(
			logModule,
			"failed to set selected skin %s, fallback setskin to default: %v",
			selected,
			err,
		)
	}

	if err := s.setDefaultSkin(); err != nil {
		logger.Warnf(
			logModule,
			"default setskin fallback failed, manual fallback: %v",
			err,
		)

		s.setSelectedSkin("default")
	}

	return nil
}

// setDefaultSkin attempts to select the installed default skin and persists
// that selection when successful.
func (s *Service) setDefaultSkin() error {
	if err := s.SetSkin(
		"default",
		true,
	); err != nil {
		logger.Warnf(
			logModule,
			"failed to set default skin, manual fallback: %v",
			err,
		)

		s.setSelectedSkin("default")

		return err
	}

	return nil
}

// setSelectedSkin updates the selected skin under the service lock.
func (s *Service) setSelectedSkin(name string) {
	s.m.Lock()
	s.selectedSkin = name
	s.m.Unlock()
}

// startSkinWatcher begins watching the currently selected skin directory.
func (s *Service) startSkinWatcher() {
	selected := s.SelectedSkin()

	if selected == "" {
		return
	}

	watchDir := filepath.Join(
		s.skinDir,
		selected,
	)

	logger.Infof(
		logModule,
		"watching skin directory %s",
		watchDir,
	)

	s.watcher.Start(
		watchDir,
		s.skinUpdated,
	)
}
