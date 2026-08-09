package parser

// FindRule returns the rule with the supplied ID.
//
// The search is recursive and includes rules nested inside at-rules.
func FindRule(
	rules []Rule,
	id string,
) (*Rule, bool) {
	for index := range rules {
		rule := &rules[index]

		if rule.ID == id {
			return rule, true
		}

		if found, ok := FindRule(
			rule.Children,
			id,
		); ok {
			return found, true
		}
	}

	return nil, false
}

// ReplaceRule replaces the rule with the supplied ID.
//
// The replacement is applied recursively to nested rules.
// It returns true when a rule was replaced.
func ReplaceRule(
	rules []Rule,
	replacement Rule,
) bool {
	for index := range rules {
		if rules[index].ID == replacement.ID {
			rules[index] = replacement

			return true
		}

		if ReplaceRule(
			rules[index].Children,
			replacement,
		) {
			return true
		}
	}

	return false
}

// DeleteRule removes the rule with the supplied ID.
//
// The operation searches recursively through nested rule trees.
// It returns the updated slice and whether a rule was removed.
func DeleteRule(
	rules []Rule,
	id string,
) ([]Rule, bool) {
	for index := range rules {
		if rules[index].ID == id {
			return removeRuleAt(
				rules,
				index,
			), true
		}

		children, deleted := DeleteRule(
			rules[index].Children,
			id,
		)

		if !deleted {
			continue
		}

		rules[index].Children = children

		return rules, true
	}

	return rules, false
}

// removeRuleAt removes one rule from a sibling slice.
//
// The three-index slice expression keeps the backing array from retaining
// the removed rule in the unused portion of the slice.
func removeRuleAt(
	rules []Rule,
	index int,
) []Rule {
	return append(
		rules[:index:index],
		rules[index+1:]...,
	)
}

// Walk visits every rule in the tree depth-first.
//
// The callback receives a pointer to the rule in the supplied tree, allowing
// callers to modify the rule in place.
func Walk(
	rules []Rule,
	fn func(*Rule),
) {
	for index := range rules {
		rule := &rules[index]

		fn(rule)

		Walk(
			rule.Children,
			fn,
		)
	}
}
