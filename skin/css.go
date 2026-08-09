package skin

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zellydev-games/opensplit/logger"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// skinRoot returns the absolute filesystem path of the currently selected skin.
func (s *Service) skinRoot() (string, error) {
	s.m.RLock()
	selected := s.selectedSkin
	s.m.RUnlock()

	if selected == "" {
		return "", errors.New("no skin selected")
	}

	return filepath.Join(
		s.skinDir,
		selected,
	), nil
}

// resolveSkinPath resolves a path relative to the selected skin while ensuring
// that the resulting path remains inside the selected skin directory.
//
// All filesystem access to skin files should pass through this boundary.
func (s *Service) resolveSkinPath(
	file string,
) (string, error) {
	root, err := s.skinRoot()
	if err != nil {
		return "", err
	}

	clean := filepath.Clean(file)

	if clean == "." ||
		clean == ".." ||
		strings.HasPrefix(
			clean,
			".."+string(os.PathSeparator),
		) {
		return "", errors.New("invalid skin path")
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
		return "", errors.New("file outside skin directory")
	}

	return target, nil
}

func (s *Service) SkinFiles() ([]string, error) {
	root, err := s.skinRoot()
	if err != nil {
		return nil, err
	}

	var files []string

	err = filepath.Walk(
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

	for _, file := range files {
		logger.Infof(logModule, "skin file: %q", file)
	}

	return files, nil
}

func (s *Service) CSSRules(
	file string,
	selector string,
) ([]parser.Rule, error) {
	css, err := s.ReadCSS(file)
	if err != nil {
		return nil, err
	}

	return parser.FindRules(
		file,
		string(css),
		selector,
	), nil
}

// CSSRuleHierarchy returns every matching rule in skin load order.
func (s *Service) CSSRuleHierarchy(
	selector string,
) ([]parser.Rule, error) {
	files, err := s.SkinFiles()
	if err != nil {
		return nil, err
	}

	var out []parser.Rule

	for _, file := range files {
		if filepath.Ext(file) != ".css" {
			continue
		}

		rules, err := s.CSSRules(
			file,
			selector,
		)
		if err != nil {
			return nil, err
		}

		out = append(
			out,
			rules...,
		)
	}

	return out, nil
}

func (s *Service) ReadFile(
	file string,
) (string, error) {
	target, err := s.resolveSkinPath(file)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *Service) ReadCSS(
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

	target, err := s.resolveSkinPath(file)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
