import type { RuntimeElement } from "../../../models/skin/element";
import type { PreviewMetrics } from "../../../models/skin/preview";
import { isClippedByScrollContainer } from "./previewRuntime";

const OVERFLOW_EPSILON = 1;

export function calculatePreviewMetrics(
    splitter: Element,
    runtime: RuntimeElement[],
    defaultPaddingX: number,
    defaultPaddingY: number,
): PreviewMetrics {
    const splitterBounds = splitter.getBoundingClientRect();

    const splitterWidth = splitterBounds.width;
    const splitterHeight = splitterBounds.height;

    let contentLeft = 0;
    let contentTop = 0;
    let contentRight = splitterWidth;
    let contentBottom = splitterHeight;

    let overflowLeft = 0;
    let overflowTop = 0;
    let overflowRight = splitterWidth;
    let overflowBottom = splitterHeight;

    const overflowingIds = new Set<string>();

    for (const runtimeElement of runtime) {
        if (runtimeElement.element === splitter) {
            continue;
        }

        const bounds = runtimeElement.element.getBoundingClientRect();

        const relative = {
            left: bounds.left - splitterBounds.left,
            top: bounds.top - splitterBounds.top,
            right: bounds.right - splitterBounds.left,
            bottom: bounds.bottom - splitterBounds.top,
        };

        /*
         * Track the complete bounds for overflow detection.
         */
        overflowLeft = Math.min(overflowLeft, relative.left);
        overflowTop = Math.min(overflowTop, relative.top);
        overflowRight = Math.max(overflowRight, relative.right);
        overflowBottom = Math.max(overflowBottom, relative.bottom);

        const clippedByScrollContainer = isClippedByScrollContainer(runtimeElement.element, splitter);

        const overflowing =
            relative.left < -OVERFLOW_EPSILON ||
            relative.top < -OVERFLOW_EPSILON ||
            relative.right > splitterWidth + OVERFLOW_EPSILON ||
            relative.bottom > splitterHeight + OVERFLOW_EPSILON;

        /*
         * Elements inside scroll containers may naturally extend
         * outside the visible splitter area.
         */
        if (overflowing && !clippedByScrollContainer) {
            overflowingIds.add(runtimeElement.id);
        }

        /*
         * Scroll containers should not expand the preview canvas.
         */
        if (clippedByScrollContainer) {
            continue;
        }

        /*
         * Real skin elements extending outside the splitter should
         * expand the preview canvas.
         */
        contentLeft = Math.min(contentLeft, relative.left);
        contentTop = Math.min(contentTop, relative.top);
        contentRight = Math.max(contentRight, relative.right);
        contentBottom = Math.max(contentBottom, relative.bottom);
    }

    const content = new DOMRect(contentLeft, contentTop, contentRight - contentLeft, contentBottom - contentTop);

    const leftOverflow = Math.max(0, -content.left);
    const rightOverflow = Math.max(0, content.right - splitterWidth);
    const topOverflow = Math.max(0, -content.top);
    const bottomOverflow = Math.max(0, content.bottom - splitterHeight);

    const hasCanvasOverflow =
        leftOverflow > OVERFLOW_EPSILON ||
        rightOverflow > OVERFLOW_EPSILON ||
        topOverflow > OVERFLOW_EPSILON ||
        bottomOverflow > OVERFLOW_EPSILON;

    const horizontalPadding = Math.max(defaultPaddingX, leftOverflow, rightOverflow);

    const verticalPadding = Math.max(defaultPaddingY, topOverflow, bottomOverflow);

    const paddingLeft = horizontalPadding;
    const paddingRight = horizontalPadding;
    const paddingTop = verticalPadding;
    const paddingBottom = verticalPadding;

    return {
        splitter: new DOMRect(splitterBounds.left, splitterBounds.top, splitterBounds.width, splitterBounds.height),

        content,

        canvasWidth: splitterWidth + paddingLeft + paddingRight,
        canvasHeight: splitterHeight + paddingTop + paddingBottom,

        paddingLeft,
        paddingRight,
        paddingTop,
        paddingBottom,

        overflowX: Math.max(leftOverflow, rightOverflow),
        overflowY: Math.max(topOverflow, bottomOverflow),

        splitterOffsetX: paddingLeft,
        splitterOffsetY: paddingTop,

        hasCanvasOverflow,

        hasElementOverflow: overflowingIds.size > 0,

        overflowingIds,
    };
}

export function createPreviewUpdateKey(runtime: RuntimeElement[], metrics: PreviewMetrics): string {
    return [
        runtime.map(({ id }) => id).join(","),
        [
            metrics.content.x,
            metrics.content.y,
            metrics.content.width,
            metrics.content.height,
            metrics.canvasWidth,
            metrics.canvasHeight,
            metrics.overflowX,
            metrics.overflowY,
            metrics.splitterOffsetX,
            metrics.splitterOffsetY,
            metrics.hasCanvasOverflow,
            metrics.hasElementOverflow,
            [...metrics.overflowingIds].sort().join(","),
        ].join(","),
    ].join("|");
}
