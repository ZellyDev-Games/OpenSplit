package skin

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zellydev-games/opensplit/dto"
	skinfiles "github.com/zellydev-games/opensplit/skin/files"
	"github.com/zellydev-games/opensplit/skin/parser"
)

func (s *Service) skinRoot() (string, error) {
	s.m.RLock()
	selected := s.selectedSkin
	skinDir := s.skinDir
	s.m.RUnlock()

	return skinfiles.Root(
		skinDir,
		selected,
	)
}

func (s *Service) resolveSkinPath(
	file string,
) (string, error) {
	root, err := s.skinRoot()
	if err != nil {
		return "", err
	}

	return skinfiles.ResolvePath(
		root,
		file,
	)
}

func (s *Service) SkinFiles() ([]string, error) {
	root, err := s.skinRoot()
	if err != nil {
		return nil, err
	}

	return skinfiles.List(root)
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
	root, err := s.skinRoot()
	if err != nil {
		return "", err
	}

	return skinfiles.Read(
		root,
		file,
	)
}

func (s *Service) ReadCSS(
	file string,
) (string, error) {
	root, err := s.skinRoot()
	if err != nil {
		return "", err
	}

	return skinfiles.ReadCSS(
		root,
		file,
	)
}

func (s *Service) CreateCSSFile(
	name string,
) error {
	clean, err := skinfiles.ValidateCSSFilePath(name)
	if err != nil {
		return err
	}

	if _, err := s.skinRoot(); err != nil {
		return err
	}

	files,
		_,
		_,
		_,
		_,
		_ := s.editor.Snapshot()

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

	return s.EmitSkinModel()
}

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

	if err := validateSkinName(name); err != nil {
		return err
	}

	if base == "" {
		base = "default"
	}

	available := s.GetAvailableSkins()

	found := false

	for _, skin := range available {
		if skin == base {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf(
			"base skin %q does not exist",
			base,
		)
	}

	s.m.RLock()
	skinDir := s.skinDir
	s.m.RUnlock()

	if err := skinfiles.CreateSkin(
		skinDir,
		name,
		base,
	); err != nil {
		return err
	}

	return nil
}
