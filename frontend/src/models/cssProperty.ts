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
