import { useEffect, useRef } from "react";

import type { SkinElement } from "../../models/skin/element";
import type { PreviewMetrics, PreviewUpdate } from "../../models/skin/preview";
import { calculatePreviewMetrics, createPreviewUpdateKey } from "./previewDocument/previewMetrics";
import { resolveRuntimeElements } from "./previewDocument/previewRuntime";

interface PreviewDocumentOptions {
    document: Document | null;
    elements: SkinElement[];
    defaultPaddingX: number;
    defaultPaddingY: number;
    callback?: (update: PreviewUpdate) => void;
}

function createEmptyMetrics(): PreviewMetrics {
    return {
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

        overflowingIds: new Set<string>(),
    };
}

export function usePreviewDocument({
    document: previewDocument,
    elements,
    defaultPaddingX,
    defaultPaddingY,
    callback,
}: PreviewDocumentOptions): void {
    const previousKey = useRef("");

    useEffect(() => {
        if (!previewDocument || !callback) {
            return;
        }

        /*
         * Capture the narrowed values so nested callbacks/functions
         * retain their non-null types.
         */
        const document = previewDocument;
        const onUpdate = callback;

        let frame = 0;

        const resizeObserver = new ResizeObserver(scheduleUpdate);
        const mutationObserver = new MutationObserver(scheduleUpdate);

        function scheduleUpdate(): void {
            cancelAnimationFrame(frame);

            frame = requestAnimationFrame(update);
        }

        function update(): void {
            const splitter = document.getElementById("splitter");
            const runtime = resolveRuntimeElements(document, elements);

            const metrics = splitter
                ? calculatePreviewMetrics(splitter, runtime, defaultPaddingX, defaultPaddingY)
                : createEmptyMetrics();

            /*
             * Runtime elements can be replaced by React while retaining
             * their skin element IDs, so refresh the ResizeObserver
             * targets on every update.
             */
            resizeObserver.disconnect();

            for (const runtimeElement of runtime) {
                resizeObserver.observe(runtimeElement.element);
            }

            const key = createPreviewUpdateKey(runtime, metrics);

            if (key === previousKey.current) {
                return;
            }

            previousKey.current = key;

            onUpdate({
                elements: runtime,
                metrics,
            });
        }

        const splitter = document.getElementById("splitter");

        if (splitter) {
            mutationObserver.observe(splitter, {
                childList: true,
                subtree: true,
                attributes: true,
            });
        }

        scheduleUpdate();

        return () => {
            cancelAnimationFrame(frame);

            mutationObserver.disconnect();
            resizeObserver.disconnect();
        };
    }, [previewDocument, elements, callback, defaultPaddingX, defaultPaddingY]);
}
