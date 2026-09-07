import { getBoxPadding, getCSSPixelValue, getMinimum } from "./css";
import { calculateFlexMinimumSize } from "./flex";
import { calculateGridMinimumSize } from "./grid";
import { calculateSplitListMinimumSize } from "./splitList";
import type { MinimumSize } from "./types";

function isRenderable(element: HTMLElement): boolean {
    const style = getComputedStyle(element);

    return style.display !== "none" && style.visibility !== "hidden" && style.position !== "fixed";
}

function getExplicitMinimumSize(element: HTMLElement): MinimumSize {
    return {
        width: getCSSPixelValue(element, "min-width"),
        height: getCSSPixelValue(element, "min-height"),
    };
}

/**
 * Returns the minimum size declared by the OpenSplit skin component
 * variables.
 *
 * These are intentionally independent from CSS min-width/min-height.
 *
 * A skin may use:
 *
 *     min-width: 0;
 *
 * to allow an element to shrink during normal layout while still
 * declaring:
 *
 *     --splitter-segment-name-min-width: 110px;
 *
 * as the minimum size required by the component/window.
 *
 * The component variables are complete border-box dimensions.
 * Padding is therefore not added here.
 */
function getComponentMinimumSize(element: HTMLElement): MinimumSize {
    if (element.matches("#gameTitle")) {
        return {
            width: getMinimum(element, "--splitter-game-title-min-width"),
            height: getMinimum(element, "--splitter-game-title-min-height"),
        };
    }

    if (element.matches("#gameCategory")) {
        return {
            width: getMinimum(element, "--splitter-game-category-min-width"),
            height: getMinimum(element, "--splitter-game-category-min-height"),
        };
    }

    if (element.matches(".game-variable")) {
        return {
            width: getMinimum(element, "--splitter-game-variable-min-width"),
            height: getMinimum(element, "--splitter-game-variable-min-height"),
        };
    }

    if (element.matches("#attempts")) {
        return {
            width: getMinimum(element, "--splitter-attempts-min-width"),
            height: getMinimum(element, "--splitter-attempts-min-height"),
        };
    }

    if (element.matches(".segmentIcon")) {
        return {
            width: getMinimum(element, "--splitter-segment-icon-min-width"),
            height: getMinimum(element, "--splitter-segment-icon-min-height"),
        };
    }

    if (element.matches(".splitName")) {
        return {
            width: getMinimum(element, "--splitter-segment-name-min-width"),
            height: getMinimum(element, "--splitter-segment-name-min-height"),
        };
    }

    if (element.matches(".splitDelta")) {
        return {
            width: getMinimum(element, "--splitter-segment-delta-min-width"),
            height: getMinimum(element, "--splitter-segment-delta-min-height"),
        };
    }

    if (element.matches(".splitComparison")) {
        return {
            width: getMinimum(element, "--splitter-segment-comparison-min-width"),
            height: getMinimum(element, "--splitter-segment-comparison-min-height"),
        };
    }

    if (element.matches(".splitTime")) {
        return {
            width: getMinimum(element, "--splitter-segment-time-min-width"),
            height: getMinimum(element, "--splitter-segment-time-min-height"),
        };
    }

    if (element.matches("#time-sign")) {
        return {
            width: getMinimum(element, "--splitter-timer-sign-min-width"),
            height: getMinimum(element, "--splitter-timer-sign-min-height"),
        };
    }

    if (element.matches("#time-hours")) {
        return {
            width: getMinimum(element, "--splitter-timer-hours-min-width"),
            height: getMinimum(element, "--splitter-timer-hours-min-height"),
        };
    }

    if (element.matches("#time-sep-hm")) {
        return {
            width: getMinimum(element, "--splitter-timer-separator-hours-minutes-min-width"),
            height: getMinimum(element, "--splitter-timer-separator-hours-min-height"),
        };
    }

    if (element.matches("#time-minutes")) {
        return {
            width: getMinimum(element, "--splitter-timer-minutes-min-width"),
            height: getMinimum(element, "--splitter-timer-minutes-min-height"),
        };
    }

    if (element.matches("#time-sep-ms")) {
        return {
            width: getMinimum(element, "--splitter-timer-separator-minutes-seconds-min-width"),
            height: getMinimum(element, "--splitter-timer-separator-minutes-seconds-min-height"),
        };
    }

    if (element.matches("#time-seconds")) {
        return {
            width: getMinimum(element, "--splitter-timer-seconds-min-width"),
            height: getMinimum(element, "--splitter-timer-seconds-min-height"),
        };
    }

    if (element.matches("#time-sep-sc")) {
        return {
            width: getMinimum(element, "--splitter-timer-separator-seconds-centis-min-width"),
            height: getMinimum(element, "--splitter-timer-separator-seconds-centis-min-height"),
        };
    }

    if (element.matches("#time-centis")) {
        return {
            width: getMinimum(element, "--splitter-timer-centis-min-width"),
            height: getMinimum(element, "--splitter-timer-centis-min-height"),
        };
    }

    if (element.matches(".comparison-mode")) {
        return {
            width: getMinimum(element, "--splitter-comparison-min-width"),
            height: getMinimum(element, "--splitter-comparison-min-height"),
        };
    }

    if (element.matches(".world-record-player")) {
        return {
            width: getMinimum(element, "--splitter-world-record-player-min-width"),
            height: getMinimum(element, "--splitter-world-record-player-min-height"),
        };
    }

    if (element.matches(".world-record-real-time")) {
        return {
            width: getMinimum(element, "--splitter-world-record-real-time-min-width"),
            height: getMinimum(element, "--splitter-world-record-real-time-min-height"),
        };
    }

    if (element.matches(".world-record-in-game-time")) {
        return {
            width: getMinimum(element, "--splitter-world-record-in-game-time-min-width"),
            height: getMinimum(element, "--splitter-world-record-in-game-time-min-height"),
        };
    }

    return {
        width: 0,
        height: 0,
    };
}

export function calculateElementMinimumSize(element: HTMLElement): MinimumSize {
    if (!isRenderable(element)) {
        return {
            width: 0,
            height: 0,
        };
    }

    if (element.matches("#splitList")) {
        return calculateSplitListMinimumSize(element);
    }

    const style = getComputedStyle(element);

    const children = Array.from(element.children).filter(
        (child): child is HTMLElement => child instanceof HTMLElement && isRenderable(child),
    );

    const childSizes = children.map((child) => ({
        element: child,
        size: calculateElementMinimumSize(child),
    }));

    let layoutSize: MinimumSize;

    switch (style.display) {
        case "flex":
        case "inline-flex":
            layoutSize = calculateFlexMinimumSize(
                element,
                childSizes.map((child) => child.size),
            );
            break;

        case "grid":
        case "inline-grid":
            layoutSize = calculateGridMinimumSize(element, childSizes);
            break;

        default: {
            const padding = getBoxPadding(element);

            layoutSize = {
                width: Math.max(...childSizes.map((child) => child.size.width), 0) + padding.left + padding.right,

                height: Math.max(...childSizes.map((child) => child.size.height), 0) + padding.top + padding.bottom,
            };

            break;
        }
    }

    const explicitMinimum = getExplicitMinimumSize(element);
    const componentMinimum = getComponentMinimumSize(element);

    return {
        width: Math.max(layoutSize.width, explicitMinimum.width, componentMinimum.width),

        height: Math.max(layoutSize.height, explicitMinimum.height, componentMinimum.height),
    };
}
