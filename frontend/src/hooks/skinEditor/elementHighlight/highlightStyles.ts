import type { HighlightType } from "./highlightTypes";

interface HighlightStyle {
    border: string;
    background: string;
    outline?: string;
    outlineOffset?: string;
}

const HIGHLIGHT_STYLES: Record<HighlightType, HighlightStyle> = {
    selected: {
        border: "3px solid #00b7ff",
        background: "rgba(0,183,255,0.20)",
    },

    overflow: {
        border: "3px solid #ff3333",
        background: "rgba(255,0,0,0.15)",
    },

    "selected-overflow": {
        border: "3px solid #ff3333",
        outline: "3px solid #00b7ff",
        outlineOffset: "-6px",
        background: "rgba(255,0,0,0.20)",
    },
};

export function getHighlightStyle(type: HighlightType): HighlightStyle {
    return HIGHLIGHT_STYLES[type];
}
