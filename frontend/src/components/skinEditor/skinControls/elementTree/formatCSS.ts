import { CSSAtRule } from "../../../../models/skin/css";

export function formatCSSRule(selector: string, body: string): string {
    return `${selector} {\n${body}\n}`;
}

export function formatCSSParents(parents: CSSAtRule[] = []): string {
    if (parents.length === 0) {
        return "";
    }
    return (
        parents
            .map((parent) => {
                switch (parent.type) {
                    case "layer":
                        return `@layer ${parent.name} {`;
                    case "media":
                        return `@media ${parent.params ?? ""} {`;
                    case "supports":
                        return `@supports ${parent.params ?? ""} {`;
                    case "font-face":
                        return "@font-face {";
                    case "import":
                        return `@import ${parent.params ?? parent.name};`;
                    default:
                        return `@${parent.name}${parent.params ? ` ${parent.params}` : ""} {`;
                }
            })
            .join("\n") + "\n"
    );
}
export function formatCSSClosingBraces(parents: CSSAtRule[] = []): string {
    return parents
        .filter((parent) => parent.type !== "import")
        .map(() => "}")
        .reverse()
        .join("\n");
}
