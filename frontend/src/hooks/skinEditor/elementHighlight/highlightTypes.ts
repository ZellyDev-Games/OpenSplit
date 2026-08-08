export type HighlightType = "selected" | "overflow" | "selected-overflow";

export interface Highlight {
    element: Element;
    type: HighlightType;
}
