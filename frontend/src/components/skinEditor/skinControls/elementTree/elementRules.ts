import type { SkinCSSRule } from "../../../../models/skin/css";
import type { SkinModel } from "../../../../models/skin/editor";
import type { SkinElement } from "../../../../models/skin/element";
import { selectorMatches } from "../utils/selectorMatch";

export function getElementRules(model: SkinModel, element: SkinElement): SkinCSSRule[] {
    return model.rules.filter((rule) => selectorMatches(rule.selector, element.selector)).sort(compareRules);
}

function compareRules(a: SkinCSSRule, b: SkinCSSRule): number {
    if (a.file !== b.file) {
        return a.file.localeCompare(b.file);
    }

    if (a.line !== b.line) {
        return a.line - b.line;
    }

    return a.order - b.order;
}
