export function getCSSPixelValue(element: HTMLElement, property: string, fallback = 0): number {
    const value = getComputedStyle(element).getPropertyValue(property).trim();

    if (!value) {
        return fallback;
    }

    const pixels = Number.parseFloat(value);

    return Number.isFinite(pixels) ? pixels : fallback;
}

export function getCSSVariablePixelValue(element: HTMLElement, variable: string, fallback = 0): number {
    const value = getComputedStyle(element).getPropertyValue(variable).trim();

    if (!value) {
        return fallback;
    }

    const pixels = Number.parseFloat(value);

    return Number.isFinite(pixels) ? pixels : fallback;
}

/**
 * Returns a skin-declared component minimum.
 *
 * The component minimum variables are authoritative for the
 * splitter minimum-size calculation. They are independent of
 * CSS min-width/min-height because skins intentionally use
 * min-width/min-height to control normal layout/shrinking.
 *
 * There is deliberately no fallback from width to min-height or
 * from height to min-width.
 */
export function getMinimum(element: HTMLElement, variable: string): number {
    return getCSSVariablePixelValue(element, variable, 0);
}

export function getBoxPadding(element: HTMLElement) {
    const style = getComputedStyle(element);

    return {
        left: parseFloat(style.paddingLeft) || 0,
        right: parseFloat(style.paddingRight) || 0,
        top: parseFloat(style.paddingTop) || 0,
        bottom: parseFloat(style.paddingBottom) || 0,
    };
}

export function getBoxBorder(element: HTMLElement) {
    const style = getComputedStyle(element);

    return {
        left: parseFloat(style.borderLeftWidth) || 0,
        right: parseFloat(style.borderRightWidth) || 0,
        top: parseFloat(style.borderTopWidth) || 0,
        bottom: parseFloat(style.borderBottomWidth) || 0,
    };
}

export function getGap(element: HTMLElement) {
    const style = getComputedStyle(element);

    return {
        row: parseFloat(style.rowGap) || 0,
        column: parseFloat(style.columnGap) || 0,
    };
}
