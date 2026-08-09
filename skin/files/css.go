package files

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Root returns the absolute filesystem path of the selected skin.
func Root(
	skinDir string,
	selected string,
) (string, error) {
	if selected == "" {
		return "", errors.New(
			"no skin selected",
		)
	}

	return filepath.Join(
		skinDir,
		selected,
	), nil
}

// ResolvePath resolves file relative to root while ensuring the resulting
// path remains inside root.
func ResolvePath(
	root string,
	file string,
) (string, error) {
	clean := filepath.Clean(file)

	if clean == "." ||
		clean == ".." ||
		strings.HasPrefix(
			clean,
			".."+string(os.PathSeparator),
		) {
		return "", errors.New(
			"invalid skin path",
		)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	target := filepath.Join(
		root,
		clean,
	)

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	if absTarget != absRoot &&
		!strings.HasPrefix(
			absTarget,
			absRoot+string(os.PathSeparator),
		) {
		return "", errors.New(
			"file outside skin directory",
		)
	}

	return target, nil
}

// List returns all files inside root as slash-separated relative paths.
func List(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(
		root,
		func(
			path string,
			info os.FileInfo,
			err error,
		) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			relative, err := filepath.Rel(
				root,
				path,
			)
			if err != nil {
				return err
			}

			files = append(
				files,
				filepath.ToSlash(relative),
			)

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	return files, nil
}

// Read reads any file inside root.
func Read(
	root string,
	file string,
) (string, error) {
	target, err := ResolvePath(
		root,
		file,
	)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ReadCSS reads a CSS file inside root.
func ReadCSS(
	root string,
	file string,
) (string, error) {
	if !strings.EqualFold(
		filepath.Ext(file),
		".css",
	) {
		return "", errors.New(
			"only css files may be read",
		)
	}

	return Read(
		root,
		file,
	)
}
