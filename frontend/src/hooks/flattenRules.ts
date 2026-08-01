import { CSSRule } from "../models/skinModel";

export function flattenRules(rules: CSSRule[]): CSSRule[] {
    const result: CSSRule[] = [];

    function walk(items: CSSRule[]) {
        for (const rule of items) {
            if (rule.selector) {
                result.push(rule);
            }

            if (rule.children) {
                walk(rule.children);
            }
        }
    }

    walk(rules);

    return result;
}
