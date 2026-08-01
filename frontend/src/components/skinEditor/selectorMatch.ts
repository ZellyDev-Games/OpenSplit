function selectorTokens(selector: string): string[] {
    selector = selector.replaceAll(">", " ").replaceAll("+", " ").replaceAll("~", " ").replaceAll(",", " ");

    const tokens: string[] = [];

    for (const field of selector.split(/\s+/)) {
        if (!field) {
            continue;
        }

        let value = field;

        const pseudo = value.indexOf(":");

        if (pseudo >= 0) {
            value = value.slice(0, pseudo);
        }

        let current = "";

        for (const char of value) {
            if (char === "." || char === "#" || char === "[") {
                if (current) {
                    tokens.push(current);
                }

                current = char;
            } else {
                current += char;
            }
        }

        if (current) {
            tokens.push(current);
        }
    }

    return tokens;
}

export function selectorMatches(ruleSelector: string, selectedSelector: string): boolean {
    const ruleTokens = selectorTokens(ruleSelector);
    const selectedTokens = selectorTokens(selectedSelector);

    return selectedTokens.every((token) => ruleTokens.includes(token));
}
