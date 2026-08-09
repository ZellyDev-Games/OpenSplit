package skin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/logger"
)

func (s *Service) CreateSkin(
	name string,
	base string,
) error {
	name = strings.TrimSpace(name)
	base = strings.TrimSpace(base)

	if name == "" {
		return errors.New(
			"skin name cannot be empty",
		)
	}

	if strings.ContainsAny(
		name,
		`/\`,
	) {
		return errors.New(
			"invalid skin name",
		)
	}

	if base == "" {
		base = "default"
	}

	available := s.GetAvailableSkins()

	if !slices.Contains(
		available,
		base,
	) {
		return fmt.Errorf(
			"base skin %q does not exist",
			base,
		)
	}

	target := filepath.Join(
		s.skinDir,
		name,
	)

	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf(
			"skin %q already exists",
			name,
		)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	source := filepath.Join(
		s.skinDir,
		base,
	)

	err := filepath.Walk(
		source,
		func(
			path string,
			info os.FileInfo,
			err error,
		) error {
			if err != nil {
				return err
			}

			relative, err := filepath.Rel(
				source,
				path,
			)
			if err != nil {
				return err
			}

			destination := filepath.Join(
				target,
				relative,
			)

			if info.IsDir() {
				return os.MkdirAll(
					destination,
					0o755,
				)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(
				filepath.Dir(destination),
				0o755,
			); err != nil {
				return err
			}

			return os.WriteFile(
				destination,
				data,
				0o644,
			)
		},
	)

	if err != nil {
		return err
	}

	logger.Infof(
		logModule,
		"created skin %s from %s",
		name,
		base,
	)

	return nil
}

/*
CreateCSSFile creates a new editable CSS file
inside the selected skin working copy.

The file is NOT written to disk here.

Disk changes are committed only by SKIN_SAVE.
*/
func (s *Service) CreateCSSFile(
	name string,
) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.New(
			"css file name cannot be empty",
		)
	}

	if filepath.Ext(name) != ".css" {
		name += ".css"
	}

	clean := filepath.ToSlash(
		filepath.Clean(name),
	)

	if clean == "." ||
		clean == ".." ||
		strings.HasPrefix(
			clean,
			"../",
		) {
		return errors.New(
			"invalid css file path",
		)
	}

	if filepath.IsAbs(clean) {
		return errors.New(
			"invalid css file path",
		)
	}

	if _, err := s.skinRoot(); err != nil {
		return err
	}

	files, _, _, _, _, _ := s.editor.Snapshot()

	for _, file := range files {
		if file.Path == clean {
			return fmt.Errorf(
				"css file %q already exists",
				clean,
			)
		}
	}

	base := strings.TrimSuffix(
		s.GetSkinAddress(),
		"/"+EntryPoint,
	)

	s.editor.AddFile(
		dto.CSSFile{
			Name: filepath.Base(clean),

			Path: clean,

			Contents: "",

			Text: true,

			URL: base + "/" + clean,

			Type: "css",
		},
	)

	logger.Infof(
		logModule,
		"created css file in working copy %s",
		clean,
	)

	return s.EmitSkinModel()
}
