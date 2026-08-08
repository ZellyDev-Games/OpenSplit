import { useLayoutEffect } from "react";

import type { PreviewMetrics } from "../../../models/skin/preview";

interface PreviewScrollOptions {
    workspace: HTMLDivElement | null;
    metrics: PreviewMetrics;

    defaultPaddingX: number;
    defaultPaddingY: number;
}

export function usePreviewScroll({ workspace, metrics, defaultPaddingX, defaultPaddingY }: PreviewScrollOptions): void {
    useLayoutEffect(() => {
        if (!workspace) {
            return;
        }

        if (!metrics.hasCanvasOverflow && !metrics.hasElementOverflow) {
            workspace.scrollLeft = 0;
            workspace.scrollTop = 0;

            return;
        }

        workspace.scrollLeft = metrics.paddingLeft - defaultPaddingX;
        workspace.scrollTop = metrics.paddingTop - defaultPaddingY;
    }, [
        workspace,
        metrics.paddingLeft,
        metrics.paddingTop,
        metrics.hasCanvasOverflow,
        metrics.hasElementOverflow,
        defaultPaddingX,
        defaultPaddingY,
    ]);
}
