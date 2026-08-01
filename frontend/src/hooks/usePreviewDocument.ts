import { useEffect, useRef } from "react";

import { PreviewMetrics, PreviewUpdate, RuntimeElement, SkinElement } from "../models/skinModel";

interface Props {
    document: Document | null;

    elements: SkinElement[];

    callback?: (update: PreviewUpdate) => void;
}

export function usePreviewDocument({ document, elements, callback }: Props) {
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

                const app = document.getElementById("App");

                const previousTransform = app?.style.transform ?? "";

                if (app) {
                    app.style.transform = "none";
                }

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

                    hasOverflow: false,
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

                    const overflowingElements = new Set<string>();

                    for (const runtimeElement of runtime) {
                        const bounds = runtimeElement.element.getBoundingClientRect();

                        const relative = {
                            left: bounds.left - splitterBounds.left,
                            top: bounds.top - splitterBounds.top,
                            right: bounds.right - splitterBounds.left,
                            bottom: bounds.bottom - splitterBounds.top,
                        };

                        const overflow =
                            relative.left < 0 ||
                            relative.top < 0 ||
                            relative.right > splitterBounds.width ||
                            relative.bottom > splitterBounds.height;

                        if (overflow) {
                            overflowingElements.add(runtimeElement.id);
                        }

                        contentLeft = Math.min(contentLeft, relative.left);

                        contentTop = Math.min(contentTop, relative.top);

                        contentRight = Math.max(contentRight, relative.right);

                        contentBottom = Math.max(contentBottom, relative.bottom);
                    }

                    const content = new DOMRect(
                        contentLeft,
                        contentTop,
                        contentRight - contentLeft,
                        contentBottom - contentTop,
                    );

                    updateDocumentBounds(document, content);

                    metrics = {
                        splitter: new DOMRect(
                            splitterBounds.left,
                            splitterBounds.top,
                            splitterBounds.width,
                            splitterBounds.height,
                        ),

                        content,

                        hasOverflow: overflowingElements.size > 0,

                        overflowingElements: [...overflowingElements],
                    };
                }

                /*
                 * Restore translation after measuring.
                 */
                if (app) {
                    app.style.transform = previousTransform;
                }

                const key = [
                    runtime.map((element) => element.id).join(","),

                    [metrics.content.x, metrics.content.y, metrics.content.width, metrics.content.height].join(","),
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
            });
        }

        return () => {
            cancelAnimationFrame(frame);

            mutationObserver.disconnect();

            resizeObserver.disconnect();
        };
    }, [document, elements, callback]);
}

function updateDocumentBounds(document: Document, content: DOMRect) {
    const app = document.getElementById("App");

    if (!app) {
        return;
    }

    /*
     * Only translate the rendered content.
     *
     * The iframe viewport remains fixed.
     * This prevents ResizeObserver feedback loops.
     */
    const offsetX = Math.max(0, -content.x);

    const offsetY = Math.max(0, -content.y);

    app.style.transform = `translate(${offsetX}px, ${offsetY}px)`;
}
