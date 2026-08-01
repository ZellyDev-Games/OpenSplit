package skin

import (
	"sync"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// EditorState contains the active skin editor working copy.
//
// Filesystem is the persisted source.
// EditorState is the editable in-memory copy.
//
// Changes are committed only by SaveWorkingCopy.
// ReloadEditor discards this state and rebuilds from disk.
type EditorState struct {
	m sync.RWMutex

	Files []dto.CSSFile
	Rules []parser.Rule

	PreviewElements []dto.SkinPreviewElement

	Target dto.SkinEditorTarget

	ActiveRule *dto.CSSRuleEditor

	Dirty bool

	Revision uint64
}

func NewEditorState() *EditorState {
	return &EditorState{
		Files: make([]dto.CSSFile, 0),
		Rules: make([]parser.Rule, 0),

		Target: dto.SkinEditorTarget{
			Mode: "file",
		},
	}
}

func cloneFiles(
	files []dto.CSSFile,
) []dto.CSSFile {

	if files == nil {
		return nil
	}

	out := make(
		[]dto.CSSFile,
		len(files),
	)

	copy(
		out,
		files,
	)

	return out
}

func cloneDeclarations(
	declarations []parser.Declaration,
) []parser.Declaration {

	if declarations == nil {
		return nil
	}

	out := make(
		[]parser.Declaration,
		len(declarations),
	)

	for i, declaration := range declarations {

		out[i] = declaration

		out[i].Value =
			append(
				[]string{},
				declaration.Value...,
			)

		out[i].Raw =
			append(
				[]string{},
				declaration.Raw...,
			)

		out[i].LeadingComment =
			append(
				[]string{},
				declaration.LeadingComment...,
			)
	}

	return out
}

func cloneRules(
	rules []parser.Rule,
) []parser.Rule {

	if rules == nil {
		return nil
	}

	out := make(
		[]parser.Rule,
		len(rules),
	)

	for i, rule := range rules {

		out[i] = rule

		out[i].Declarations =
			cloneDeclarations(
				rule.Declarations,
			)

		out[i].Children =
			cloneRules(
				rule.Children,
			)
	}

	return out
}

func (e *EditorState) SetPreviewElements(
	elements []dto.SkinPreviewElement,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.PreviewElements = append(
		[]dto.SkinPreviewElement{},
		elements...,
	)
}

func (e *EditorState) ReplaceWorkingCopyFromDisk(
	files []dto.CSSFile,
	rules []parser.Rule,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.Files =
		cloneFiles(
			files,
		)

	e.Rules =
		cloneRules(
			rules,
		)

	e.Dirty = false

	e.Revision++
}

func (e *EditorState) ReplaceWorkingCopy(
	files []dto.CSSFile,
	rules []parser.Rule,
) {

	e.m.Lock()
	defer e.m.Unlock()

	e.Files =
		cloneFiles(
			files,
		)

	e.Rules =
		cloneRules(
			rules,
		)

	e.Dirty = true
}

func (e *EditorState) Snapshot() (
	[]dto.CSSFile,
	[]parser.Rule,
	bool,
	dto.SkinEditorTarget,
	*dto.CSSRuleEditor,
	uint64,
) {
	e.m.RLock()
	defer e.m.RUnlock()

	var active *dto.CSSRuleEditor

	if e.ActiveRule != nil {

		copy := *e.ActiveRule
		active = &copy
	}

	return cloneFiles(e.Files),
		cloneRules(e.Rules),
		e.Dirty,
		e.Target,
		active,
		e.Revision
}

func (e *EditorState) SetTarget(
	target dto.SkinEditorTarget,
) {

	e.m.Lock()
	defer e.m.Unlock()

	e.Target = target
}

func (e *EditorState) GetTarget() dto.SkinEditorTarget {

	e.m.RLock()
	defer e.m.RUnlock()

	return e.Target
}

func (e *EditorState) SetActiveRule(
	rule *dto.CSSRuleEditor,
) {

	e.m.Lock()
	defer e.m.Unlock()

	if rule == nil {

		e.ActiveRule = nil
		return
	}

	copy := *rule
	e.ActiveRule = &copy
}

func (e *EditorState) GetActiveRule() *dto.CSSRuleEditor {

	e.m.RLock()
	defer e.m.RUnlock()

	if e.ActiveRule == nil {
		return nil
	}

	copy := *e.ActiveRule

	return &copy
}

func (e *EditorState) MarkDirty() {

	e.m.Lock()
	defer e.m.Unlock()

	e.Dirty = true
}

func (e *EditorState) ClearDirty() {

	e.m.Lock()
	defer e.m.Unlock()

	e.Dirty = false
}

func (e *EditorState) IsDirty() bool {

	e.m.RLock()
	defer e.m.RUnlock()

	return e.Dirty
}

func (e *EditorState) ClearSelection() {

	e.m.Lock()
	defer e.m.Unlock()

	e.Target = dto.SkinEditorTarget{
		Mode: "file",
	}

	e.ActiveRule = nil
}

func (e *EditorState) UpdateFile(
	file dto.CSSFile,
) {

	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Files {

		if e.Files[i].Path == file.Path {

			e.Files[i] = file
			e.Dirty = true

			return
		}
	}
}

func (e *EditorState) AddFile(
	file dto.CSSFile,
) {

	e.m.Lock()
	defer e.m.Unlock()

	e.Files =
		append(
			e.Files,
			file,
		)

	e.Dirty = true
}

func (e *EditorState) UpdateRule(
	rule parser.Rule,
) bool {

	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Rules {

		if e.Rules[i].ID == rule.ID {

			e.Rules[i] =
				cloneRules(
					[]parser.Rule{rule},
				)[0]

			e.Dirty = true

			return true
		}
	}

	return false
}

func (e *EditorState) AddRule(
	rule parser.Rule,
) {

	e.m.Lock()
	defer e.m.Unlock()

	e.Rules =
		append(
			e.Rules,
			cloneRules(
				[]parser.Rule{rule},
			)[0],
		)

	e.Dirty = true
}

func (e *EditorState) DeleteRule(
	id string,
) bool {

	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Rules {

		if e.Rules[i].ID == id {

			e.Rules =
				append(
					e.Rules[:i],
					e.Rules[i+1:]...,
				)

			e.Dirty = true

			return true
		}
	}

	return false
}

func (e *EditorState) NewRule(
	file string,
	selector string,
) dto.CSSRuleEditor {

	return dto.CSSRuleEditor{
		File:     file,
		Selector: selector,
		Create:   true,
	}
}

func (e *EditorState) UpdateFileContents(
	path string,
	contents string,
) {

	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Files {

		if e.Files[i].Path != path {
			continue
		}

		e.Files[i].Contents =
			contents

		e.Dirty = true

		e.Revision++

		return
	}
}
