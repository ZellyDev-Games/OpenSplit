package editor

import (
	"sync"

	"github.com/zellydev-games/opensplit/dto"
	"github.com/zellydev-games/opensplit/skin/parser"
)

// EditorState owns the in-memory working copy used by the skin editor.
//
// The filesystem remains the persisted source of truth.
//
// EditorState owns:
//   - editable files
//   - parsed CSS rules
//   - preview elements
//   - current selection
//   - active rule editor
//   - dirty state
//   - working-copy revision
//
// Changes are committed to disk only by the save operation.
//
// Reloading the editor replaces the working copy with the current
// filesystem contents and clears the dirty state.
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

// NewEditorState creates an empty editor working copy.
func NewEditorState() *EditorState {
	return &EditorState{
		Files: make(
			[]dto.CSSFile,
			0,
		),

		Rules: make(
			[]parser.Rule,
			0,
		),

		PreviewElements: make(
			[]dto.SkinPreviewElement,
			0,
		),

		Target: dto.SkinEditorTarget{
			Mode: "file",
		},
	}
}

// SetPreviewElements replaces the cached preview elements.
//
// Preview elements are derived editor data and do not make the working
// copy dirty.
func (e *EditorState) SetPreviewElements(
	elements []dto.SkinPreviewElement,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.PreviewElements = ClonePreviewElements(
		elements,
	)
}

// GetPreviewElements returns an independent copy of the cached preview
// elements.
func (e *EditorState) GetPreviewElements() []dto.SkinPreviewElement {
	e.m.RLock()
	defer e.m.RUnlock()

	return ClonePreviewElements(
		e.PreviewElements,
	)
}

// ReplaceWorkingCopyFromDisk replaces the complete working copy with
// freshly loaded filesystem state.
func (e *EditorState) ReplaceWorkingCopyFromDisk(
	files []dto.CSSFile,
	rules []parser.Rule,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.Files = CloneFiles(files)
	e.Rules = CloneRules(rules)

	e.Dirty = false
	e.Revision++
}

// ReplaceWorkingCopy replaces the complete working copy with another
// in-memory working copy.
//
// This operation represents an editor mutation and therefore marks the
// state dirty.
func (e *EditorState) ReplaceWorkingCopy(
	files []dto.CSSFile,
	rules []parser.Rule,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.Files = CloneFiles(files)
	e.Rules = CloneRules(rules)

	e.Dirty = true
	e.Revision++
}

// Snapshot returns isolated copies of the current editor state.
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

	return CloneFiles(e.Files),
		CloneRules(e.Rules),
		e.Dirty,
		e.Target,
		active,
		e.Revision
}

// SetTarget changes the current editor selection.
func (e *EditorState) SetTarget(
	target dto.SkinEditorTarget,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.Target = target
}

// GetTarget returns the current editor selection.
func (e *EditorState) GetTarget() dto.SkinEditorTarget {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.Target
}

// SetActiveRule changes the rule currently displayed by the rule editor.
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

// GetActiveRule returns an isolated copy of the active rule.
func (e *EditorState) GetActiveRule() *dto.CSSRuleEditor {
	e.m.RLock()
	defer e.m.RUnlock()

	if e.ActiveRule == nil {
		return nil
	}

	copy := *e.ActiveRule

	return &copy
}

// MarkDirty marks the working copy as modified.
func (e *EditorState) MarkDirty() {
	e.m.Lock()
	defer e.m.Unlock()

	e.Dirty = true
}

// ClearDirty marks the current working copy as persisted.
func (e *EditorState) ClearDirty() {
	e.m.Lock()
	defer e.m.Unlock()

	e.Dirty = false
}

// IsDirty reports whether the working copy contains unsaved changes.
func (e *EditorState) IsDirty() bool {
	e.m.RLock()
	defer e.m.RUnlock()

	return e.Dirty
}

// ClearSelection resets the editor selection to file mode.
func (e *EditorState) ClearSelection() {
	e.m.Lock()
	defer e.m.Unlock()

	e.Target = dto.SkinEditorTarget{
		Mode: "file",
	}

	e.ActiveRule = nil
}

// NewRule creates the editor representation for a rule that has not yet
// been inserted into the working rule tree.
func (e *EditorState) NewRule(
	file string,
	selector string,
	layer string,
) dto.CSSRuleEditor {
	return dto.CSSRuleEditor{
		File:     file,
		Selector: selector,
		Layer:    layer,
		Create:   true,
	}
}
