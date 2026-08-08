import type { SkinCSSRule } from "../../models/skin/css";

export function flattenRules(rules: SkinCSSRule[]): SkinCSSRule[] {
    const result: SkinCSSRule[] = [];

    const walk = (items: SkinCSSRule[]): void => {
        for (const rule of items) {
            if (rule.selector) {
                result.push(rule);
            }

            if (rule.children?.length) {
                walk(rule.children);
            }
        }
    };

    walk(rules);

    return result;
}
