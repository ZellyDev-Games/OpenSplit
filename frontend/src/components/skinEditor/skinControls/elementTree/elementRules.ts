import type { SkinCSSRule } from "../../../../models/skin/css";
import type { SkinModel } from "../../../../models/skin/editor";
import type { SkinElement } from "../../../../models/skin/element";
import { selectorMatches } from "../utils/selectorMatch";

export function getElementRules(model: SkinModel, element: SkinElement): SkinCSSRule[] {
    return model.rules.filter((rule) => selectorMatches(rule.selector, element.selector)).sort(compareRules);
}

/**
 * Returns only rules that originate from the actual skin files.
 *
 * Runtime rules describe the CSS currently present in the preview. They are
 * useful for showing what affects an element, but they must not cause an
 * element to be considered styled by the skin.
 */
export function getSkinElementRules(model: SkinModel, element: SkinElement): SkinCSSRule[] {
    return getElementRules(model, element).filter((rule) => rule.file !== "runtime");
}

function compareRules(a: SkinCSSRule, b: SkinCSSRule): number {
    const aRuntime = a.file === "runtime";
    const bRuntime = b.file === "runtime";

    /*
     * Runtime rules are generated from the current preview and are
     * not editable source-file rules. Keep them after all skin rules.
     */
    if (aRuntime !== bRuntime) {
        return aRuntime ? 1 : -1;
    }

    if (a.file !== b.file) {
        return a.file.localeCompare(b.file);
    }

    if (a.line !== b.line) {
        return a.line - b.line;
    }

    return a.order - b.order;
}
