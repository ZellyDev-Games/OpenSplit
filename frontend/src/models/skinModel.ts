import { CSSRuleEditor } from "./cssProperty";

export interface CSSFile {
    name: string;
    path: string;
    contents: string;
    originalContents?: string;
    text: boolean;
    url: string;
    type: string;
}

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

export interface CSSAtRule {
    type: "import" | "media" | "supports" | "font-face" | "layer" | "unknown";

    name: string;

    params?: string;
}

/**
 * Metadata describing an editable skin element.
 *
 * This is the canonical element definition.
 */
export interface SkinElement {
    /**
     * Stable identifier.
     */
    id: string;

    /**
     * Human readable name.
     */
    label: string;

    /**
     * Selector used to locate the runtime element.
     */
    selector: string;

    description?: string;

    file?: string;

    line?: number;

    layer?: string;
}

/**
 * Skin element bound to a live DOM node.
 */
export interface RuntimeElement extends SkinElement {
    element: Element;
}

export interface PreviewMetrics {
    splitter: DOMRect;

    content: DOMRect;

    canvasWidth: number;

    canvasHeight: number;

    /**
     * Mirrored canvas padding around splitter.
     */
    paddingLeft: number;

    paddingRight: number;

    paddingTop: number;

    paddingBottom: number;

    /**
     * Maximum overflow beyond splitter bounds.
     */
    overflowX: number;

    overflowY: number;

    /**
     * Actual splitter position inside canvas.
     */
    splitterOffsetX: number;

    splitterOffsetY: number;

    hasOverflow: boolean;

    overflowingElements: string[];
}

export interface PreviewUpdate {
    elements: RuntimeElement[];
    metrics: PreviewMetrics;
}

export interface SkinEditorTarget {
    elementId: string | null;

    file: string | null;

    selector: string | null;

    ruleId: string | null;

    mode: "file" | "existing" | "create";
}

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
