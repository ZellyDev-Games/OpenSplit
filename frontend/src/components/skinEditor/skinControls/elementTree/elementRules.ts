import type { SkinCSSRule } from "../../../../models/skin/css";
import type { SkinModel } from "../../../../models/skin/editor";
import type { SkinElement } from "../../../../models/skin/element";
import { getElementSelectors } from "../utils/elementSelectors";
import { selectorMatches } from "../utils/selectorMatch";

export function getElementRules(model: SkinModel, element: SkinElement): SkinCSSRule[] {
    const selectors = getElementSelectors(element);

    return model.rules
        .filter((rule) => selectors.some((selector) => selectorMatches(rule.selector, selector)))
        .sort(compareRules);
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
