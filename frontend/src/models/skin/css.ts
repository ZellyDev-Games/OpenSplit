export interface CSSFile {
    name: string;
    path: string;
    contents: string;
    originalContents?: string;
    text: boolean;
    url: string;
    type: string;
}

export interface CSSAtRule {
    type: "import" | "media" | "supports" | "font-face" | "layer" | "unknown";
    name: string;
    params?: string;
}

export interface SkinCSSRule {
    id: string;
    file: string;
    selector: string;
    layer: string;
    body: string;
    line: number;
    order: number;
    parentId: string;
    children?: SkinCSSRule[];
    parents?: CSSAtRule[];
}

/**
 * Editable working copy of a CSS rule.
 *
 * The editor changes this object, while CSSRule remains the
 * parsed/source representation.
 */
export interface CSSRuleEditor {
    id: string;

    file: string;
    selector: string;
    layer: string;
    body: string;

    parentId: string;
    parents: CSSAtRule[];

    originalFile: string;
    originalId: string;
}
