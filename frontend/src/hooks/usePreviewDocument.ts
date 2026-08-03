import { useEffect, useRef } from "react";

import { PreviewMetrics, PreviewUpdate, RuntimeElement, SkinElement } from "../models/skinModel";

interface Props {
    document: Document | null;

    elements: SkinElement[];

    defaultPaddingX: number;

    defaultPaddingY: number;

    callback?: (update: PreviewUpdate) => void;
}

export function usePreviewDocument({ document, elements, defaultPaddingX, defaultPaddingY, callback }: Props) {
    const previous = useRef<string>("");

    useEffect(() => {
        if (!document || !callback) {
            return;
        }

        let frame = 0;

        const resizeObserver = new ResizeObserver(() => {
            update();
        });

        const update = () => {
            cancelAnimationFrame(frame);

            frame = requestAnimationFrame(() => {
                const splitter = document.getElementById("splitter");

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
                        continue;
                    }
                }

                let metrics: PreviewMetrics = {
                    splitter: new DOMRect(),

                    content: new DOMRect(),

                    canvasWidth: 0,

                    canvasHeight: 0,

                    paddingLeft: 0,

                    paddingRight: 0,

                    paddingTop: 0,

                    paddingBottom: 0,

                    overflowX: 0,

                    overflowY: 0,

                    splitterOffsetX: 0,

                    splitterOffsetY: 0,

                    hasCanvasOverflow: false,

                    hasElementOverflow: false,

                    overflowingElements: [],
                };

                if (splitter) {
                    const splitterBounds = splitter.getBoundingClientRect();

                    /*
                     * Content coordinates are relative to splitter.
                     *
                     * These values intentionally allow negative x/y.
                     */
                    let contentLeft = 0;
                    let contentTop = 0;
                    let contentRight = splitterBounds.width;
                    let contentBottom = splitterBounds.height;

                    const splitterWidth = splitterBounds.width;
                    const splitterHeight = splitterBounds.height;

                    let overflowLeft = 0;
                    let overflowTop = 0;
                    let overflowRight = splitterWidth;
                    let overflowBottom = splitterHeight;

                    const OVERFLOW_EPSILON = 1;

                    const overflowingElements = new Set<string>();

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

                        overflowLeft = Math.min(overflowLeft, relative.left);

                        overflowTop = Math.min(overflowTop, relative.top);

                        overflowRight = Math.max(overflowRight, relative.right);

                        overflowBottom = Math.max(overflowBottom, relative.bottom);

                        /*
                         * Scroll containers should not expand the preview canvas.
                         *
                         * Example:
                         *   #splitContainer
                         *       hidden segment rowsgg
                         *
                         * Those rows may be outside the visible splitter area but
                         * should not make the iframe larger.
                         */
                        const clippedByScrollContainer = isClippedByScrollContainer(runtimeElement.element, splitter);

                        const overflow =
                            relative.left < -OVERFLOW_EPSILON ||
                            relative.top < -OVERFLOW_EPSILON ||
                            relative.right > splitterBounds.width + OVERFLOW_EPSILON ||
                            relative.bottom > splitterBounds.height + OVERFLOW_EPSILON;

                        /*
                         * Only report visible overflow.
                         *
                         * Elements inside scroll containers (#splitContainer, tables,
                         * segment lists, etc.) are expected to leave the viewport.
                         */
                        if (overflow && !clippedByScrollContainer) {
                            overflowingElements.add(runtimeElement.id);
                        }

                        /*
                         * Scroll containers should not contribute to canvas sizing.
                         */
                        if (clippedByScrollContainer) {
                            continue;
                        }

                        /*
                         * Content sizing should still include the element bounds.
                         *
                         * This allows real skin elements extending beyond the splitter
                         * to expand the preview canvas.
                         */
                        contentLeft = Math.min(contentLeft, relative.left);

                        contentTop = Math.min(contentTop, relative.top);

                        contentRight = Math.max(contentRight, relative.right);

                        contentBottom = Math.max(contentBottom, relative.bottom);
                    }
                    const content = new DOMRect(
                        overflowLeft,
                        overflowTop,
                        overflowRight - overflowLeft,
                        overflowBottom - overflowTop,
                    );

                    const leftOverflow = Math.max(0, -content.left);

                    const rightOverflow = Math.max(0, content.right - splitterWidth);

                    const topOverflow = Math.max(0, -content.top);

                    const bottomOverflow = Math.max(0, content.bottom - splitterHeight);

                    /*
                     * The canvas can be larger than the splitter because it contains
                     * centering padding. This is independent from whether any skin
                     * element is actually overflowing.
                     *
                     * This controls initial workspace scrolling.
                     */
                    const hasCanvasOverflow =
                        leftOverflow > OVERFLOW_EPSILON ||
                        rightOverflow > OVERFLOW_EPSILON ||
                        topOverflow > OVERFLOW_EPSILON ||
                        bottomOverflow > OVERFLOW_EPSILON;

                    /*
                     * Default centered padding from workspace size.
                     *
                     * Example:
                     *
                     * workspace 468px
                     * splitter  320px
                     *
                     * padding = (468 - 320) / 2 = 74px
                     *
                     * This padding must remain mirrored.
                     */

                    const horizontalPadding = Math.max(defaultPaddingX, leftOverflow, rightOverflow);

                    const verticalPadding = Math.max(defaultPaddingY, topOverflow, bottomOverflow);

                    const paddingLeft = horizontalPadding;

                    const paddingRight = horizontalPadding;

                    const paddingTop = verticalPadding;

                    const paddingBottom = verticalPadding;

                    const canvasWidth = splitterWidth + paddingLeft + paddingRight;

                    const canvasHeight = splitterHeight + paddingTop + paddingBottom;

                    const splitterOffsetX = paddingLeft;

                    const splitterOffsetY = paddingTop;

                    metrics = {
                        splitter: new DOMRect(
                            splitterBounds.left,
                            splitterBounds.top,
                            splitterBounds.width,
                            splitterBounds.height,
                        ),

                        content,

                        canvasWidth,

                        canvasHeight,

                        paddingLeft,

                        paddingRight,

                        paddingTop,

                        paddingBottom,

                        overflowX: Math.max(leftOverflow, rightOverflow),

                        overflowY: Math.max(topOverflow, bottomOverflow),

                        splitterOffsetX,

                        splitterOffsetY,

                        hasCanvasOverflow,

                        hasElementOverflow: overflowingElements.size > 0,

                        overflowingElements: [...overflowingElements],
                    };
                }

                const key = [
                    runtime.map((element) => element.id).join(","),
                    [
                        metrics.content.x,
                        metrics.content.y,
                        metrics.content.width,
                        metrics.content.height,
                        metrics.canvasWidth,
                        metrics.canvasHeight,
                        metrics.overflowingElements,
                    ].join(","),
                ].join("|");

                if (key === previous.current) {
                    return;
                }

                previous.current = key;

                for (const element of runtime) {
                    resizeObserver.observe(element.element);
                }

                callback({
                    elements: runtime,
                    metrics,
                });
            });
        };

        update();

        const mutationObserver = new MutationObserver(() => {
            update();
        });

        const splitter = document.getElementById("splitter");

        if (splitter) {
            mutationObserver.observe(splitter, {
                childList: true,
                subtree: true,
                attributes: true,
            });
        }

        return () => {
            cancelAnimationFrame(frame);

            mutationObserver.disconnect();

            resizeObserver.disconnect();
        };
    }, [document, elements, callback]);
}

function isClippedByScrollContainer(element: Element, root: Element): boolean {
    let current = element.parentElement;

    while (current && current !== root) {
        if (current.scrollHeight > current.clientHeight || current.scrollWidth > current.clientWidth) {
            return true;
        }

        current = current.parentElement;
    }

    return false;
}
