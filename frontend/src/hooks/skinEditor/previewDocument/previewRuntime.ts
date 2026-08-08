import { RuntimeElement, SkinElement } from "../../../models/skin/element";

export function resolveRuntimeElements(document: Document, elements: SkinElement[]): RuntimeElement[] {
    const runtime: RuntimeElement[] = [];

    for (const element of elements) {
        try {
            const match = document.querySelector(element.selector);

            if (!match) {
                continue;
            }

            runtime.push({
                ...element,
                element: match,
            });
        } catch {
            // Ignore invalid selectors. A malformed skin selector should not
            // prevent the rest of the preview from updating.
        }
    }

    return runtime;
}

export function isClippedByScrollContainer(element: Element, root: Element): boolean {
    let current = element.parentElement;

    while (current && current !== root) {
        if (current.scrollHeight > current.clientHeight || current.scrollWidth > current.clientWidth) {
            return true;
        }

        current = current.parentElement;
    }

    return false;
}
