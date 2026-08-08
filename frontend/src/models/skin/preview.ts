import type { RuntimeElement } from "./element";

export interface PreviewMetrics {
    splitter: DOMRect;
    content: DOMRect;

    canvasWidth: number;
    canvasHeight: number;

    /**
     * Canvas padding required around the splitter.
     */
    paddingLeft: number;
    paddingRight: number;
    paddingTop: number;
    paddingBottom: number;

    /**
     * Maximum overflow outside the splitter.
     */
    overflowX: number;
    overflowY: number;

    /**
     * Splitter position inside the preview canvas.
     */
    splitterOffsetX: number;
    splitterOffsetY: number;

    hasCanvasOverflow: boolean;
    hasElementOverflow: boolean;

    /**
     * Runtime ids of overflowing elements.
     */
    overflowingIds: Set<string>;
}

export interface PreviewUpdate {
    elements: RuntimeElement[];
    metrics: PreviewMetrics;
}
