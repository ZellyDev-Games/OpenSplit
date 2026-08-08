export interface SkinElement {
    id: string;

    label: string;

    selector: string;

    description?: string;

    file?: string;

    line?: number;

    layer?: string;

    order?: number;
}

export interface RuntimeElement {
    id: string;

    element: SkinElement;

    node: Element | null;
}
