import type { PreviewUpdate } from "../../../models/skin/preview";

export function createEmptyPreviewUpdate(): PreviewUpdate {
    return {
        elements: [],
        metrics: {
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
        },
    };
}
