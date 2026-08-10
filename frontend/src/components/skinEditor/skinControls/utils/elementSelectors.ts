import type { SkinElement } from "../../../../models/skin/element";

/**
 * Returns all CSS selectors that belong to a preview element.
 *
 * Non-stateful elements have one selector.
 *
 * Stateful elements support the three runtime states:
 *
 *   <selector>
 *   <selector>.complete
 *   <selector>.complete.pb
 *
 * The application never produces a `pb` state without `complete`.
 */
export function getElementSelectors(element: SkinElement): string[] {
    if (!element.stateful) {
        return [element.selector];
    }

    return [element.selector, `${element.selector}.complete`, `${element.selector}.complete.pb`];
}
