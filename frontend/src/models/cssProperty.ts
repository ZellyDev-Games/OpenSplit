export type CSSPropertyType =
    | "text"
    | "color"
    | "length"
    | "number"
    | "percentage"
    | "font"
    | "image"
    | "gradient"
    | "variable"
    | "enum"
    | "boolean";

/**
 * Complete editor rule supplied by the backend.
 *
 * The frontend no longer determines whether a rule is being created,
 * moved or deleted. It simply edits this object and sends the complete
 * working copy back during SKIN_SAVE.
 */
export interface CSSRuleEditor {
    id: string;

    file: string;

    selector: string;

    layer: string;

    body: string;

    parents?: {
        type: "import" | "media" | "supports" | "font-face" | "layer" | "unknown";

        name: string;

        params?: string;
    }[];

    originalFile: string;

    originalId: string;
}

export interface CSSPropertyOption {
    label: string;

    value: string;
}

export interface CSSProperty {
    name: string;

    value: string;

    type: CSSPropertyType;

    inherited: boolean;

    isVariable: boolean;

    variableName?: string;

    numericValue?: number;

    unit?: string;

    options?: CSSPropertyOption[];

    removable?: boolean;
}
