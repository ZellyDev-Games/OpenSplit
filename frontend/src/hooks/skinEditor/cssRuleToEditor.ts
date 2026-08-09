import { CSSRuleEditor, SkinCSSRule } from "../../models/skin/css";

/**
 * Convert a parsed CSS rule into the editable rule model used by the editor.
 */
export function cssRuleToEditor(rule: SkinCSSRule): CSSRuleEditor {
    return {
        id: rule.id,
        originalFile: rule.file,
        originalId: rule.id,
        file: rule.file,
        selector: rule.selector,
        layer: rule.layer,
        body: rule.body,
        parentId: rule.parentId,
        parents: rule.parents ?? [],
    };
}
