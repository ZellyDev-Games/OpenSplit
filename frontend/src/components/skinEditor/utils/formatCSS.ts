export function formatCSSRule(selector: string, body: string): string {
    return `${selector} {\n${body}\n}`;
}
