import type { MinimumSize } from "./types";

export function getContentMinimumSize(element: HTMLElement, axes: { width?: boolean; height?: boolean }): MinimumSize {
    const style = getComputedStyle(element);

    const borderLeft = parseFloat(style.borderLeftWidth) || 0;
    const borderRight = parseFloat(style.borderRightWidth) || 0;
    const borderTop = parseFloat(style.borderTopWidth) || 0;
    const borderBottom = parseFloat(style.borderBottomWidth) || 0;

    return {
        width: axes.width ? element.scrollWidth + borderLeft + borderRight : 0,
        height: axes.height ? element.scrollHeight + borderTop + borderBottom : 0,
    };
}

export function getEffectiveMinimumSize(
    element: HTMLElement,
    widthVariable: string,
    heightVariable: string,
    contentAware?: { width?: boolean; height?: boolean },
): MinimumSize {
    const declared = {
        width: getMinimum(element, widthVariable),
        height: getMinimum(element, heightVariable),
    };

    const content = contentAware ? getContentMinimumSize(element, contentAware) : { width: 0, height: 0 };

    return {
        width: contentAware?.width ? Math.max(declared.width, content.width) : declared.width,
        height: contentAware?.height ? Math.max(declared.height, content.height) : declared.height,
    };
}

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
