import type { CSSAtRule, SkinCSSRule } from "../../../models/skin/css";

export function collectRuntimeCSS(): SkinCSSRule[] {
    const rules: SkinCSSRule[] = [];

    for (const sheet of Array.from(document.styleSheets)) {
        let cssRules: CSSRuleList;

        try {
            cssRules = sheet.cssRules;
        } catch {
            // Cross-origin stylesheet.
            continue;
        }

        collectRules(cssRules, rules, sheet.href ?? "runtime", []);
    }

    return rules.map((rule, index) => ({
        ...rule,
        order: index,
        line: 0,
    }));
}

function collectRules(cssRules: CSSRuleList, output: SkinCSSRule[], source: string, parents: CSSAtRule[]) {
    for (const rule of Array.from(cssRules)) {
        const atRule = getAtRule(rule);

        if (atRule && hasChildren(rule)) {
            collectRules(rule.cssRules, output, source, [...parents, atRule]);

            continue;
        }

        if (!(rule instanceof CSSStyleRule)) {
            continue;
        }

        const selectors = rule.selectorText
            .split(",")
            .map((selector) => selector.trim())
            .filter(Boolean);

        for (const selector of selectors) {
            output.push({
                id: `${source}:${output.length}`,

                file: source,

                layer: [...parents].reverse().find((p) => p.type === "layer")?.name ?? "",

                selector,

                body: formatStyleDeclaration(rule.style),

                parentId: "",

                parents,

                line: 0,

                order: output.length,
            });
        }
    }
}

function hasChildren(rule: CSSRule): rule is CSSGroupingRule {
    return "cssRules" in rule;
}

function getAtRule(rule: CSSRule): CSSAtRule | null {
    if (typeof CSSLayerBlockRule !== "undefined" && rule instanceof CSSLayerBlockRule) {
        return {
            type: "layer",
            name: rule.name,
        };
    }

    if (rule instanceof CSSMediaRule) {
        return {
            type: "media",
            name: "media",
            params: rule.conditionText,
        };
    }

    if (typeof CSSSupportsRule !== "undefined" && rule instanceof CSSSupportsRule) {
        return {
            type: "supports",
            name: "supports",
            params: rule.conditionText,
        };
    }

    if (typeof CSSImportRule !== "undefined" && rule instanceof CSSImportRule) {
        return {
            type: "import",
            name: "import",
            params: rule.href,
        };
    }

    if (rule instanceof CSSFontFaceRule) {
        return {
            type: "font-face",
            name: "font-face",
        };
    }

    return null;
}

function formatStyleDeclaration(style: CSSStyleDeclaration): string {
    const lines: string[] = [];

    for (let i = 0; i < style.length; i++) {
        const property = style.item(i);

        const value = style.getPropertyValue(property);

        const priority = style.getPropertyPriority(property);

        lines.push(`${property}: ${value}${priority ? ` !${priority}` : ""};`);
    }

    return lines.join("\n");
}
