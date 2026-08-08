import { CSSRuleEditor } from "./skin/css";

/**
 * CSS file loaded by the skin editor.
 */
export interface CSSFile {
    name: string;
    path: string;
    contents: string;
    originalContents?: string;
    text: boolean;
    url: string;
    type: string;
}

/**
 * Parsed CSS rule.
 */
export interface CSSRule {
    id: string;
    file: string;
    layer: string;
    selector: string;
    body: string;
    line: number;
    order: number;
    children?: CSSRule[];
    parents?: CSSAtRule[];
}

/**
 * CSS at-rule associated with a CSS rule.
 */
export interface CSSAtRule {
    type: "import" | "media" | "supports" | "font-face" | "layer" | "unknown";
    name: string;
    params?: string;
}

/**
 * Skin element exposed to the skin editor.
 */
export interface SkinElement {
    id: string;
    label: string;
    selector: string;
    description?: string;
    file?: string;
    line?: number;
    layer?: string;
}

/**
 * Skin element associated with a live DOM element in the preview.
 */
export interface RuntimeElement extends SkinElement {
    element: Element;
}

/**
 * Layout and overflow measurements for the skin preview.
 */
export interface PreviewMetrics {
    splitter: DOMRect;
    content: DOMRect;
    canvasWidth: number;
    canvasHeight: number;
    paddingLeft: number;
    paddingRight: number;
    paddingTop: number;
    paddingBottom: number;
    overflowX: number;
    overflowY: number;
    splitterOffsetX: number;
    splitterOffsetY: number;
    hasCanvasOverflow: boolean;
    hasElementOverflow: boolean;
    overflowingElements: string[];
}

/**
 * Current state of the skin preview.
 */
export interface PreviewUpdate {
    elements: RuntimeElement[];
    metrics: PreviewMetrics;
}

/**
 * Current target of the skin editor.
 */
export interface SkinEditorTarget {
    elementId: string | null;
    file: string | null;
    selector: string | null;
    ruleId: string | null;
    mode: "file" | "existing" | "create";
}

/**
 * Complete skin editor state.
 */
export default interface SkinModel {
    name: string;
    skins: string[];
    directory: string;
    styleSheet: string;
    files: CSSFile[];
    rules: CSSRule[];
    elements: SkinElement[];
    target: SkinEditorTarget;
    activeRule: CSSRuleEditor | null;
    dirty: boolean;
    revision: number;
}
