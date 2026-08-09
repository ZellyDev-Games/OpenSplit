package files

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateSkin copies an existing skin directory to a new skin directory.
func CreateSkin(
	skinDir string,
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

	if name == "." ||
		name == ".." ||
		strings.ContainsAny(
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

	source := filepath.Join(
		skinDir,
		base,
	)

	target := filepath.Join(
		skinDir,
		name,
	)

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf(
			"base skin %q does not exist: %w",
			base,
			err,
		)
	}

	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf(
			"skin %q already exists",
			name,
		)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return filepath.Walk(
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
}

// ValidateCSSFilePath validates a CSS path without touching the filesystem.
func ValidateCSSFilePath(name string) (string, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return "", errors.New(
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
		return "", errors.New(
			"invalid css file path",
		)
	}

	if filepath.IsAbs(clean) {
		return "", errors.New(
			"invalid css file path",
		)
	}

	return clean, nil
}
