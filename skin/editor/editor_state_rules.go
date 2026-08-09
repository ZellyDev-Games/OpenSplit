package editor

import (
	"github.com/zellydev-games/opensplit/skin/parser"
)

// UpdateRule replaces an existing top-level rule.
func (e *EditorState) UpdateRule(
	rule parser.Rule,
) bool {
	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Rules {
		if e.Rules[i].ID != rule.ID {
			continue
		}

		e.Rules[i] = CloneRules(
			[]parser.Rule{
				rule,
			},
		)[0]

		e.syncActiveRuleDeclarations(
			rule.ID,
			rule.Declarations,
		)

		e.Dirty = true
		e.Revision++

		return true
	}

	return false
}

// UpdateRuleRecursive replaces the declarations of a rule anywhere in the
// working rule tree.
func (e *EditorState) UpdateRuleRecursive(
	id string,
	declarations []parser.Declaration,
) bool {
	e.m.Lock()
	defer e.m.Unlock()

	if !updateRuleDeclarationsRecursive(
		e.Rules,
		id,
		declarations,
	) {
		return false
	}

	e.syncActiveRuleDeclarations(
		id,
		declarations,
	)

	e.Dirty = true
	e.Revision++

	return true
}

// syncActiveRuleDeclarations keeps the active editor rule synchronized with
// the parser rule tree.
func (e *EditorState) syncActiveRuleDeclarations(
	id string,
	declarations []parser.Declaration,
) {
	if e.ActiveRule == nil {
		return
	}

	if e.ActiveRule.ID != id {
		return
	}

	e.ActiveRule.Body = FormatRuleDeclarations(
		declarations,
	)
}

// updateRuleDeclarationsRecursive updates declarations for a rule anywhere
// in the rule tree.
func updateRuleDeclarationsRecursive(
	rules []parser.Rule,
	id string,
	declarations []parser.Declaration,
) bool {
	for i := range rules {
		if rules[i].ID == id {
			rules[i].Declarations = CloneDeclarations(
				declarations,
			)

			return true
		}

		if updateRuleDeclarationsRecursive(
			rules[i].Children,
			id,
			declarations,
		) {
			return true
		}
	}

	return false
}

// AddRule adds a new rule to the working copy.
func (e *EditorState) AddRule(
	rule parser.Rule,
) {
	e.m.Lock()
	defer e.m.Unlock()

	e.Rules = append(
		e.Rules,
		CloneRules(
			[]parser.Rule{
				rule,
			},
		)[0],
	)

	e.Dirty = true
	e.Revision++
}

// DeleteRule deletes a top-level rule from the working copy.
func (e *EditorState) DeleteRule(
	id string,
) bool {
	e.m.Lock()
	defer e.m.Unlock()

	for i := range e.Rules {
		if e.Rules[i].ID != id {
			continue
		}

		e.Rules = append(
			e.Rules[:i],
			e.Rules[i+1:]...,
		)

		e.Dirty = true
		e.Revision++

		return true
	}

	return false
}

// DeleteRuleRecursive deletes a rule anywhere in the working rule tree.
func (e *EditorState) DeleteRuleRecursive(
	id string,
) bool {
	e.m.Lock()
	defer e.m.Unlock()

	updated, deleted := deleteRuleRecursive(
		e.Rules,
		id,
	)

	if !deleted {
		return false
	}

	e.Rules = updated

	e.Dirty = true
	e.Revision++

	return true
}

// deleteRuleRecursive returns an updated rule slice and whether a rule was
// removed.
func deleteRuleRecursive(
	rules []parser.Rule,
	id string,
) ([]parser.Rule, bool) {
	for i := range rules {
		if rules[i].ID == id {
			return append(
				rules[:i],
				rules[i+1:]...,
			), true
		}

		updated, deleted := deleteRuleRecursive(
			rules[i].Children,
			id,
		)

		if deleted {
			rules[i].Children = updated

			return rules, true
		}
	}

	return rules, false
}
